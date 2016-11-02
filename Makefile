.PHONY: all
all: install-tools generate-certificates start

.PHONY: build
build:
	go install .

.PHONY: start
start: build
	goreman start

.PHONY: format
format:
	goimports -w .

GOMETALINTER_REQUIRED_FLAGS := --vendor --tests --errors
# gotype is broken, see https://github.com/alecthomas/gometalinter/issues/91
GOMETALINTER_COMMON_FLAGS := --concurrency 2 --deadline 60s --line-length 120 --enable lll --disable gotype

.PHONY: lint
lint:
	gometalinter \
		$(GOMETALINTER_COMMON_FLAGS) \
		$(GOMETALINTER_REQUIRED_FLAGS) \
		.

.PHONY: check
check:
	gometalinter \
		--enable goimports \
		--disable errcheck \
		--disable golint \
		--fast \
		$(GOMETALINTER_COMMON_FLAGS) \
		$(GOMETALINTER_REQUIRED_FLAGS) \
		.

.PHONY: test
test: lint
	go test -cover -v .

.PHONY: sloccount
sloccount:
	find . -path ./vendor -prune -o -name "*.go" -print0 | xargs -0 wc -l

.PHONY: info
info:
	depscheck -totalonly -tests .

.PHONY: std-info
std-info:
	depscheck -stdlib -v .

PACKAGES := \
	golang.org/x/tools/cmd/goimports \
	github.com/mattn/goreman \
	github.com/tools/godep \
	github.com/alecthomas/gometalinter \
	github.com/divan/depscheck

.PHONY: install-tools
install-tools:
	$(foreach pkg,$(PACKAGES),go get -u $(pkg);)
	gometalinter --install --update

.PHONY: generate-certificates
generate-certificates:
	go run certgen/main.go
