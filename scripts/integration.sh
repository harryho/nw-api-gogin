#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
cd "$ROOT_DIR"

cleanup() {
  docker compose down --volumes --remove-orphans >/dev/null 2>&1 || true
}

trap cleanup EXIT

# Use a non-default host port to avoid conflicts with locally running Postgres
export DB_PORT=${DB_PORT:-55432}

echo "[integration] Starting PostgreSQL container"
docker compose up -d db >/dev/null

echo "[integration] Waiting for database readiness"
for attempt in {1..30}; do
  if docker compose exec -T db pg_isready -U postgres -d northwind >/dev/null 2>&1; then
    break
  fi
  if [[ $attempt -eq 30 ]]; then
    echo "database did not become ready in time" >&2
    exit 1
  fi
  sleep 2
  echo "  retrying..."
done

export DATABASE_HOST=${DATABASE_HOST:-localhost}
export DATABASE_PORT=${DATABASE_PORT:-$DB_PORT}
export DATABASE_USER=${DATABASE_USER:-postgres}
export DATABASE_PASSWORD=${DATABASE_PASSWORD:-changeme}
export DATABASE_NAME=${DATABASE_NAME:-northwind}
export DATABASE_SSL_MODE=${DATABASE_SSL_MODE:-disable}

echo "[integration] Running migrations"
go run ./cmd/migrate --action up >/dev/null

echo "[integration] Running catalog integration tests"
go test -tags=integration ./internal/catalog -run PostgresIntegration
