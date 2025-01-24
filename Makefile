
TAG = 0.1
# Image URL to use for building/pushing
IMG ?= alexdecb/hello-world:$(TAG)

# ENVTEST_K8S_VERSION refers to the version of kubebuilder assets to be downloaded by envtest binary.
ENVTEST_K8S_VERSION = 1.29.0

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# CONTAINER_TOOL defines the container tool to be used for building images.
# Be aware that the target commands are only tested with Docker which is
# scaffolded by default. However, you might want to replace it to use other
# tools. (i.e. podman)
CONTAINER_TOOL ?= docker

# Setting SHELL to bash allows bash commands to be executed by recipes.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

REPOSITORY=l2sm-md
## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

## Tool Binaries
KUBECTL ?= kubectl
KUSTOMIZE ?= $(LOCALBIN)/kustomize-$(KUSTOMIZE_VERSION)
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen-$(CONTROLLER_TOOLS_VERSION)
ENVTEST ?= $(LOCALBIN)/setup-envtest-$(ENVTEST_VERSION)
GOLANGCI_LINT = $(LOCALBIN)/golangci-lint-$(GOLANGCI_LINT_VERSION)
KIND ?= kind
DOCKER ?= docker 

WORKER_CLUSTER_NUM ?= 2
## Tool Versions
KUSTOMIZE_VERSION ?= v5.3.0
CONTROLLER_TOOLS_VERSION ?= v0.14.0
ENVTEST_VERSION ?= latest
GOLANGCI_LINT_VERSION ?= v1.54.2
##@ Build and Push



.PHONY: docker-build
docker-build: ## Build docker image with the server.
	$(CONTAINER_TOOL) build -t ${IMG} .

.PHONY: docker-push
docker-push: ## Push docker image to the repository.
	$(CONTAINER_TOOL) push ${IMG}

##@ Help

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

.PHONY: generate-proto
export PATH := $(PATH):$(LOCALBIN)
generate-proto: install-tools ## Generate gRPC code from .proto file.
	protoc -I=api/v1 --go_out=paths=source_relative:./api/v1/hello --go-grpc_out=paths=source_relative:./api/v1/hello api/v1/hello.proto

.PHONY: run
include .env
export $(shell sed 's/=.*//' .env)
run: 
	go run ./cmd/hello

.PHONY: build
build: fmt vet 
	go build -o $(LOCALBIN)/server ./cmd/hello/


.PHONY: build-installer
build-installer: kustomize ## Generate a consolidated YAML with CRDs and deployment.
	echo "" > deployments/deployment.yaml
	echo "---" >> deployments/deployment.yaml  
	cd config/hello && $(KUSTOMIZE) edit set image hello=${IMG}
	$(KUSTOMIZE) build config/default >> deployments/deployment.yaml

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...


.PHONY: deploy
deploy: kustomize ## Deploy server to the K8s cluster specified in ~/.kube/config.
	cd config/hello && $(KUSTOMIZE) edit set image hello=${IMG}
	$(KUSTOMIZE) build config/default | $(KUBECTL) apply -f - 

.PHONY: undeploy
undeploy: kustomize ## Undeploy server from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build config/default | $(KUBECTL) delete --ignore-not-found=true -f -

.PHONY: kustomize
kustomize: $(KUSTOMIZE) ## Download kustomize locally if necessary.
$(KUSTOMIZE): $(LOCALBIN)
	$(call go-install-tool,$(KUSTOMIZE),sigs.k8s.io/kustomize/kustomize/v5,$(KUSTOMIZE_VERSION))
	
.PHONY: deploy-dev
deploy-dev: apply-cert kustomize
	$(KUSTOMIZE) build config/dev | $(KUBECTL) apply -f - 
	
.PHONY: undeploy-dev
undeploy-dev: kustomize ## Undeploy server from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build config/dev | $(KUBECTL) delete --ignore-not-found=true -f -

# Define file extensions for various formats
FILES := $(shell find . -type f \( -name "*.go" -o -name "*.json" -o -name "*.yaml" -o -name "*.yml" -o -name "*.md" \))

# Install the addlicense tool if not installed
.PHONY: install-tools
install-tools:
	GOBIN=$(LOCALBIN) go install github.com/google/addlicense@latest
	GOBIN=$(LOCALBIN) go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	GOBIN=$(LOCALBIN) go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest


.PHONY: add-license
add-license: install-tools
	@for file in $(FILES); do \
		$(LOCALBIN)/addlicense -f ./hack/LICENSE.txt -l apache "$${file}"; \
	done



.PHONY: clean
clean:
	
