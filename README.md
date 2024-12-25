# etcd-k8s-extract

**etcd-k8s-extract** is a tool that processes an etcd data directory or database file used in Kubernetes, extracts Kubernetes resources, and writes them to disk in YAML format.

This tool is especially helpful when debugging Kubernetes clusters in environments where you have limited access, or when you only have an etcd backup available. It requires no running etcd instance or additional tooling—just the etcd data directory or database file and this tool.

## Output Example

When you run **etcd-k8s-extract**, it organizes the extracted resources into directories based on their types and namespaces, like this:

```plaintext
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

## Why Use etcd-k8s-extract?

- Debug Kubernetes clusters in restricted or remote environments.
- Analyze Kubernetes configurations when only an etcd backup is available.
- Operate independently of a running etcd instance or Kubernetes control plane.

---

## Installation

### Download Binary

You can download the latest release:

```sh
curl -L https://github.com/zawachte/etcd-k8s-extract/releases/download/v0.0.2/etcd-k8s-extract -o etcd-k8s-extract
```

### Install via Go

If you already have Go installed:

```sh
go install github.com/zawachte/etcd-k8s-extract@latest
```

---

## Building from Source

If you want to build the tool yourself:

```sh
make build
```

---

## Usage

### Extract from an etcd Data Directory

```sh
./bin/etcd-k8s-extract <path-to-etcd-data-directory>
```

Example:

```sh
./bin/etcd-k8s-extract testetcd/etcd
```

### Extract from an etcd Database File

```sh
./etcd-k8s-extract <path-to-etcd-database-file>
```

Example:

```sh
./etcd-k8s-extract testetcd/etcd/member/snap/db
```

### Specify a Custom Output Path

```sh
./etcd-k8s-extract <source> --output-path <custom-output-directory>
```

Example:

```sh
./bin/etcd-k8s-extract testetcd/etcd --output-path customdir
```