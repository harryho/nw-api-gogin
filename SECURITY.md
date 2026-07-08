# Security

## Reporting a vulnerability

Please report security issues privately via GitHub Security Advisories: open a private security advisory on the repository (GitHub → Security → Advisories → New draft security advisory).

Do not file public issues for security bugs.

## Supported versions

Only the latest commit on the `main` branch receives security fixes. No formal LTS releases yet.

## Default credentials

The bundled `admin` / `changeit` credentials are **for local development only**. A startup `WARN` log fires when they are in use. Always override `AUTH_ADMIN_USERNAME` and `AUTH_ADMIN_PASSWORD` (or set `AUTH_ADMIN_PASSWORD_HASH`) in any non-local deployment.

## Production checklist

- [ ] `AUTH_ADMIN_PASSWORD_HASH` set (bcrypt pre-hashed); do not ship plaintext.
- [ ] `TOKEN_SECRET` set to a strong random value.
- [ ] `DATABASE_SSL_MODE=require` (or stricter).
- [ ] `APP_ENV=production` (suppresses the default-credentials WARN).
- [ ] `OTEL_EXPORTER_OTLP_INSECURE=false`.
- [ ] Rate limits (`RATE_LIMIT_RPS`, `RATE_LIMIT_BURST`) reviewed.