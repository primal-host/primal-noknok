# noknok

AT Protocol authentication gateway for Traefik forwardAuth.

## Project Structure

- `cmd/noknok/` — Entry point
- `internal/config/` — Environment + file-based config
- `internal/database/` — pgx pool, schema bootstrap
- `internal/atproto/` — Handle resolution, PDS discovery, createSession
- `internal/session/` — Server-side session management + cookies
- `internal/server/` — Echo HTTP server, routes, handlers

## Build & Run

```bash
go build -o noknok ./cmd/noknok
go vet ./...

# Docker
./.launch.sh
```

## Conventions

- Go module: `github.com/primal-host/noknok`
- HTTP framework: Echo v4
- Database: pgx v5 on infra-postgres, database `noknok`
- Container name: `primal-noknok`
- Schema auto-bootstraps via `CREATE TABLE IF NOT EXISTS`
- Config uses env vars with `_FILE` suffix support for Docker secrets

## Database

Postgres on `infra-postgres:5432` (host port 5433), database `noknok`, user `dba_noknok`.

Tables: `sessions`, `users`, `services`, `grants`.

## Docker

- Image/container: `primal-noknok`
- Network: `infra` (postgres/traefik)
- Port: 4321
- Traefik: `noknok.primal.host` / `noknok.localhost`
- DNS: `192.168.147.53` (infra CoreDNS)
- Defines `noknok-auth` forwardAuth middleware for other services

## Auth Flow

1. Unauthenticated request to protected service
2. Traefik calls `GET /auth` on noknok (forwardAuth)
3. No valid session cookie → 302 redirect to `/login?redirect=...`
4. User logs in with AT Protocol handle + app password
5. noknok resolves handle → DID → PDS, calls createSession on PDS
6. DID verified against OWNER_DID → session created → cookie set
7. Redirect back to original service → forwardAuth passes with X-User-DID header
