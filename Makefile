GO ?= go
GOVERSION ?= go1.9
OS := $(shell uname)
SHELL := /bin/bash
GIT_VERSION = $(shell git describe --tags)

.DEFAULT_GOAL := help

GOPATH := $(firstword $(subst :, ,$(GOPATH)))
SWAGGER := $(GOPATH)/bin/swagger
GOBINDATA := $(GOPATH)/bin/go-bindata

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
	$(GO) test -race $(shell go list -v ./... | grep -v /vendor/ | grep -v integration )

.PHONY: swagger-validate
swagger-validate: ## validate the swagger spec
	swagger validate ./swagger/image-manager.yaml

.PHONY: run-dev
run-dev: ## run the dev server
	@./scripts/run-dev.sh

.PHONY: linux
linux: ## build the server binary
	GOOS=linux go build -o bin/image-manager-linux ./cmd/image-manager

.PHONY: darwin
darwin: ## build the server binary
	GOOS=darwin go build -o bin/image-manager-darwin ./cmd/image-manager

.PHONY: generate
generate: ## run go generate
	go generate ./pkg/image-manager
	scripts/header-check.sh fix

.PHONY: distclean
distclean:  ## Clean ALL files including ignored ones
	git clean -f -d -x .


.PHONY: clean
clean:  ## Clean all modified files
	git clean -f -d .

help: ## Display make help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
