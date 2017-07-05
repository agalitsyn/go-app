PROJECT_PKGS := $$(go list ./... | grep -v /vendor/)

APPLICATION ?= $(shell basename $(CURDIR))

REGISTRY := hub.docker.com
TEAM := agalitsyn
IMAGE := $(TEAM)/$(APPLICATION)
IMAGE_TAG ?= $(shell git describe --long --tags --always --abbrev=8)

BUILD_DIR ?= bin

.PHONY: all
all: clean $(BUILD_DIR)

$(BUILD_DIR):
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -a -installsuffix cgo -o $(BUILD_DIR)/$(APPLICATION) .

.PHONY: clean
clean:
	-rm -r $(BUILD_DIR)

.PHONY: what-version
what-version:
	@echo $(REGISTRY)/$(IMAGE):$(IMAGE_TAG)

# Docker
NOROOT := -u $$(id -u):$$(id -g)
SRCDIR := /go/src/github.com/$(TEAM)/$(APPLICATION)
DOCKERFLAGS := --rm=true $(NOROOT) -v $(CURDIR):$(SRCDIR) -w $(SRCDIR)
BUILD_IMAGE := golang
BUILD_IMAGE_TAG := 1.8

.PHONY: docker-build
docker-build:
	docker run $(DOCKERFLAGS) $(BUILD_IMAGE):$(BUILD_IMAGE_TAG) make

.PHONY: docker-image
docker-image:
	docker build --rm --pull --tag $(IMAGE):$(IMAGE_TAG) .

.PHONY: docker-push
docker-push:
	docker tag $(IMAGE):$(IMAGE_TAG) $(REGISTRY)/$(IMAGE):$(IMAGE_TAG)
	docker push $(REGISTRY)/$(IMAGE):$(IMAGE_TAG)

.PHONY: docker-clean
docker-clean:
	-docker rmi -f $$(docker images $(IMAGE) -q)

# Dev
.PHONY: infra-start
infra-start:
	docker-compose build --pull
	docker-compose up --force-recreate -d

.PHONY: infra-stop
infra-stop:
	docker-compose down

.PHONY: install
install:
	go install .

.PHONY: start
start: install
	$$GOPATH/bin/goreman start

.PHONY: lint
lint:
	$$GOPATH/bin/gometalinter \
		--vendor \
		--tests \
		--errors \
		--deadline 360s \
		--concurrency 2 \
		--sort=severity \
		--format='({{.Linter}}) | {{.Severity}} | {{.Path}}:{{.Line}}:{{if .Col}}{{.Col}}{{end}} | {{.Message}}' \
		--disable-all \
		--enable goimports \
		--enable vet \
		--enable vetshadow \
		--enable golint \
		--enable errcheck \
		--enable staticcheck \
		./...

.PHONY: test
test:
	echo "mode: count" > coverage-all.out
	for pkg in $(PROJECT_PKGS); do \
		go test -coverprofile=coverage.out -covermode=count -v -race $$pkg || exit 1 ; \
		if [ -f coverage.out ]; then \
			tail -n +2 coverage.out >> coverage-all.out; \
			rm coverage.out; \
		fi; \
	done
	go tool cover -func=coverage-all.out

.PHONY: sloccount
sloccount:
	find . -path ./vendor -prune -o -name "*.go" -print0 | xargs -0 wc -l | sort -n

.PHONY: godoc
godoc:
	@echo "Open in browser: http://localhost:6060/pkg/github.com/agalitsyn/goapi"
	godoc -http :6060

PACKAGES := \
	golang.org/x/tools/cmd/goimports \
	github.com/mattn/goreman \
	github.com/alecthomas/gometalinter \
	github.com/kardianos/govendor

.PHONY: install-packages
install-packages:
	$(foreach pkg,$(PACKAGES),go get -u $(pkg);)
	$$GOPATH/bin/gometalinter --install --update
