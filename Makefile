.PHONY: lint test build migrate seed smoke coverage fmt

GO ?= go
GOLANGCI_LINT ?= golangci-lint

fmt:
	gofmt -w $(shell find . -name '*.go' -not -path './vendor/*')

lint: fmt
	$(GOLANGCI_LINT) run ./...

test:
	$(GO) test ./...

coverage:
	$(GO) test ./... -coverprofile=coverage.out

build:
	$(GO) build ./cmd/api

migrate:
	$(GO) run ./cmd/migrate --action up

seed:
	$(GO) run ./cmd/migrate --action seed

smoke:
	restclient --file local-test.http --env local
