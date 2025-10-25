.PHONY: lint test build migrate seed smoke coverage fmt generate integration k6-smoke sbom

GO ?= go
GOLANGCI_LINT ?= golangci-lint
OAPI_CODEGEN ?= $(shell go env GOPATH)/bin/oapi-codegen
BASE_URL ?= http://localhost:8080
AUTH_ADMIN_USERNAME ?= admin
AUTH_ADMIN_PASSWORD ?= changeit

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

generate:
	$(OAPI_CODEGEN) -package api -generate gin,types,spec -o internal/api/api.gen.go api/openapi.yaml
	gofmt -w internal/api/api.gen.go

integration:
	./scripts/integration.sh

k6-smoke:
	BASE_URL=$(BASE_URL) AUTH_ADMIN_USERNAME=$(AUTH_ADMIN_USERNAME) AUTH_ADMIN_PASSWORD=$(AUTH_ADMIN_PASSWORD) k6 run scripts/k6/smoke.js

sbom:
	./scripts/generate-sbom.sh
