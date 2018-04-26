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
DOCKER_REGISTRY ?= vmware
PUSH_IMAGES ?= 1
BUILD ?= $(shell date +%s)
CLUSTER_NAME_SUFFIX ?= $(shell whoami)

DOCKER_REGISTRY_HOST = $(DOCKER_REGISTRY)
export GKE_KEY
export GKE_PROJECT_ID
export DOCKER_REGISTRY
export DOCKER_REGISTRY_HOST
export PUSH_IMAGES
export BUILD
export CLUSTER_NAME_SUFFIX

.PHONY: all
all: images create-cluster build-cli deploy-dispatch e2e-tests cleanup

.PHONY: login
login:
	fly -t dispatch login

.PHONY: images
images:
	if [[ "$(DOCKER_REGISTRY)" =~ ^gcr\.io.* ]]; then gcloud docker --authorize-only; fi
	$(MAKE) -f $(DISPATCH_PATH)/Makefile images
	mkdir -p ciinputs/properties
	echo "tag=dev-$(BUILD)" > ciinputs/properties/keyval.properties


.PHONY: create-cluster
create-cluster: ## Create a Kubernetes cluster on GKE
	rm -rf ciinputs/cluster
	mkdir -p ciinputs/cluster
	fly -t dispatch e -c $(MAKE_PATH)/e2e/gke-cluster-create.yml -i dispatch=$(DISPATCH_PATH) -o k8s-cluster=ciinputs/cluster
	@echo "Cluster $(shell cat ciinputs/cluster/name) created"

.PHONY: build-cli
build-cli: ## Build CLI to be used in CI
	mkdir -p ciinputs/dispatch-cli
	GOOS=linux go build -o ciinputs/dispatch-cli/dispatch github.com/vmware/dispatch/cmd/dispatch

.PHONY: deploy-dispatch-openfaas
deploy-dispatch-openfaas: ## Deploy dispatch with OpenFaaS on Kubernetes cluster
	FAAS=openfaas EVENT_TRANSPORT=rabbitmq fly -t dispatch e -c $(MAKE_PATH)/e2e/deploy-dispatch.yml -i dispatch=$(DISPATCH_PATH) -i cluster=ciinputs/cluster -i dispatch-cli=ciinputs/dispatch-cli -i properties=ciinputs/properties -o ci-keys=ciinputs/ci-keys

.PHONY: deploy-dispatch-riff
deploy-dispatch-riff: ## Deploy dispatch with Riff on Kubernetes cluster
	FAAS=riff fly -t dispatch e -c $(MAKE_PATH)/e2e/deploy-dispatch.yml -i dispatch=$(DISPATCH_PATH) -i cluster=ciinputs/cluster -i dispatch-cli=ciinputs/dispatch-cli -i properties=ciinputs/properties -o ci-keys=ciinputs/ci-keys

.PHONY: e2e-tests-openfaas
e2e-tests-openfaas: ## Run e2e tests with OpenFaaS
	FAAS=riff EVENT_TRANSPORT=rabbitmq fly -t dispatch e -c $(MAKE_PATH)/e2e/run-tests.yml -i dispatch=$(DISPATCH_PATH) -i cluster=ciinputs/cluster -i dispatch-cli=ciinputs/dispatch-cli -i ci-keys=ciinputs/ci-keys -o tests-logs=ciinputs/tests-logs

.PHONY: e2e-tests-riff
e2e-tests-riff: ## Run e2e tests with Riff
	FAAS=riff fly -t dispatch e -c $(MAKE_PATH)/e2e/run-tests.yml -i dispatch=$(DISPATCH_PATH) -i cluster=ciinputs/cluster -i dispatch-cli=ciinputs/dispatch-cli -i ci-keys=ciinputs/ci-keys -o tests-logs=ciinputs/tests-logs

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
