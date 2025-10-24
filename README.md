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
