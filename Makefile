# See https://tech.davis-hansson.com/p/make/
SHELL := bash
.DELETE_ON_ERROR:
.SHELLFLAGS := -eu -o pipefail -c
.DEFAULT_GOAL := all
MAKEFLAGS += --warn-undefined-variables
MAKEFLAGS += --no-builtin-rules
MAKEFLAGS += --no-print-directory
BIN=$(abspath .tmp/bin)
export PATH := $(BIN):$(PATH)
export GOBIN := $(abspath $(BIN))
COPYRIGHT_YEARS := 2022-2024
LICENSE_IGNORE := --ignore /testdata/ --ignore internal/proto/connectext/

.PHONY: help
help: ## Describe useful make targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "%-30s %s\n", $$1, $$2}'

.PHONY: all
all: ## Build, test, and lint (default)
	$(MAKE) test
	$(MAKE) lint

.PHONY: clean
clean: ## Delete intermediate build artifacts
	@# -X only removes untracked files, -d recurses into directories, -f actually removes files/dirs
	git clean -Xdf

.PHONY: test
test: build ## Run unit tests
	go test -vet=off -race -cover ./...
	cd ./internal/resolvertest && go test -vet=off -race -cover ./...

.PHONY: build
build: generate ## Build all packages
	go build ./...
	cd ./internal/resolvertest && go build ./...

.PHONY: lint
lint: $(BIN)/golangci-lint $(BIN)/buf ## Lint Go and protobuf
	test -z "$$(buf format -d . | tee /dev/stderr)"
	go vet ./...
	cd ./internal/resolvertest && go vet ./...
	golangci-lint run
	buf lint --exclude-path internal/proto/connectext

.PHONY: lintfix
lintfix: $(BIN)/golangci-lint $(BIN)/buf ## Automatically fix some lint errors
	golangci-lint run --fix
	buf format -w .

.PHONY: generate
generate: $(BIN)/buf $(BIN)/protoc-gen-go $(BIN)/license-header services.bin ## Regenerate code and licenses
	rm -rf internal/gen
	PATH=$(abspath $(BIN)) buf generate
	license-header \
		--license-type apache \
		--copyright-holder "The Connect Authors" \
		--year-range "$(COPYRIGHT_YEARS)" $(LICENSE_IGNORE)

.PHONY: upgrade
upgrade: ## Upgrade dependencies
	go get -u -t ./... && go mod tidy -v
	cd ./internal/resolvertest && go get -u -t ./... && go mod tidy -v
	go work sync

.PHONY: checkgenerate
checkgenerate:
	@# Used in CI to verify that `make generate` doesn't produce a diff.
	test -z "$$(git status --porcelain | tee /dev/stderr)"

$(BIN)/buf: Makefile
	@mkdir -p $(@D)
	go install github.com/bufbuild/buf/cmd/buf@v1.29.0

$(BIN)/license-header: Makefile
	@mkdir -p $(@D)
	go install github.com/bufbuild/buf/private/pkg/licenseheader/cmd/license-header@v1.29.0

$(BIN)/golangci-lint: Makefile
	@mkdir -p $(@D)
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.60.0

$(BIN)/protoc-gen-go: Makefile
	@mkdir -p $(@D)
	@# The version of protoc-gen-go is determined by the version in go.mod
	go install google.golang.org/protobuf/cmd/protoc-gen-go

services.bin: $(BIN)/buf
	buf build --as-file-descriptor-set --output $(@F) \
		buf.build/grpc/grpc:26635376b3f47a11126a0f4b4b5b6de7fe5a074a \
		--type grpc.health.v1.Health \
		--type grpc.reflection.v1.ServerReflection \
		--type grpc.reflection.v1alpha.ServerReflection

