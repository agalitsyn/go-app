GOLANGCI_BIN=$(LOCAL_BIN)/golangci-lint
$(GOLANGCI_BIN):
	GOBIN=$(LOCAL_BIN) go install github.com/golangci/golangci-lint/cmd/golangci-lint


TERN_BIN=$(LOCAL_BIN)/tern
$(TERN_BIN):
	GOBIN=$(LOCAL_BIN) go install github.com/jackc/tern
