package main

import (
	"encoding/binary"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"sigs.k8s.io/yaml"

	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	appsv1 "k8s.io/api/apps/v1"
	coordinationv1 "k8s.io/api/coordination/v1"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	flowcontrolv1 "k8s.io/api/flowcontrol/v1beta1"
	rbacv1 "k8s.io/api/rbac/v1"
	schedulingv1 "k8s.io/api/scheduling/v1"
	storagev1 "k8s.io/api/storage/v1"

	"k8s.io/apimachinery/pkg/runtime"

	"go.etcd.io/etcd/api/v3/mvccpb"
	"k8s.io/apimachinery/pkg/runtime/serializer/protobuf"
	"k8s.io/kubectl/pkg/scheme"

	bolt "go.etcd.io/bbolt"
)

func WriteResourceToFile(kuberegistryPath string, data []byte, ret mvccpb.KeyValue) error {
	keyPath := strings.ReplaceAll(string(ret.Key), "/registry/", "")
	subDirectory := path.Join(kuberegistryPath, keyPath)
	err := os.MkdirAll(subDirectory, 0755)
	if err != nil {
		return fmt.Errorf("failed make directory: %w", err)
	}

	versionFileName := fmt.Sprintf("%d-%d.yaml", ret.Version, ret.ModRevision)
	err = os.WriteFile(path.Join(subDirectory, versionFileName), data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

func snapDir(dataDir string) string {
	return filepath.Join(dataDir, "member", "snap")
}

type revision struct {
	main int64
	sub  int64
}

func bytesToRev(bytes []byte) revision {
	return revision{
		main: int64(binary.BigEndian.Uint64(bytes[0:8])),
		sub:  int64(binary.BigEndian.Uint64(bytes[9:])),
	}
}

func decodeAndWrite(kuberegistryPath string, k, v []byte) error {
	var kv mvccpb.KeyValue
	if err := kv.Unmarshal(v); err != nil {
		return err
	}

	fmt.Printf("writing key %q | created_key %d | current_key %d | version %d\n", string(kv.Key), kv.CreateRevision, kv.ModRevision, kv.Version)

	codec := protobuf.NewSerializer(scheme.Scheme, scheme.Scheme)
	obj := &runtime.Unknown{}

	_, _, err := codec.Decode(kv.Value, nil, obj)
	if err != nil {
		yamlData, errJson := yaml.JSONToYAML(kv.Value)
		if errJson != nil {
			return fmt.Errorf("failed to decode protobuf data: %w and json data %w\n", err, errJson)
		}

		return WriteResourceToFile(kuberegistryPath, yamlData, kv)
	}

	intoObject := kindToObject(obj.Kind)
	_, _, err = codec.Decode(kv.Value, nil, intoObject)
	if err != nil {
		return fmt.Errorf("failed to decode protobuf data into runtime object: %w\n", err)
	}
	yamlData, err := yaml.Marshal(intoObject)
	if err != nil {
		return fmt.Errorf("failed to marshal runtime object into yaml: %w\n", err)
	}

	err = WriteResourceToFile(kuberegistryPath, yamlData, kv)
	if err != nil {
		return fmt.Errorf("failed to write protobuf resource to file: %w\n", err)
	}

	return nil
}

func kindToObject(kind string) runtime.Object {
	switch kind {
	case "Binding":
		return &corev1.Binding{}
	case "ComponentStatus":
		return &corev1.ComponentStatus{}
	case "ConfigMap":
		return &corev1.ConfigMap{}
	case "Endpoints":
		return &corev1.Endpoints{}
	case "Event":
		return &corev1.Event{}
	case "LimitRange":
		return &corev1.LimitRange{}
	case "Namespace":
		return &corev1.Namespace{}
	case "Node":
		return &corev1.Node{}
	case "PersistentVolume":
		return &corev1.PersistentVolume{}
	case "PersistentVolumeClaim":
		return &corev1.PersistentVolumeClaim{}
	case "Pod":
		return &corev1.Pod{}
	case "PodTemplate":
		return &corev1.PodTemplate{}
	case "ReplicationController":
		return &corev1.ReplicationController{}
	case "ResourceQuota":
		return &corev1.ResourceQuota{}
	case "Secret":
		return &corev1.Secret{}
	case "Service":
		return &corev1.Service{}
	case "ServiceAccount":
		return &corev1.ServiceAccount{}
	case "Lease":
		return &coordinationv1.Lease{}
	case "DaemonSet":
		return &appsv1.DaemonSet{}
	case "Deployment":
		return &appsv1.Deployment{}
	case "ReplicaSet":
		return &appsv1.ReplicaSet{}
	case "StatefulSet":
		return &appsv1.StatefulSet{}
	case "EndpointSlice":
		return &discoveryv1.EndpointSlice{}
	case "MutatingWebhookConfiguration":
		return &admissionregistrationv1.MutatingWebhookConfiguration{}
	case "ValidatingWebhookConfiguration":
		return &admissionregistrationv1.ValidatingWebhookConfiguration{}
	case "RangeAllocation":
		return &corev1.RangeAllocation{}
	case "ClusterRole":
		return &rbacv1.ClusterRole{}
	case "ClusterRoleBinding":
		return &rbacv1.ClusterRoleBinding{}
	case "Role":
		return &rbacv1.Role{}
	case "RoleBinding":
		return &rbacv1.RoleBinding{}
	case "ControllerRevision":
		return &appsv1.ControllerRevision{}
	case "StorageClass":
		return &storagev1.StorageClass{}
	case "VolumeAttachment":
		return &storagev1.VolumeAttachment{}
	case "CSINode":
		return &storagev1.CSINode{}
	case "PriorityClass":
		return &schedulingv1.PriorityClass{}
	case "FlowSchema":
		return &flowcontrolv1.FlowSchema{}
	case "PriorityLevelConfiguration":
		return &flowcontrolv1.PriorityLevelConfiguration{}
	default:
		return nil
	}
}

func iterateEtcdKeys(dbPath, outputPath string, limit uint64) error {
	bucket := "key"
	kuberegistryPath := path.Join(outputPath, "kuberegistry")
	_, err := os.Stat(kuberegistryPath)
	if err == nil {
		return fmt.Errorf("kuberegistry directory already exists, please delete it first")
	}

	err = os.MkdirAll(kuberegistryPath, 0755)
	if err != nil {
		return fmt.Errorf("failed make kuberegistry directory: %w", err)
	}

	db, err := bolt.Open(dbPath, 0600, &bolt.Options{Timeout: flockTimeout})
	if err != nil {
		return fmt.Errorf("failed to open bolt DB %v", err)
	}
	defer db.Close()

	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return fmt.Errorf("got nil bucket for %s", bucket)
		}

		c := b.Cursor()
		for k, v := c.Last(); k != nil; k, v = c.Prev() {
			err := decodeAndWrite(kuberegistryPath, k, v)
			if err != nil {
				fmt.Printf("failed to decode key %v\n", err)
			}

			limit--
			if limit == 0 {
				break
			}
		}

		return nil
	})
	return err
}
