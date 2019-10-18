BUILD_DIR ?= bin
BINARIES_DIR := cmd
BINARIES ?= $$(find $(BINARIES_DIR) -maxdepth 1 \( ! -iname "$(BINARIES_DIR)" \) \( ! -iname internal \) -type d -exec basename {} \;)

GIT_VERSION := $(shell git describe --long --tags --always --abbrev=8 --dirty)
GIT_BRANCH := $(shell git name-rev --name-only HEAD)

IMPORT_PATH ?= github.com/agalitsyn/goapi
BUILD_VERSION_ARGS := "-X $(IMPORT_PATH)/cmd/internal/flag.version=$(GIT_VERSION)"

# Build targets
.PHONE: build
build: ### Build binaries.
	for bin in $(BINARIES); do \
		go build -o $(BUILD_DIR)/$$bin -ldflags $(BUILD_VERSION_ARGS) $(IMPORT_PATH)/$(BINARIES_DIR)/$$bin;\
	done

.PHONY: clean
clean: ### Clean binaries.
	for bin in $(BINARIES); do \
		rm -f $$bin/$$bin;\
	done

.PHONY: install
install: ### Run go install.
	go install -race -ldflags $(BUILD_VERSION_ARGS) ./cmd/...

.PHONY: generate
generate: ### Run go generate
	go generate -v ./...

# Dev targets
.PHONY: what-version
what-version: ### Prints current version.
	@echo $(GIT_VERSION)

.PHONY: lint
lint: ### Run govet.
	go vet ./...

PHONY: test
test: ### Run go test. Ensure that backing service are up using `make compose-start-testing`.
	go test -v -race ./...

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
DOCKER_APPLICATION ?= $(shell basename $(CURDIR))
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
	docker save "$(DOCKER_IMAGE):latest" | gzip --stdout > "$(BUILD_DIR)/$(DOCKER_REGISTRY_REPO)-$(DOCKER_APPLICATION).tar.gz"

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

.PHONY: docker-reset
docker-reset: containers := $$(docker ps -aq)
docker-reset: images := $$(docker images --filter dangling=true -qa)
docker-reset: volumes := $$(docker volume ls --filter dangling=true -q)
docker-reset: ### Totally resets docker entities (containers, images, networks, volumes).
	if [ "$(containers)" ]; then docker stop $(containers) && docker rm -f $(containers); fi
	-docker network prune -f
	if [ "$(images)" ]; then docker rmi -f $(images); fi
	if [ "$(volumes)" ]; then docker volume rm -f $(volumes); fi

# Helper targets
.PHONY: help
help: ### Show this help.
	@sed -e '/__hidethis__/d; /###/!d; s/:.\+### /\t/g' $(MAKEFILE_LIST)
