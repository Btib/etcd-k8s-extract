# etcd-k8s-extract

etcd-k8s-extract takes in an etcd data directory or db file used in kubernetes, extracts the kubernetes resources and then writes the resources to disk in yaml format.

Running this tool will result in something like this:
```sh
├── pods
│   ├── capd-system
│   │   └── capd-controller-manager-7dfb5f78df-bcz44
│   │       └── 5-2301.yaml
│   ├── capi-kubeadm-bootstrap-system
│   │   └── capi-kubeadm-bootstrap-controller-manager-6845c9fc68-q7z2c
│   │       └── 5-2254.yaml
│   ├── capi-kubeadm-control-plane-system
│   │   └── capi-kubeadm-control-plane-controller-manager-c9c8bccc6-97fck
│   │       └── 5-2277.yaml
│   ├── capi-system
│   │   └── capi-controller-manager-66dbddf59f-67x5g
│   │       └── 5-2233.yaml
```

This tool is really useful if you have kubernetes running in a remote area where you can't debug easily or only have a etcd backup for a cluster. No need for a running etcd or any extra tooling, just the etcd data-directory/db-file and this tool and you are good to go. 

## Installing

If you have go already setup:

```sh
go install github.com/zawachte/etcd-k8s-extract@latest
```

## Building 

```sh
make build
```

## Running

### With an etcd data directory:
```sh
./bin/etcd-k8s-extract testetcd/etcd
```

### With an etcd database file

```sh
./bin/etcd-k8s-extract testetcd/etcd/member/snap/db
```

### With a custom output path

```sh
./bin/etcd-k8s-extract testetcd/etcd --output-path customdir
```