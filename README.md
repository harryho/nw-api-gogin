# Network API for Go + Gin

## Overview
This project implements a network commerce API built with Go and Gin. It provides category, product, and supplier management backed by PostgreSQL and is designed for deployment on AWS Lambda behind API Gateway.

## Architecture Summary
- `go-gin-boilerplate`: baseline Gin application structure. We extend it with layered repositories, services, and OpenAPI-first scaffolding.
- `db-samples`: provides relational data samples used for schema inspiration. We adapt naming to align with domain-driven aggregates and add auditing columns.
- We introduce AWS Lambda deployment, OpenTelemetry instrumentation, and RBAC that are not present in the reference repositories.

### Component Diagram
```mermaid
graph TD
  Client -->|HTTPS| APIGW[API Gateway]
  APIGW --> Lambda
  Lambda --> GinRouter[GIN Router]
  GinRouter --> ServiceLayer
  ServiceLayer --> Repository
  Repository --> PostgreSQL[(PostgreSQL)]
  ServiceLayer --> AuthN[JWT Auth]
  ServiceLayer --> Metrics[Prometheus Metrics]
```

## Repository Layout
- `cmd/` application entrypoints (`api`, `migrate`)
- `internal/` domain logic, repositories, and services
- `pkg/` shared utilities (`config`, `logger`, `auth`)
- `configs/` static configuration templates
- `scripts/` automation helpers for local and CI workflows
- `docs/` architecture, roadmap, and implementation notes
- `testdata/` seed and fixture files
- `api/` OpenAPI specifications and codegen configuration

## Next Steps
Follow the roadmap in `docs/implementation-roadmap.md` to execute each phase.

## API Contract
- The canonical OpenAPI definition lives at `api/openapi.yaml` and models CRUD workflows for categories, products, and suppliers along with JWT scope expectations.
- Regenerate Gin handlers and types after editing the specification with `make generate` (requires `oapi-codegen` on your `PATH`).

## Authentication
- All category, product, and supplier routes require a bearer token with the scopes defined in the OpenAPI document (viewer for read, manager for updates, admin for destructive operations).
- A static administrator account is provided via the environment variables `AUTH_ADMIN_USERNAME` and `AUTH_ADMIN_PASSWORD`; defaults are `admin` / `changeit` for local development only.
- Signing configuration is controlled through `TOKEN_SECRET`, `TOKEN_KEY_ID`, `TOKEN_TTL`, `TOKEN_AUDIENCE`, and `TOKEN_ISSUER`.

### Obtaining a Token Locally
```bash
curl -s \
  -X POST http://localhost:8080/auth/token \
  -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"changeit","scope":"viewer"}'
```

Use the returned `accessToken` in the `Authorization: Bearer <token>` header when calling protected endpoints.
