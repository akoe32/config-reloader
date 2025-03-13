# config-reloader

## Demo
<script src="https://asciinema.org/a/hlKdfk0emGpXKyxvXfyvY52PO" id="asciicast-hlKdfk0emGpXKyxvXfyvY52PO" async></script>

## Getting Started

### Prerequisites
- kustomize
- kind
- go v.1.23
- kubebuilder

### Create Cluster with kind

```sh
make create-cluster
```
### build docker image

```sh
make docker-build IMG={IMAGE_NAME}
```
### load docker from local to kind cluster

```sh
make load-image IMG={IMAGE_NAME} CLUSTER_NAME=kind
```

### Deploy operator to cluster

```sh
make deploy-operator IMG={IMAGE_NAME}
```

### Deploy dummy apps for trial

```sh
make deploy-operator IMG={IMAGE_NAME}
```