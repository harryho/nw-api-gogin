# Network API for Go + Gin

## Architecture Overview
This project implements a network commerce API built with Go and Gin. It provides category, product, and supplier management backed by PostgreSQL and is designed for deployment on AWS Lambda behind API Gateway.

### Key Components
- `go-gin-boilerplate`: baseline Gin application structure. We extend it with layered repositories, services, and OpenAPI-first scaffolding.
- `db-samples`: provides relational data samples used for schema inspiration. We adapt naming to align with domain-driven aggregates and add auditing columns.
- We introduce AWS Lambda deployment, OpenTelemetry instrumentation, and RBAC that are not present in the reference repositories.

### Component Diagram
```mermaid
graph LR
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
- `infra/` infrastructure-as-code (AWS CDK stack and deployment scripts)

## Setup
- Install Go 1.23, Docker, and Docker Compose.
- Copy `.env.example` to `.env` and adjust database credentials if necessary.
- Start dependencies with `docker compose up -d db otel-collector prometheus`.
- Run `go mod tidy` (only required after dependency updates).
- Launch the API locally with `go run ./cmd/api` or `make build && ./cmd/api`.
- Generate an admin token via `/auth/token` (defaults `admin` / `changeit`).

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

## Database Migrations
- `make migrate` applies the latest migrations to the configured database.
- `make seed` loads sample data defined under `internal/db/migrations`.
- `go run ./cmd/migrate --action up` / `--action down` runs the migration CLI directly.
- Local development uses Dockerized Postgres; production environments should supply managed database credentials via environment variables or Secrets Manager.

## Testing
- `make test` runs the Go unit and integration suite.
- `make integration` spins up PostgreSQL via Docker Compose, runs database migrations, and executes the catalog integration suite against the real database (requires Docker).
- The integration harness exposes Postgres on `DB_PORT` (default `55432`) so it can run alongside a local Postgres instance.
- `make smoke` exercises the API using `local-test.http`; update `local-test.http` if environment values differ.
- `k6 run scripts/k6/smoke.js` (or `make k6-smoke`) issues an auth/token request, lists categories, and creates then deletes a category to validate basic flows; the scenario enforces zero failures and a 95th percentile latency under 500ms with default `K6_ITERATIONS=10` and `K6_VUS=1`.
- `make sbom` generates a CycloneDX SBOM at `sbom/bom.json` using `cyclonedx-gomod`.
- `make build-lambda` cross-compiles the API for the Lambda execution environment (`linux/amd64`) and writes the binary to `dist/bootstrap`.
- `make package-lambda` builds and zips the Lambda artifact at `dist/bootstrap.zip` for deployment.

## Deployment
- Generate the Lambda artifact with `make package-lambda` (runs `build-lambda` as a prerequisite).
- Review `infra/cdk/README.md` for instructions on bootstrapping the AWS CDK app, synthesizing the stack, and deploying to an AWS account.
- The CDK stack provisions Secrets Manager credentials, API Gateway, a Lambda alias with provisioned concurrency, and a CloudWatch alarm feeding an SNS topic. Update database environment variables before promoting to production.

### CI/CD Pipeline
- The `deploy` job in `.github/workflows/ci.yml` runs on pushes to `main` when the repository variable `ENABLE_CDK_DEPLOY` is set to `true`.
- Provide AWS access by storing `AWS_DEPLOY_ROLE_ARN` and `AWS_DEPLOY_REGION` as encrypted repository secrets (the job assumes the role via OpenID Connect).
- Optional repository variables (`PROVISIONED_CONCURRENCY`, `CANARY_PERCENTAGE`, `CANARY_INTERVAL`) feed stack parameters so each deployment can tune traffic shifting and pre-warmed capacity.
- The job downloads the `lambda-bootstrap` artifact produced by the build stage, unpacks it into `dist/`, installs the CDK toolchain, and executes ` NorthwindApiStack --require-approval never`.

### Rollback & Promotion
- Deployments use AWS  canary routing (default 10% traffic for 5 minutes) backed by a CloudWatch error alarm; failures trigger automatic rollback and, if configured, SNS notifications.
- To promote a deployment manually or adjust concurrency targets, rerun `` locally or via the pipeline with updated parameter values (for example, `--parameters NorthwindApiStack:ProvisionedConcurrency=10`).
- For emergency rollback, redeploy the previously known-good artifact (downloadable from the `lambda-bootstrap` workflow artifact history) or run `` with the desired Git commit checked out.

## Security & Observability
- Default response headers include `Content-Security-Policy`, `X-Frame-Options`, `X-Content-Type-Options`, and `Strict-Transport-Security`. Set `DISABLE_HSTS=true` if running over plain HTTP in non-production environments.
- Rate limiting is enabled per client IP with `RATE_LIMIT_RPS` (default `25`) and `RATE_LIMIT_BURST` (default `50`). Requests beyond the burst window receive `429` responses.
- Audit logs are emitted for mutating verbs (POST/PUT/PATCH/DELETE) and include subject, scopes, latency, and request identifier.
- OpenTelemetry tracing is configured via `OTEL_EXPORTER_OTLP_ENDPOINT`, optional `OTEL_EXPORTER_OTLP_HEADERS`, and `OTEL_EXPORTER_OTLP_INSECURE=true` when sending to an insecure collector. Service identity defaults can be tuned with `SERVICE_NAME`, `SERVICE_VERSION`, and `APP_ENV`.
- A scheduled `govulncheck` workflow scans dependencies weekly; SBOM artifacts are generated in CI for downstream transparency.

## Troubleshooting
- **Database connection refused**: ensure `docker compose up db` is running and `DATABASE_HOST/PORT` match the container settings.
- **Migrations fail due to missing schema**: run `make migrate` followed by `make seed` to bootstrap the database; verify credentials in `.env`.
- **`make package-lambda` errors about `zip`**: install the system `zip` utility (`brew install zip` on macOS).
- **`cyclonedx-gomod: command not found`**: rerun `make sbom`; the script installs the tool into `$(go env GOPATH)/bin` when absent.
- **CI deploy job skipped**: confirm the repo variable `ENABLE_CDK_DEPLOY` is `true` and secrets `AWS_DEPLOY_ROLE_ARN` / `AWS_DEPLOY_REGION` are configured.
