PROJECT_PKGS := $$(go list ./... | grep -v /vendor/)

GOMETALINTER_REQUIRED_FLAGS := --vendor --tests --errors
# gotype is broken, see https://github.com/alecthomas/gometalinter/issues/91
GOMETALINTER_COMMON_FLAGS := --concurrency 2 --deadline 60s --line-length 120 --enable lll --disable gotype


.PHONY: format
format:
	go fmt ./...
	goimports -w .

.PHONY: lint
lint:
	gometalinter \
		$(GOMETALINTER_COMMON_FLAGS) \
		$(GOMETALINTER_REQUIRED_FLAGS) \
		./...

.PHONY: check
check:
	gometalinter \
		--enable goimports \
		--disable errcheck \
		--disable golint \
		--fast \
		$(GOMETALINTER_COMMON_FLAGS) \
		$(GOMETALINTER_REQUIRED_FLAGS) \
		./...

.PHONY: test
test: lint
	for pkg in $(PROJECT_PKGS); do \
        go test -cover -v $$pkg || exit 1 ;\
    done

.PHONY: test-docker
test-docker:
	docker run -it -v $$(pwd):/go/src/$$(go list .) agalitsyn/goci:1.6 bash -c "cd /go/src/$$(go list .) && make test"

.PHONY: sloccount
sloccount:
	find . -path ./vendor -prune -o -name "*.go" -print0 | xargs -0 wc -l

.PHONY: info
info: sloccount
	depscheck -totalonly -tests $(PROJECT_PKGS)

.PHONY: std-info
std-info: sloccount
	depscheck -stdlib -v $(PROJECT_PKGS)
