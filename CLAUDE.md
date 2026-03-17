# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Run server and client locally
make run-server
make run-client

# Tidy dependencies
make tidy

# Build binaries
make build-windows-server   # Windows binary in server/
make build-windows-client   # Windows binary in client/
make build-linux            # Linux binaries for both (server-linux, client-linux)

# Docker
make docker-build           # Build both images
make docker-run-server      # Run server on port 8081
make docker-run-client      # Run client (connects via host.docker.internal:8081)
```

Each component has its own Go module — run `go` commands with `-C server` or `-C client`, or `cd` into the subdirectory first.

## Configuration

Both components load a `.env` file via `godotenv.Load()` at startup (does not override variables already set in the environment, so shell/Docker env always takes precedence).

**`server/.env`**
```
HOST=localhost
PORT=8081
```

**`client/.env`**
```
SERVER_HOST=localhost
SERVER_PORT=8081
```

The client also accepts `BASE_URL` directly (takes priority over `SERVER_HOST`/`SERVER_PORT`).

Docker runs pass `--env-file` at `docker run` time. The `docker-run-server` target reads `SERVER_PORT` from `server/.env` to set the `-p` mapping. The `docker-run-client` target overrides `SERVER_HOST=host.docker.internal` so the client reaches the host machine from inside the container.

## Architecture

Two independent Go binaries in `server/` and `client/`, each with their own `go.mod`.

**Server** (`server/main.go`): In-memory Notes REST API using `go-chi/chi/v5`. All data is stored in a `sync.RWMutex`-protected in-memory map — no database. Listens on `HOST:PORT` (defaults to `localhost:8081`).

API endpoints:
- `POST /notes` → 201
- `GET /notes` → paginated list (`page`, `limit` query params)
- `GET /notes/{id}`
- `PUT /notes/{id}` → full replace
- `PATCH /notes/{id}` → partial update
- `DELETE /notes/{id}` → 204

**Client** (`client/main.go`): Exercises all API operations sequentially as an integration test. `BASE_URL` env var sets the server address (default `http://localhost:8081`). Also includes `client/client.rest` for manual testing with the VS Code REST Client extension.

There are no unit test files (`*_test.go`). The client binary serves as the integration test.
