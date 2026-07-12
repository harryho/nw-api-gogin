# Northwind API (Go + Gin)

A small reference REST API providing Category / Product / Supplier management on top of PostgreSQL.

## Stack

- Go 1.25.0 + Gin
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


## Development

```bash
make fmt        # gofmt
make lint       # golangci-lint run
make test       # go test ./...
make coverage   # tests + coverage.out
make generate   # regenerate internal/api/api.gen.go from api/openapi.yaml
make sbom       # cyclonedx SBOM
```

## Test coverage

`make coverage` runs all tests with `-coverprofile=coverage.out` and prints the per-package coverage summary inline. `coverage.out` and `coverage.html` are gitignored (regenerated on each run).

```bash
make coverage                                       # run tests, write coverage.out
go tool cover -func=coverage.out                     # per-function summary in terminal
go tool cover -html=coverage.out -o coverage.html   # HTML report; open in a browser
```

Packages without tests (`internal/app`, `internal/db`, `pkg/telemetry`, `cmd/*`) are reached through integration tests rather than unit tests — see `make integration`.

## Security

See [SECURITY.md](./SECURITY.md). Default local credentials are dev-only; override before any non-local deployment.

## License

MIT. See [LICENSE](./LICENSE).

## Contributing

See [CONTRIBUTING.md](./CONTRIBUTING.md).