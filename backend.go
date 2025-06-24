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
	appsv1beta2 "k8s.io/api/apps/v1beta2"

	autoscalingv2 "k8s.io/api/autoscaling/v2"
	batchv1 "k8s.io/api/batch/v1"
	coordinationv1 "k8s.io/api/coordination/v1"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	flowcontrolv1 "k8s.io/api/flowcontrol/v1"
	networkingv1 "k8s.io/api/networking/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	rbacv1beta1 "k8s.io/api/rbac/v1beta1"
	schedulingv1 "k8s.io/api/scheduling/v1"
	storagev1 "k8s.io/api/storage/v1"

	"k8s.io/apimachinery/pkg/runtime"

	"go.etcd.io/etcd/api/v3/mvccpb"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/protobuf"
	"k8s.io/kubectl/pkg/scheme"

	bolt "go.etcd.io/bbolt"
)

func WriteResourceToFile(kuberegistryPath string, data []byte, ret mvccpb.KeyValue) error {
	keyPath := strings.ReplaceAll(string(ret.Key), "/registry/", "")
	subDirectory := filepath.Join(kuberegistryPath, keyPath)
	err := os.MkdirAll(subDirectory, 0755)
	if err != nil {
		return fmt.Errorf("failed make directory: %w", err)
	}

	versionFileName := fmt.Sprintf("%d-%d.yaml", ret.Version, ret.ModRevision)
	err = os.WriteFile(filepath.Join(subDirectory, versionFileName), data, 0644)
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

	intoObject, err := kindToObject(obj.GroupVersionKind())
	if err != nil {
		return fmt.Errorf("failed to get runtime object from gvk %s: %w\n", obj.GroupVersionKind().String(), err)
	}

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
func kindToObjectCore(gvk schema.GroupVersionKind) (runtime.Object, error) {
	switch gvk.Kind {
	case "Binding":
		if gvk.Version == "v1" {
			return &corev1.Binding{}, nil
		}
	case "ComponentStatus":
		if gvk.Version == "v1" {
			return &corev1.ComponentStatus{}, nil
		}
	case "ConfigMap":
		if gvk.Version == "v1" {
			return &corev1.ConfigMap{}, nil
		}
	case "Endpoints":
		if gvk.Version == "v1" {
			return &corev1.Endpoints{}, nil
		}
	case "Event":
		if gvk.Version == "v1" {
			return &corev1.Event{}, nil
		}
	case "LimitRange":
		if gvk.Version == "v1" {
			return &corev1.LimitRange{}, nil
		}
	case "Namespace":
		if gvk.Version == "v1" {
			return &corev1.Namespace{}, nil
		}
	case "Node":
		if gvk.Version == "v1" {
			return &corev1.Node{}, nil
		}
	case "PersistentVolume":
		if gvk.Version == "v1" {
			return &corev1.PersistentVolume{}, nil
		}
	case "PersistentVolumeClaim":
		if gvk.Version == "v1" {
			return &corev1.PersistentVolumeClaim{}, nil
		}
	case "Pod":
		if gvk.Version == "v1" {
			return &corev1.Pod{}, nil
		}
	case "PodTemplate":
		if gvk.Version == "v1" {
			return &corev1.PodTemplate{}, nil
		}
	case "ReplicationController":
		if gvk.Version == "v1" {
			return &corev1.ReplicationController{}, nil
		}
	case "ResourceQuota":
		if gvk.Version == "v1" {
			return &corev1.ResourceQuota{}, nil
		}
	case "Secret":
		if gvk.Version == "v1" {
			return &corev1.Secret{}, nil
		}
	case "Service":
		if gvk.Version == "v1" {
			return &corev1.Service{}, nil
		}
	case "ServiceAccount":
		if gvk.Version == "v1" {
			return &corev1.ServiceAccount{}, nil
		}
	case "RangeAllocation":
		if gvk.Version == "v1" {
			return &corev1.RangeAllocation{}, nil
		}
	}
	return nil, fmt.Errorf("unsupported version %s for kind %s", gvk.Version, gvk.Kind)
}

func kindToObjectApps(gvk schema.GroupVersionKind) (runtime.Object, error) {
	switch gvk.Kind {
	case "DaemonSet":
		if gvk.Version == "v1" {
			return &appsv1.DaemonSet{}, nil
		} else if gvk.Version == "v1beta2" {
			return &appsv1beta2.DaemonSet{}, nil
		}
	case "Deployment":
		if gvk.Version == "v1" {
			return &appsv1.Deployment{}, nil
		} else if gvk.Version == "v1beta2" {
			return &appsv1beta2.Deployment{}, nil
		}
	case "ReplicaSet":
		if gvk.Version == "v1" {
			return &appsv1.ReplicaSet{}, nil
		} else if gvk.Version == "v1beta2" {
			return &appsv1beta2.ReplicaSet{}, nil
		}
	case "StatefulSet":
		if gvk.Version == "v1" {
			return &appsv1.StatefulSet{}, nil
		} else if gvk.Version == "v1beta2" {
			return &appsv1beta2.StatefulSet{}, nil
		}
	case "ControllerRevision":
		if gvk.Version == "v1" {
			return &appsv1.ControllerRevision{}, nil
		} else if gvk.Version == "v1beta2" {
			return &appsv1beta2.ControllerRevision{}, nil
		}
	}
	return nil, fmt.Errorf("unsupported version %s for kind %s", gvk.Version, gvk.Kind)
}

func kindToObjectCoordination(gvk schema.GroupVersionKind) (runtime.Object, error) {
	if gvk.Kind == "Lease" && gvk.Version == "v1" {
		return &coordinationv1.Lease{}, nil
	}
	return nil, fmt.Errorf("unsupported version %s for kind %s", gvk.Version, gvk.Kind)
}

func kindToObjectDiscovery(gvk schema.GroupVersionKind) (runtime.Object, error) {
	if gvk.Kind == "EndpointSlice" && gvk.Version == "v1" {
		return &discoveryv1.EndpointSlice{}, nil
	}
	return nil, fmt.Errorf("unsupported version %s for kind %s", gvk.Version, gvk.Kind)
}

func kindToObjectAdmissionRegistration(gvk schema.GroupVersionKind) (runtime.Object, error) {
	switch gvk.Kind {
	case "MutatingWebhookConfiguration":
		if gvk.Version == "v1" {
			return &admissionregistrationv1.MutatingWebhookConfiguration{}, nil
		}
	case "ValidatingWebhookConfiguration":
		if gvk.Version == "v1" {
			return &admissionregistrationv1.ValidatingWebhookConfiguration{}, nil
		}
	}
	return nil, fmt.Errorf("unsupported version %s for kind %s", gvk.Version, gvk.Kind)
}

func kindToObjectRbac(gvk schema.GroupVersionKind) (runtime.Object, error) {
	switch gvk.Kind {
	case "ClusterRole":
		if gvk.Version == "v1" {
			return &rbacv1.ClusterRole{}, nil
		} else if gvk.Version == "v1beta1" {
			return &rbacv1beta1.ClusterRole{}, nil
		}
	case "ClusterRoleBinding":
		if gvk.Version == "v1" {
			return &rbacv1.ClusterRoleBinding{}, nil
		} else if gvk.Version == "v1beta1" {
			return &rbacv1beta1.ClusterRoleBinding{}, nil
		}
	case "Role":
		if gvk.Version == "v1" {
			return &rbacv1.Role{}, nil
		} else if gvk.Version == "v1beta1" {
			return &rbacv1beta1.Role{}, nil
		}
	case "RoleBinding":
		if gvk.Version == "v1" {
			return &rbacv1.RoleBinding{}, nil
		} else if gvk.Version == "v1beta1" {
			return &rbacv1beta1.RoleBinding{}, nil
		}
	}
	return nil, fmt.Errorf("unsupported version %s for kind %s", gvk.Version, gvk.Kind)
}

func kindToObjectScheduling(gvk schema.GroupVersionKind) (runtime.Object, error) {
	if gvk.Kind == "PriorityClass" && gvk.Version == "v1" {
		return &schedulingv1.PriorityClass{}, nil
	}
	return nil, fmt.Errorf("unsupported version %s for kind %s", gvk.Version, gvk.Kind)
}

func kindToObjectFlowControl(gvk schema.GroupVersionKind) (runtime.Object, error) {
	switch gvk.Kind {
	case "FlowSchema":
		if gvk.Version == "v1" {
			return &flowcontrolv1.FlowSchema{}, nil
		}
	case "PriorityLevelConfiguration":
		if gvk.Version == "v1" {
			return &flowcontrolv1.PriorityLevelConfiguration{}, nil
		}
	}
	return nil, fmt.Errorf("unsupported version %s for kind %s", gvk.Version, gvk.Kind)
}

func kindToObjectStorage(gvk schema.GroupVersionKind) (runtime.Object, error) {
	switch gvk.Kind {
	case "StorageClass":
		if gvk.Version == "v1" {
			return &storagev1.StorageClass{}, nil
		}
	case "VolumeAttachment":
		if gvk.Version == "v1" {
			return &storagev1.VolumeAttachment{}, nil
		}
	case "CSINode":
		if gvk.Version == "v1" {
			return &storagev1.CSINode{}, nil
		}
	}
	return nil, fmt.Errorf("unsupported version %s for kind %s", gvk.Version, gvk.Kind)
}

func kindToObjectNetworking(gvk schema.GroupVersionKind) (runtime.Object, error) {
	switch gvk.Kind {
	case "Ingress":
		if gvk.Version == "v1" {
			return &networkingv1.Ingress{}, nil
		}
	case "NetworkPolicy":
		if gvk.Version == "v1" {
			return &networkingv1.NetworkPolicy{}, nil
		}
	}
	return nil, fmt.Errorf("unsupported version %s for kind %s", gvk.Version, gvk.Kind)
}

func kindToObjectBatch(gvk schema.GroupVersionKind) (runtime.Object, error) {
	switch gvk.Kind {
	case "CronJob":
		if gvk.Version == "v1" {
			return &batchv1.CronJob{}, nil
		}
	}
	return nil, fmt.Errorf("unsupported version %s for kind %s", gvk.Version, gvk.Kind)
}

func kindToObjectAutoscaling(gvk schema.GroupVersionKind) (runtime.Object, error) {
	switch gvk.Kind {
	case "HorizontalPodAutoscaler":
		if gvk.Version == "v2" {
			return &autoscalingv2.HorizontalPodAutoscaler{}, nil
		}
	}
	return nil, fmt.Errorf("unsupported version %s for kind %s", gvk.Version, gvk.Kind)
}

func kindToObject(gvk schema.GroupVersionKind) (runtime.Object, error) {
	switch gvk.Group {
	case corev1.GroupName:
		return kindToObjectCore(gvk)
	case appsv1.GroupName:
		return kindToObjectApps(gvk)
	case coordinationv1.GroupName:
		return kindToObjectCoordination(gvk)
	case discoveryv1.GroupName:
		return kindToObjectDiscovery(gvk)
	case admissionregistrationv1.GroupName:
		return kindToObjectAdmissionRegistration(gvk)
	case rbacv1.GroupName:
		return kindToObjectRbac(gvk)
	case schedulingv1.GroupName:
		return kindToObjectScheduling(gvk)
	case flowcontrolv1.GroupName:
		return kindToObjectFlowControl(gvk)
	case storagev1.GroupName:
		return kindToObjectStorage(gvk)
	case networkingv1.GroupName:
		return kindToObjectNetworking(gvk)
	case batchv1.GroupName:
		return kindToObjectBatch(gvk)
	case autoscalingv2.GroupName:
		return kindToObjectAutoscaling(gvk)
	default:
		return nil, fmt.Errorf("unsupported group %s", gvk.Group)
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
