# Matchmaking Service

## Overview
`matchmaking-service` is a Go microservice responsible for account authentication, profile management, profile discovery, like/match state transitions, and user notifications. It is designed to run as part of a broader multi-service system while owning its own persistence boundary in PostgreSQL.

## Architecture Summary
The service is structured in domain modules (auth, profiles, matches, notifications) exposed through HTTP routing and middleware, with PostgreSQL-backed repositories implementing persistence concerns. Request flow is router → middleware → handlers → domain services → repositories, with SQL migrations versioned under `storage/migrations`.

## Run Locally with Docker
```bash
cp .env.example .env
./scripts/start.sh
```

Service endpoint: `http://localhost:${APP_PORT:-8080}`

## Run Database Migrations
```bash
./scripts/migrate.sh
```

This applies `storage/schema/schema.sql` against the configured Postgres container/database.

## Execute Tests
```bash
go test ./...
```

Run integration tests (requires reachable test PostgreSQL):
```bash
go test ./test/integration -count=1
```

## API Documentation (Swagger/OpenAPI)
- OpenAPI spec file: `api/openapi.yaml`
- Swagger serving snippet: `api/swagger-serving-snippet.md`
- If route is enabled in your runtime, access docs at: `GET /swagger`

## Environment Variables
| Variable | Required | Default | Description |
|---|---|---|---|
| `APP_ENV` | No | `development` | Runtime environment label used for deployment/runtime configuration. |
| `APP_PORT` | No | `8080` | Public HTTP port for the service container/process. |
| `SERVER_ADDR` | No | `:8080` | Listen address used by the Go HTTP server. |
| `DATABASE_URL` | Yes | `postgres://postgres:postgres@localhost:5432/matchmaking?sslmode=disable` | PostgreSQL DSN used by repository layer. |
| `POSTGRES_DB` | No | `matchmaking` | Database name for Docker Compose Postgres service. |
| `POSTGRES_USER` | No | `postgres` | Database user for Docker Compose Postgres service. |
| `POSTGRES_PASSWORD` | Yes | `postgres` | Database password for Docker Compose Postgres service. |
| `JWT_SECRET` | Yes | `dev-secret` | Signing secret for JWT access tokens. |
| `ACCESS_TOKEN_TTL_MINUTES` | No | `15` | Access token TTL in minutes. |
| `REFRESH_TOKEN_TTL_HOURS` | No | `168` | Refresh token TTL in hours. |
| `RESET_TOKEN_TTL_MINUTES` | No | `30` | Password reset token TTL in minutes. |
| `PASSWORD_ALGO` | No | `bcrypt` | Password hashing algorithm (`bcrypt` or `argon2`). |
