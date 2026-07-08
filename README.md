# Northwind API (Go + Gin)

A small reference REST API providing Category / Product / Supplier management on top of PostgreSQL.

## Stack

- Go 1.24.5 + Gin
- GORM (PostgreSQL driver)
- JWT auth (HMAC-SHA256, bcrypt-hashed passwords)
- OpenTelemetry tracing + Prometheus metrics
- golang-migrate for schema migrations
- oapi-codegen for the OpenAPI 3.1 contract

## Quickstart

```bash
docker compose up -d db otel-collector prometheus
make migrate
make test
make smoke
```

Then point a REST client (e.g. `local-test.http`) at `http://localhost:8080`.

Default local credentials (dev only - see SECURITY): `admin` / `changeit`.

## Configuration

All config is via environment variables. See `.env.example`. Notable env vars:

- `DATABASE_HOST`, `DATABASE_PORT`, `DATABASE_USER`, `DATABASE_PASSWORD`, `DATABASE_NAME`, `DATABASE_SSL_MODE`
- `AUTH_ADMIN_USERNAME`, `AUTH_ADMIN_PASSWORD` (plaintext) or `AUTH_ADMIN_PASSWORD_HASH` (bcrypt; recommended for prod)
- `TOKEN_SECRET`, `TOKEN_KEY_ID`, `TOKEN_TTL`, `TOKEN_AUDIENCE`, `TOKEN_ISSUER`
- `RATE_LIMIT_RPS`, `RATE_LIMIT_BURST`
- `OTEL_EXPORTER_OTLP_ENDPOINT`, `OTEL_EXPORTER_OTLP_HEADERS`, `OTEL_EXPORTER_OTLP_INSECURE`
- `SERVICE_NAME`, `SERVICE_VERSION`, `APP_ENV`

## Deployment

This is a plain HTTP service. Build the binary with `make build` (output: `./build/app`) or the Docker image with `docker build -t nw-api .`.

For cloud-specific deployment recipes (AWS / Azure / GCP) and cost comparisons, see the local-only deployment guide (kept out of git). Recommended: **GCP Cloud Run + Cloud SQL `db-f1-micro`** (~$10-15/mo at low traffic). Runner-up: **AWS Lightsail** at ~$20/mo flat.

## Development

```bash
make fmt        # gofmt
make lint       # golangci-lint run
make test       # go test ./...
make coverage   # tests + coverage.out
make generate   # regenerate internal/api/api.gen.go from api/openapi.yaml
make sbom       # cyclonedx SBOM
```

## Security

See [SECURITY.md](./SECURITY.md). Default local credentials are dev-only; override before any non-local deployment.

## License

MIT. See [LICENSE](./LICENSE).

## Contributing

See [CONTRIBUTING.md](./CONTRIBUTING.md).