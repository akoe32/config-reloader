## Build

.PHONY: build
build: manifests generate fmt vet ## Build manager binary.
	go build -o bin/manager cmd/main.go

.PHONY: docker-build
docker-build: ## Build docker image with the manager.
	docker build -t ${IMG} .

.PHONY: docker-push
docker-push: ## Push docker image with the manager.
	docker push ${IMG}

## Create Cluster
.PHONY: create-cluster
create-cluster:
	kind create cluster --config infra/kind-manifest/cluster.yaml
	kind export kubeconfig --name kind --kubeconfig ~/.kube/kind 
	export KUBECONFIG=~/.kube/kind

.PHONY: load-image
load-image:
	kind load docker-image ${IMG} --name ${CLUSTER_NAME}

## Deployment

.PHONY: install-crd
install-crd: 
	kustomize build config/crd | kubectl apply -f -

.PHONY: deploy-operator
deploy-operator:
	mkdir -p dist
	cd config/manager && kustomize edit set image controller=${IMG}
	kustomize build config/default > dist/install.yaml
	kubectl apply -f dist/install.yaml

.PHONY: deploy-apps
deploy-apps:
	kustomize build infra/apps-manifest | kubectl apply -f -

.PHONY: gen-kubeconfig
gen-kubeconfig:
	kubectl cluster-info --context kind-kind	

.PHONY: undeploy
undeploy: 
	kustomize build config/default | kubectl delete -f -


.PHONY: undeploy-apps
undeploy-apps:
	kustomize build infra/apps-manifest | kubectl delete -f -

