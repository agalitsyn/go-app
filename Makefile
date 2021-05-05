LOCAL_BIN=$(CURDIR)/bin
PROJECT_NAME=$(shell basename $(CURDIR))

GIT_VERSION := $(shell git describe --long --tags --always --abbrev=8 --dirty)
GIT_BRANCH := $(shell git name-rev --name-only HEAD)

IMPORT_PATH ?= github.com/agalitsyn/go-app
BUILD_VERSION_ARGS := "-X $(IMPORT_PATH)/cmd/internal/flag.version=$(GIT_VERSION)"

ifneq (,$(wildcard ./.env.local))
    include .env.local
    export
endif

RUN_ARGS :=

# Build targets
.PHONE: build
build: ### Build binaries.
	mkdir -p $(LOCAL_BIN)
	CGO_ENABLED=0 go build -v -ldflags $(BUILD_VERSION_ARGS) -o $(LOCAL_BIN) ./cmd/...

.PHONY: clean
clean: ### Clean binaries.
	rm -rf $(LOCAL_BIN)

.PHONY: install
install: ### Run go install.
	go install -race -ldflags $(BUILD_VERSION_ARGS) ./cmd/...

.PHONY: start
start: ### Starts app.
	$(GOENV) go run cmd/api/main.go $(RUN_ARGS)

.PHONY: generate
generate: ### Run go generate
	go generate -v ./...

# Dev targets
.PHONY: what-version
what-version: ### Prints current version.
	@echo $(GIT_VERSION)

include bin-deps.mk

.PHONY: migrate
migrate:
	$(TERN_BIN) migrate --migrations=migrations

.PHONY: lint
lint: $(GOLANGCI_BIN) ### Run golangci-lint.
	$(GOLANGCI_BIN) run ./...

.PHONY: test
test: ### Run go test
	go test -v -race ./...

.PHONY: test-short
test-short: ### Run go test
	go test -v -race -short ./...

.PHONY: test-with-coverage
test-with-coverage: ### Run go test and count code coverage.
	echo "mode: count" > coverage-all.out
	for pkg in $$(go list ./... | grep -v whois/cmd/whois-gateway/design); do \
		go test -coverprofile=coverage.out -covermode=atomic -v $$pkg || exit 1; \
		if [ -f coverage.out ]; then \
			tail -n +2 coverage.out >> coverage-all.out; \
			rm coverage.out; \
		fi; \
	done
	go tool cover -func=coverage-all.out

.PHONY: sloccount
sloccount: ### Count code lines.
	find . -path ./vendor -prune -o -name "*.go" -print0 | xargs -0 wc -l | sort -n

# Docker targets
DOCKER_APPLICATION ?= $(PROJECT_NAME)
DOCKER_TAG ?= $(GIT_VERSION)

DOCKER_REGISTRY ?= hub.docker.com
DOCKER_REGISTRY_REPO ?= agalitsyn
DOCKER_IMAGE := $(DOCKER_REGISTRY_REPO)/$(DOCKER_APPLICATION)
DOCKER_BUILD_ADD_ARGS ?=

.PHONY: docker-build
docker-build:
	docker build --pull --rm --tag "$(DOCKER_IMAGE):$(DOCKER_TAG)" $(DOCKER_BUILD_ADD_ARGS) .
	docker tag "$(DOCKER_IMAGE):$(DOCKER_TAG)" "$(DOCKER_IMAGE):latest"

.PHONY: docker-export
docker-export:
	docker save "$(DOCKER_IMAGE):latest" | gzip --stdout > "$(LOCAL_BIN)/$(DOCKER_REGISTRY_REPO)-$(DOCKER_APPLICATION).tar.gz"

.PHONY: docker-clean
docker-clean:
	docker rmi -f "$$(docker images -q $(DOCKER_IMAGE):$(DOCKER_TAG))"

.PHONY: docker-tag
docker-tag:
	docker tag "$(DOCKER_IMAGE):$(DOCKER_TAG)" "$(DOCKER_REGISTRY)/$(DOCKER_IMAGE):$(DOCKER_TAG)"
	if [ "$(GIT_BRANCH)" = "master" ] || [ "$(CI_COMMIT_REF_NAME)" = "master" ]; then docker tag "$(DOCKER_IMAGE):$(DOCKER_TAG)" "$(DOCKER_REGISTRY)/$(DOCKER_IMAGE):latest"; fi

.PHONY: docker-push
docker-push: docker-tag
	docker push "$(DOCKER_REGISTRY)/$(DOCKER_IMAGE):$(DOCKER_TAG)"
	if [ "$(GIT_BRANCH)" = "master" ] || [ "$(CI_COMMIT_REF_NAME)" = "master" ]; then docker push "$(DOCKER_REGISTRY)/$(DOCKER_IMAGE):latest"; fi

# Helper targets
.PHONY: help
help: ### Show this help.
	@sed -e '/__hidethis__/d; /###/!d; s/:.\+### /\t/g' $(MAKEFILE_LIST)
