.PHONY: lint test build migrate seed smoke coverage fmt generate integration k6-smoke sbom require-env setup-env

# Use the auto-downloaded Go 1.24.5 toolchain from GOMODCACHE if present
# (it has a clean stdlib). The system's /usr/local/go 1.26.5 has a
# corrupted stdlib — duplicate ctrlEmpty/bitsetLSB declarations between
# map.go and map_swiss.go from an overlaid install — and will fail any
# `go build`/`go test` invocation because the bootstrap reads its own
# stdlib BEFORE GOTOOLCHAIN re-exec can kick in. Pointing GO directly
# at the downloaded toolchain bypasses the broken bootstrap entirely.
GOMODCACHE ?= $(shell go env GOMODCACHE 2>/dev/null)
TOOLCHAIN_BIN := $(firstword $(wildcard $(GOMODCACHE)/golang.org/toolchain@v0.0.1-go1.24.5.*/bin/go))
ifeq ($(TOOLCHAIN_BIN),)
    GO ?= go
else
    GO := $(TOOLCHAIN_BIN)
endif
GOLANGCI_LINT ?= golangci-lint
OAPI_CODEGEN ?= $(shell $(GO) env GOPATH)/bin/oapi-codegen
BASE_URL ?= http://localhost:8080
AUTH_ADMIN_USERNAME ?= admin
AUTH_ADMIN_PASSWORD ?= changeit

# The project uses `.envrc` with the `direnv` + `dotenv` extension to
# auto-load `.env` when entering the directory. If `.env` is missing,
# the migration tool falls back to defaults (Password: "") and fails
# with a confusing SASL auth error. Guard against that with a clear
# error message, and provide a `setup-env` target that creates `.env`.
ENV_FILE := .env
ENV_EXAMPLE := .env.example

# Convenience target: create .env from .env.example if it doesn't exist.
# Idempotent — won't overwrite an existing .env.
setup-env:
	@if [ -f $(ENV_FILE) ]; then \
		echo "$(ENV_FILE) already exists, skipping"; \
	else \
		cp $(ENV_EXAMPLE) $(ENV_FILE); \
		echo "Created $(ENV_FILE) from $(ENV_EXAMPLE)"; \
		echo "If you use direnv, .env will be auto-loaded when you cd into this repo."; \
		echo "Otherwise: source .env manually or set the env vars inline."; \
	fi

# Guard: fail fast if .env is missing with a clear actionable message.
require-env:
	@test -f $(ENV_FILE) || { \
		echo "ERROR: $(ENV_FILE) not found."; \
		echo "Run: make setup-env   # or: cp $(ENV_EXAMPLE) $(ENV_FILE)"; \
		echo "The .envrc uses 'direnv' + 'dotenv' to auto-load it; without it the"; \
		echo "migration tool defaults to Password: \"\" and fails with SASL auth."; \
		exit 1; \
	}

fmt:
	gofmt -w $(shell find . -name '*.go' -not -path './vendor/*')

lint: fmt
	$(GOLANGCI_LINT) run ./...

test: require-env
	$(GO) test ./...

coverage: require-env
	$(GO) test ./... -coverprofile=coverage.out

build: require-env
	$(GO) build ./cmd/api

migrate: require-env
	$(GO) run ./cmd/migrate --action up

seed: require-env
	$(GO) run ./cmd/migrate --action seed

# Lightweight smoke check: /healthz, /auth/token, /categories.
# Replaces the previous `restclient` invocation (VS Code REST Client CLI,
# not always installed) with portable curl.
smoke:
	@command -v curl >/dev/null || { echo "curl not installed"; exit 1; }
	@echo "==> GET $(BASE_URL)/healthz"
	@curl -sf $(BASE_URL)/healthz >/dev/null || { echo "FAIL: /healthz"; exit 1; }
	@echo "    OK"
	@echo "==> POST $(BASE_URL)/auth/token"
	@TOKEN=$$(curl -sf -X POST $(BASE_URL)/auth/token \
		-H 'Content-Type: application/json' \
		-d '{"username":"$(AUTH_ADMIN_USERNAME)","password":"$(AUTH_ADMIN_PASSWORD)","scope":"viewer"}' \
		| sed -E 's/.*"accessToken":"([^"]+)".*/\1/'); \
	test -n "$$TOKEN" || { echo "FAIL: /auth/token"; exit 1; }
	@echo "    OK"
	@echo "==> GET $(BASE_URL)/categories (viewer scope)"
	@curl -sf -H "Authorization: Bearer $$TOKEN" $(BASE_URL)/categories >/dev/null || { echo "FAIL: /categories"; exit 1; }
	@echo "    OK"

generate:
	$(OAPI_CODEGEN) -package api -generate gin,types,spec -o internal/api/api.gen.go api/openapi.yaml
	gofmt -w internal/api/api.gen.go

integration:
	./scripts/integration.sh

k6-smoke:
	BASE_URL=$(BASE_URL) AUTH_ADMIN_USERNAME=$(AUTH_ADMIN_USERNAME) AUTH_ADMIN_PASSWORD=$(AUTH_ADMIN_PASSWORD) k6 run scripts/k6/smoke.js

sbom:
	./scripts/generate-sbom.sh