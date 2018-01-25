GO ?= go
GOVERSION ?= go1.9
OS := $(shell uname)
SHELL := /bin/bash
GIT_VERSION = $(shell git describe --tags)

.DEFAULT_GOAL := help

GOPATH := $(firstword $(subst :, ,$(GOPATH)))
SWAGGER := $(GOPATH)/bin/swagger
GOBINDATA := $(GOPATH)/bin/go-bindata
BUILD := $(shell date +%s)

PKGS := pkg

GIT_VERSION = $(shell git describe --tags)

# Output prefix, defaults to local directory if not specified
ifeq ($(PREFIX),)
	PREFIX := $(shell pwd)
endif

.PHONY: all
all: generate linux darwin

.PHONY: goversion
goversion:
	@echo Checking go version...
	@( $(GO) version | grep -q $(GOVERSION) ) || ( echo "Please install $(GOVERSION) (found: $$($(GO) version))" && exit 1 )

.PHONY: check
check: goversion checkfmt swagger-validate ## check if the source files comply to the formatting rules
	@echo running metalint ...
	# (If errors involves swagger-generated files) consider running "make generate" and retry.)
	gometalinter --disable=gotype --vendor --deadline 30s --fast --errors ./...
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

.PHONY: test
test: ## run tests
	@echo running tests...
	$(GO) test -race -v $(shell go list -v ./... | grep -v /vendor/ | grep -v integration )

.PHONY: swagger-validate
swagger-validate: ## validate the swagger spec
	swagger validate ./swagger/*.yaml

.PHONY: run-dev
run-dev: ## run the dev server
	@./scripts/run-dev.sh

.PHONY: linux
linux: ## build the server binary
	GOOS=linux go build -o bin/image-manager-linux ./cmd/image-manager
	GOOS=linux go build -o bin/function-manager-linux ./cmd/function-manager
	GOOS=linux go build -o bin/identity-manager-linux ./cmd/identity-manager
	GOOS=linux go build -o bin/api-manager-linux ./cmd/api-manager
	GOOS=linux go build -o bin/secret-store-linux ./cmd/secret-store
	GOOS=linux go build -o bin/event-manager-linux ./cmd/event-manager
	GOOS=linux go build -o bin/event-driver-linux ./cmd/event-driver
	GOOS=linux go build -o bin/application-manager-linux ./cmd/application-manager
	GOOS=linux go build -o bin/dispatch-linux ./cmd/dispatch

.PHONY: darwin
darwin: ## build the server binary
	GOOS=darwin go build -o bin/image-manager-darwin ./cmd/image-manager
	GOOS=darwin go build -o bin/function-manager-darwin ./cmd/function-manager
	GOOS=darwin go build -o bin/identity-manager-darwin ./cmd/identity-manager
	GOOS=darwin go build -o bin/secret-store-darwin ./cmd/secret-store
	GOOS=darwin go build -o bin/api-manager-darwin ./cmd/api-manager
	GOOS=darwin go build -o bin/event-manager-darwin ./cmd/event-manager
	GOOS=darwin go build -o bin/event-driver-darwin ./cmd/event-driver
	GOOS=darwin go build -o bin/application-manager-linux ./cmd/application-manager
	GOOS=darwin go build -o bin/dispatch-darwin ./cmd/dispatch

cli-darwin:
	GOOS=darwin go build -o bin/dispatch-darwin ./cmd/dispatch

cli-linux:
	GOOS=linux go build -o bin/dispatch-linux ./cmd/dispatch

.PHONY: images
images: linux ci-images

.PHONY: ci-images
ci-images:
	scripts/images.sh image-manager $(BUILD)
	scripts/images.sh identity-manager $(BUILD)
	scripts/images.sh function-manager $(BUILD)
	scripts/images.sh secret-store $(BUILD)
	scripts/images.sh api-manager $(BUILD)
	scripts/images.sh event-manager $(BUILD)
	scripts/images.sh event-driver $(BUILD)
	scripts/images.sh application-manager $(BUILD)
	scripts/values.sh $(BUILD)

.PHONY: generate
generate: ## run go generate
	scripts/generate.sh image-manager ImageManager image-manager.yaml
	scripts/generate.sh identity-manager IdentityManager identity-manager.yaml
	scripts/generate.sh function-manager FunctionManager function-manager.yaml
	scripts/generate.sh secret-store SecretStore secret-store.yaml
	scripts/generate.sh api-manager APIManager api-manager.yaml
	scripts/generate.sh event-manager EventManager event-manager.yaml
	scripts/generate.sh application-manager ApplicationManager application-manager.yaml
	scripts/header-check.sh fix

.PHONY: gen-clean
gen-clean:  ## Clean all files created with make generate
	rm -rf ./pkg/*/gen

.PHONY: distclean
distclean:  ## Clean ALL files including ignored ones
	git clean -f -d -x .


.PHONY: clean
clean:  ## Clean all modified files
	git clean -f -d .

help: ## Display make help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
