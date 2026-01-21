# URL Shortener Service

![Go](https://img.shields.io/badge/go-1.25-00ADD8?logo=go&logoColor=white)
![License](https://img.shields.io/github/license/IDilettant/go-project-278)
[![Actions Status](https://github.com/IDilettant/go-project-278/actions/workflows/hexlet-check.yml/badge.svg)](https://github.com/IDilettant/go-project-278/actions)
[![CI](https://github.com/IDilettant/go-project-278/actions/workflows/ci.yml/badge.svg)](https://github.com/IDilettant/go-project-278/actions/workflows/ci.yml)
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=IDilettant_go-project-278&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=IDilettant_go-project-278)
[![Coverage](https://sonarcloud.io/api/project_badges/measure?project=IDilettant_go-project-278&metric=coverage)](https://sonarcloud.io/summary/new_code?id=IDilettant_go-project-278)
[![Code Smells](https://sonarcloud.io/api/project_badges/measure?project=IDilettant_go-project-278&metric=code_smells)](https://sonarcloud.io/summary/new_code?id=IDilettant_go-project-278)
[![Maintainability Rating](https://sonarcloud.io/api/project_badges/measure?project=IDilettant_go-project-278&metric=sqale_rating)](https://sonarcloud.io/summary/new_code?id=IDilettant_go-project-278)

## Description

URL shortener with REST API, redirect support, and visit analytics. The service exposes link CRUD endpoints, redirects by short code, and stores visits (IP, user-agent, referer, status) in PostgreSQL. It follows a hexagonal/clean architecture and is designed to run behind Caddy (static frontend + reverse proxy) with Gin as the HTTP server.


## Project Structure

- `cmd/api` - application entrypoint.
- `internal/assembly/apiapp` - composition root (wires adapters, middleware, loggers, config).
- `internal/app/links` - use-cases and ports (application layer).
- `internal/domain` - domain models and validation.
- `internal/adapters/httpapi` - Gin handlers, middleware, DTOs, problem+json mapping.
- `internal/adapters/postgres` - repository implementation and sqlc generated code.
- `internal/platform` - config parsing and infrastructure helpers.
- `db/migrations` - database migrations.
- `openapi/openapi.yaml` - OpenAPI spec.

### Architecture (Mermaid)

```mermaid
flowchart LR
  client[Client] -->|HTTP| caddy[Caddy]
  caddy -->|/api, /r| api[Go API (Gin)]
  caddy -->|static assets| ui[Frontend UI]

  api --> handlers[HTTP Handlers]
  handlers --> usecase[links.Service]
  usecase --> repo[Postgres Repo (sqlc)]
  repo --> db[(PostgreSQL)]

  api --> sentry[Sentry]
```

## Getting Started

### Prerequisites

- Go 1.25
- Docker (for local Postgres and migrations)
- Node.js (only if you run `make dev-all` for frontend)

### Local backend (API only)

```bash
cp .env.example .env
make db-up
make migrate-up
make dev
```

`make dev` uses `air` with `.env` to run the Go API on `HTTP_ADDR` (default `:8080`).

### Full stack with Docker (Caddy + Go + static UI)

```bash
docker build -t url-shortener .
PORT=8080 docker run --env-file .env -e PORT=8080 -p 8080:8080 url-shortener
```

Caddy binds to `PORT` and proxies to the Go backend at `HTTP_ADDR` (default `:8080`).

### Swagger UI

```bash
make docs-open-up
```

## Infrastructure

- `Dockerfile` builds frontend assets (Node) and backend binary (Go), then runs on Alpine with Caddy.
- `bin/run.sh` runs Goose migrations, starts Caddy, then launches the Go app.
- `docker-compose.yml` provides local Postgres and a `migrate` job (uses `DATABASE_URL_DOCKER`).
- `Caddyfile` serves static UI from `/app/public` and proxies API/redirect traffic to the Go backend.

## Configuration

Application configuration is loaded from environment variables via `internal/platform/config`.

| Env | Required | Default | Description | Scope |
| --- | --- | --- | --- | --- |
| `HTTP_ADDR` | No | `:8080` | Go backend listen address. `PORT` is reserved for Caddy/platform. | App |
| `BASE_URL` | Yes | - | Public base URL used to build `short_url`. Must be `http(s)` with no path. | App |
| `DATABASE_URL` | Yes | - | PostgreSQL DSN for the app. | App |
| `SENTRY_DSN` | Yes | - | Sentry DSN (required by config). | App |
| `SENTRY_FLUSH_TIMEOUT` | No | `2s` | Flush timeout on shutdown. | App |
| `SENTRY_MIDDLEWARE_TIMEOUT` | No | `2s` | Timeout for Sentry Gin middleware. | App |
| `DB_MAX_OPEN_CONNS` | No | `10` | Max open DB connections. | App |
| `DB_MAX_IDLE_CONNS` | No | `10` | Max idle DB connections. | App |
| `DB_CONN_MAX_LIFETIME` | No | `30m` | Max connection lifetime. | App |
| `HTTP_READ_HEADER_TIMEOUT` | No | `5s` | Server read header timeout. | App |
| `HTTP_READ_TIMEOUT` | No | `15s` | Server read timeout. | App |
| `HTTP_WRITE_TIMEOUT` | No | `15s` | Server write timeout. | App |
| `HTTP_IDLE_TIMEOUT` | No | `60s` | Server idle timeout. | App |
| `HTTP_SHUTDOWN_TIMEOUT` | No | `5s` | Graceful shutdown timeout. | App |
| `REQUEST_BUDGET` | No | `2s` | Request-level context timeout (middleware only; no forced response). | App |
| `CORS_ALLOWED_ORIGINS` | No | empty | Comma-separated origins or `*`. | App |
| `PORT` | Required in container | - | Caddy listen port (Render sets this). Not read by Go config. | Infra |
| `DATABASE_URL_DOCKER` | Optional | - | Used by `docker-compose.yml` migration container. | Tooling |
| `DOCS_URL` | Optional | `http://localhost` | Used by `make docs-open-up`. | Tooling |
| `DOCS_PORT` | Optional | `8081` | Swagger UI port (docs compose). | Tooling |

## Development

```bash
make test            # go test -race ./...
make lint            # golangci-lint run ./...
make build           # build binary
make db-up           # start postgres via docker-compose
make migrate-up      # run goose migrations
make dev             # run API with air
make dev-all         # API + frontend dev server
```

## API Documentation

The OpenAPI spec is at `openapi/openapi.yaml`.

Key endpoints:

- `GET /ping` - health check.
- `GET /api/links` - list links; supports Range pagination.
- `POST /api/links` - create link (returns created resource).
- `GET /api/links/:id` - get by ID.
- `PUT /api/links/:id` - update.
- `DELETE /api/links/:id` - delete.
- `GET /api/link_visits` - list visit events; supports Range pagination.
- `GET /r/:code` - redirect by short code (302) and record visit.

Range pagination accepts either query param or header:

- Query: `?range=[start,count]` (e.g. `?range=[0,10]`)
- Header: `Range: links=0-49` or `Range: 0-49`

Responses include `Content-Range` (e.g. `links 0-9/42`).

## Observability

- **Health check:** `GET /ping`.
- **Sentry:** configured via `SENTRY_DSN` and middleware.
- **Request ID:** middleware sets `X-Request-ID` in responses.
