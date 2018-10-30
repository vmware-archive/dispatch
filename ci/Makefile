# Allows running CI jobs manually, each step at a time. It serves following purposes:
# * validation of tasks - if task succeeded invoked using this file, it's more likely it will succeed in pipeline.
#   Executing entire pipeline costs time, so validating individual tasks brings value.
# * debugging issues with PRs - when pipeline fails, it might be easier to debug by reducing the debugging iteration time.
#
#
# ## Details: ##
# "make images" will build images. This step and the "build-cli" step are the only steps that are executed locally.
# Other steps are executed in concourse worker.
#
#
# ## Example workflow: ##
#
# 1. Build images (set DOCKER_REGISTRY if you want to use other registry than vmware):
#    DOCKER_REGISTRY=gcr.io/project-id make -f ci/Makefile images
# At this point, images are pushed to the repo.
#
# 2. Create a GKE cluster:
#    make create-cluster
# This will create a GKE cluster and print its name. You can retrieve kubelet credentials using following command:
#    gcloud container clusters get-credentials CLUSTER_NAME
# You don't need to do it, as deploy dispatch command will deploy dispatch for you.
# NOTE: Once you created a cluster, you can skip this step in next debugging iteration.
#
# 4. Prepare Dispatch CLI:
#    make build-cli
# if you changed something in CLI code, run this step again before running tests or deploying dispatch
#
# 3. Deploy dispatch:
#    DOCKER_REGISTRY=gcr.io/project-id make deploy-dispatch-openfaas
# or
#    DOCKER_REGISTRY=gcr.io/project-id make deploy-dispatch-riff
# depending on which faas backend you would like to use. Once you deploy dispatch, you can update it manually using kubectl,
# or rerun this step (if you, for example, updated image tag).
#
# 4. Run e2e tests:
#    make e2e-tests-riff
# or
#    make e2e-tests-openfaas
# this step will execute e2e-tests.
#
# 5. Uninstall dispatch:
#    make uninstall-dispatch
# Run this when you're done with testing. Note: make sure you run this step so there are no leftovers after cluster is deleted.
#
# 6. Delete the cluster:
#    make cleanup
# Run this to delete the k8s cluster on GKE. MAKE SURE YOU RUN THIS STEP!
#

MAKE_PATH := $(dir $(realpath $(firstword $(MAKEFILE_LIST))))
DISPATCH_PATH := $(MAKE_PATH)/..

GKE_KEY ?= ((gke-key))
GKE_PROJECT_ID ?= ((gke-project-id))
DISPATCH_SERVER_DOCKER_REPOSITORY ?= dispatchframework
DOCKER_URL ?= https://gcr.io
DOCKER_REGISTRY ?= ((ci-registry-org.gcr))
DOCKER_USERNAME ?= ((ci-registry-username.gcr))
DOCKER_PASSWORD ?= ((ci-registry-password.gcr))
MINIO_USERNAME ?= dispatch
MINIO_PASSWORD ?= dispatch
DISPATCH_NAMESPACE ?= dispatch-server
RELEASE_NAME ?= dispatch-server
PUSH_IMAGES ?= 1
TAG ?= dev-$(shell date +%s)
CLUSTER_NAME_SUFFIX ?= $(shell whoami)

GIT_VERSION = $(shell git describe --tags --dirty)
VERSION ?= $(GIT_VERSION)

VERSION_PACKAGE := github.com/vmware/dispatch/pkg/version
CLICMD_PACKAGE := github.com/vmware/dispatch/pkg/dispatchcli/cmd

GO_LDFLAGS := -X $(VERSION_PACKAGE).version=$(VERSION)
GO_LDFLAGS += -X $(VERSION_PACKAGE).buildDate=$(shell date +'%Y-%m-%dT%H:%M:%SZ')
GO_LDFLAGS += -X $(VERSION_PACKAGE).commit=$(shell git rev-parse HEAD)
CLI_LDFLAGS := -X $(CLICMD_PACKAGE).imagesB64=$(shell cat $(DISPATCH_PATH)/images.yaml | gzip | base64)

DOCKER_REGISTRY_HOST = $(DOCKER_REGISTRY)
DOCKER_REPOSITORY = $(DOCKER_REGISTRY)
export GKE_KEY
export GKE_PROJECT_ID
export DISPATCH_SERVER_DOCKER_REPOSITORY
export DOCKER_URL
export DOCKER_REGISTRY
export DOCKER_REGISTRY_HOST
export DOCKER_REPOSITORY
export DOCKER_USERNAME
export DOCKER_PASSWORD
export MINIO_USERNAME
export MINIO_PASSWORD
export RELEASE_NAME
export DISPATCH_NAMESPACE
export PUSH_IMAGES
export TAG
export CLUSTER_NAME_SUFFIX

.PHONY: all
all: images create-cluster install-knative build-cli deploy-dispatch-server e2e-tests cleanup

.PHONY: login
login:
	fly -t dispatch login

.PHONY: images
images:
	PREFIX=.. $(MAKE) -f ../Makefile images
	mkdir -p ciinputs/properties
	echo "image.tag=$(TAG)" > ciinputs/properties/keyval.properties
	echo "image.host=$(DISPATCH_SERVER_DOCKER_REPOSITORY)" >> ciinputs/properties/keyval.properties

.PHONY: create-cluster
create-cluster: ## Create a Kubernetes cluster on GKE
	rm -rf ciinputs/cluster
	mkdir -p ciinputs/cluster
	fly -t dispatch e -c $(MAKE_PATH)/e2e/gke-cluster-create.yml -i dispatch=$(DISPATCH_PATH) -o k8s-cluster=ciinputs/cluster
	@echo "Cluster $(shell cat ciinputs/cluster/name) created"

.PHONY: install-knative
install-knative: ## Install knative on a GKE cluster
	fly -t dispatch e -c $(MAKE_PATH)/e2e/gke-knative-install.yml -i dispatch=$(DISPATCH_PATH) -i cluster=ciinputs/cluster
	@echo "Knative installed on $(shell cat ciinputs/cluster/name)"

.PHONY: build-cli
build-cli: ## Build CLI to be used in CI
	mkdir -p ciinputs/dispatch-cli
	GOOS=linux go build -ldflags "$(GO_LDFLAGS) $(CLI_LDFLAGS)" -o ciinputs/dispatch-cli/dispatch github.com/vmware/dispatch/cmd/dispatch

.PHONY: deploy-dispatch-server
deploy-dispatch-server: ## Deploy dispatch on cluster with knative installed
	fly -t dispatch e -c $(MAKE_PATH)/e2e/deploy-dispatch.yml -i dispatch=$(DISPATCH_PATH) -i cluster=ciinputs/cluster -i dispatch-cli=ciinputs/dispatch-cli -i properties=ciinputs/properties -o ci-keys=ciinputs/ci-keys

.PHONY: e2e-tests
e2e-tests: ## Run e2e tests
	fly -t dispatch e -c $(MAKE_PATH)/e2e/run-tests.yml -i dispatch=$(DISPATCH_PATH) -i cluster=ciinputs/cluster -i dispatch-cli=ciinputs/dispatch-cli -i ci-keys=ciinputs/ci-keys -o tests-logs=ciinputs/tests-logs

.PHONY: collect-logs
collect-logs: ## Collect logs && uninstall
	fly -t dispatch e -c $(MAKE_PATH)/e2e/collect-logs.yml -i dispatch=$(DISPATCH_PATH) -i cluster=ciinputs/cluster -i tests-logs=ciinputs/tests-logs -o dispatch-logs=ciinputs/dispatch-logs

.PHONY: uninstall-dispatch
uninstall-dispatch: ## Uninstall Dispatch from the cluster.
	fly -t dispatch e -c $(MAKE_PATH)/e2e/uninstall-dispatch.yml -i dispatch=$(DISPATCH_PATH) -i cluster=ciinputs/cluster -i dispatch-cli=ciinputs/dispatch-cli

.PHONY: cleanup
cleanup: ## Cleanup the cluster
	fly -t dispatch e -c $(MAKE_PATH)/e2e/cleanup.yml -i dispatch=$(DISPATCH_PATH) -i cluster=ciinputs/cluster

help: ## Display make help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(firstword $(MAKEFILE_LIST)) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
