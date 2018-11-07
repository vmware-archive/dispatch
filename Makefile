
GIT_VERSION = $(shell git describe --tags --dirty)
VERSION ?= $(GIT_VERSION)

GO ?= go
GOVERSIONS ?= go1.9 go1.10 go1.11
OS := $(shell uname)
SHELL := /bin/bash

.DEFAULT_GOAL := help

GOPATH := $(firstword $(subst :, ,$(GOPATH)))
SWAGGER := $(GOPATH)/bin/swagger
GOBINDATA := $(GOPATH)/bin/go-bindata

VERSION_PACKAGE := github.com/vmware/dispatch/pkg/version
CLICMD_PACKAGE := github.com/vmware/dispatch/pkg/dispatchcli/cmd

GO_LDFLAGS := -X $(VERSION_PACKAGE).version=$(VERSION)
GO_LDFLAGS += -X $(VERSION_PACKAGE).buildDate=$(shell date +'%Y-%m-%dT%H:%M:%SZ')
GO_LDFLAGS += -X $(VERSION_PACKAGE).commit=$(shell git rev-parse HEAD)

B64_ARGS :=
ifeq ($(OS),Linux)
	B64_ARGS += -w0
endif

CLI_LDFLAGS := -X $(CLICMD_PACKAGE).imagesB64=$(shell cat images.yaml | gzip | base64 $(B64_ARGS))

PKGS := pkg


# ?= cannot be used for these variables as they should be evaulated only once per Makefile
ifeq ($(PREFIX),)
PREFIX := $(shell pwd)
endif

TAG ?= $(VERSION)



.PHONY: all
all: generate linux darwin ova

.PHONY: goversion
goversion:
	@echo Checking go version...
	@(goversion=$$($(GO) version | cut -d' ' -f3 | cut -d'.' -f1,2); echo "$(GOVERSIONS)" | grep -q $$goversion || ( echo "Please install one of $(GOVERSIONS) (found: $$goversion)" && exit 2 ))

.PHONY: check
check: goversion checkfmt checklint swagger-validate ## check if the source files comply to the formatting rules
	@echo running metalint ...
	# (If errors involves swagger-generated files) consider running "make generate" and retry.)
	gometalinter --disable=gotype --vendor --deadline 30s --fast --errors ./...
	@echo running header check ...
	scripts/header-check.sh

.PHONY: check-all
check-all: goversion checkfmt checklint swagger-validate ## check if the source files comply to the formatting rules
	@echo running metalint ...
	# (If errors involves swagger-generated files) consider running "make generate" and retry.)
	gometalinter --disable=gotype --exclude gen --vendor --deadline 30s --aggregate --fast ./...
	@echo running header check ...
	scripts/header-check.sh

.PHONY: fmt
fmt: ## format go source code
	gofmt -w $$(find . -name '*.go' -not -path './vendor/*' -not -path './gen/*')

.PHONY: difffmt
difffmt: ## diplay formatting changes that would be made by fix
	gofmt -d $$(find . -name '*.go' -not -path './vendor/*' -not -path './gen/*')

.PHONY: fix-headers
fix-headers: ## fix copyright headers if they are missing
	scripts/header-check.sh fix

.PHONY: checkfmt
checkfmt: ## check formatting of source files
	scripts/gofmtcheck.sh

.PHONY: checklint
checklint: ## check lint of source files
	scripts/golintcheck.sh

.PHONY: test
test: ## run tests
	@echo running tests...
	$(GO) test -v $(shell go list -v ./... | grep -v /vendor/ | grep -v integration )

.PHONY: swagger-validate
swagger-validate: ## validate the swagger spec
	swagger validate ./swagger/*.yaml

.PHONY: run-dev
run-dev: ## run the dev server
	@./scripts/run-dev.sh

CLI = dispatch

SERVICES = dispatch-server

DARWIN_BINS = $(foreach bin,$(SERVICES),$(bin)-darwin)
LINUX_BINS = $(foreach bin,$(SERVICES),$(bin)-linux)

.PHONY: darwin linux $(LINUX_BINS) $(DARWIN_BINS)
linux: $(LINUX_BINS) cli-linux
darwin: $(DARWIN_BINS) cli-darwin

$(LINUX_BINS):
	GOOS=linux go build -ldflags "$(GO_LDFLAGS)" -o bin/$@ ./cmd/$(subst -linux,,$@)

$(DARWIN_BINS):
	GOOS=darwin go build -ldflags "$(GO_LDFLAGS)" -o bin/$@ ./cmd/$(subst -darwin,,$@)

cli-darwin:
	GOOS=darwin go build -ldflags "$(GO_LDFLAGS) $(CLI_LDFLAGS)" -o bin/$(CLI)-darwin ./cmd/$(CLI)

cli-linux:
	GOOS=linux go build -ldflags "$(GO_LDFLAGS) $(CLI_LDFLAGS)" -o bin/$(CLI)-linux ./cmd/$(CLI)

.PHONY: toolbox
toolbox:
	GOOS=linux go build -ldflags "$(GO_LDFLAGS) $(CLI_LDFLAGS)" -o bin/toolbox ./cmd/toolbox

.PHONY: rpctool
rpctool:
	GOOS=linux go build -ldflags "$(GO_LDFLAGS) $(CLI_LDFLAGS)" -o bin/rpctool ./cmd/rpctool

.PHONY: ovfenv
ovfenv:
	GOOS=linux go build -ldflags "$(GO_LDFLAGS) $(CLI_LDFLAGS)" -o bin/ovfenv ./cmd/ovfenv

.PHONY: ova-binaries
ova-binaries: linux toolbox rpctool ovfenv ## build all binaries used by the OVA

.PHONY: ova
ova: ova-binaries ## build an OVA which includes dispatch-server
	./ova/build.sh ova-dev

.PHONY: images
images: linux ci-images

.PHONY: ci-values
ci-values:
	TAG=$(TAG) scripts/values.sh

.PHONY: ci-images $(SERVICES)
ci-images: ci-values $(SERVICES)

$(SERVICES):
	TAG=$(TAG) scripts/images.sh $@

.PHONY: generate
generate: ## run go generate
	scripts/generate-models.sh swagger/models.json
	scripts/generate.sh api-manager APIManager api-manager.yaml
	scripts/generate.sh event-manager EventManager event-manager.yaml
	scripts/generate.sh function-manager FunctionManager function-manager.yaml
	scripts/generate.sh identity-manager IdentityManager identity-manager.yaml
	scripts/generate.sh image-manager ImageManager image-manager.yaml
	scripts/generate.sh secret-store SecretStore secret-store.yaml
	scripts/header-check.sh fix

.PHONY: gen-clean
gen-clean:  ## Clean all files created with make generate
	rm -rf ./pkg/*/gen

.PHONY: distclean
distclean:  ## Clean ALL files including ignored ones
	git clean -f -d -x .


.PHONY: clean
clean:  ## Clean all compiled files
	rm -f $(foreach bin,$(DARWIN_BINS),bin/$(bin))
	rm -f $(foreach bin,$(LINUX_BINS),bin/$(bin))

help: ## Display make help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
