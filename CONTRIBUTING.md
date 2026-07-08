# Contributing

Thanks for your interest in contributing.

## Local setup

```bash
docker compose up -d db otel-collector prometheus
cp .env.example .env
make migrate
make test
```

## Workflow

1. Fork and create a feature branch.
2. Make your change with tests. Run `make fmt && make lint && make test`.
3. Regenerate the API client if you touched `api/openapi.yaml`: `make generate`.
4. Open a pull request describing the change.

## Commit messages

Use [Conventional Commits](https://www.conventionalcommits.org/) prefixes: `feat:`, `fix:`, `chore:`, `refactor:`, `docs:`, `test:`, `build:`, `ci:`.

## Code style

- `gofmt` + `golangci-lint run` must pass.
- Public functions and types have doc comments.
- Tests live next to the code (`foo.go` + `foo_test.go`).

## Reporting vulnerabilities

See [SECURITY.md](./SECURITY.md). Do not file public issues for security bugs.