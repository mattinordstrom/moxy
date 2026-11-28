# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Moxy is a Go-based HTTP mocking and proxying tool for localhost development. It matches incoming requests against mock definitions first, then proxy definitions, and falls back to a default route. It includes a web-based admin UI with WebSocket-powered real-time updates.

## Commands

### Run
```bash
go run main.go          # Default port 9097
go run main.go -p 8080  # Custom port
go run main.go -s       # HTTPS mode
go run main.go -l       # List redirects without starting server
go run main.go -a       # Listen on all interfaces (default: localhost only)
go run main.go -v       # Print version and exit
```

### Release
The version lives in `version/version.go` and is shown in the terminal banner and the admin UI header.

When finishing a user-facing change (feature, fix, config or mockdef format change), ask whether to bump the version and release, and suggest the semver level: patch for fixes, minor for features or new optional fields, major for breaking changes. Never bump, tag or release without being asked.

```bash
# 1. Bump version.Version and commit
# 2. Tag and push; .github/workflows/release.yml checks the tag matches version.Version and creates the GitHub release
git tag v2.0.0 && git push origin v2.0.0
```

### Test
```bash
go test ./... -v                          # All Go tests
go test ./httphandling -v -run TestName   # Single Go test
cd ui/tests && npm install && npm test    # Playwright UI tests
```

### Lint
```bash
golangci-lint run -v
```

## Architecture

**Request flow:** `main.go` starts the HTTP server → `httphandling/httphandler.go` matches requests against mocks (top-to-bottom, first match wins, regex `.*` supported) → falls back to proxy definitions → falls back to default route from `config.yml`.

**Key packages:**
- `httphandling/` — Core HTTP handler, admin API/WebSocket endpoints, WebSocket mock handler
- `httphandling/eventlog.go` — Redacted request/error log (memory + `moxy.log`, served at `/moxyadminui/log`); never add headers, bodies, payloads or query values to it
- `models/` — Mock and Proxy struct definitions
- `config/` — YAML config loader (global `AppConfig`)
- `utils/` — JSON file reading with 1.5s TTL cache, console formatting, file helpers
- `testhelpers/` — Shared test setup utilities

**Admin UI** (`ui/`): Vanilla JS served at `/moxyadminui`. The admin API and `/moxyws` are guarded in `httphandling/adminguard.go`: loopback client address and localhost `Host` only, even with `-a`, same-origin and `application/json` for writes, and no CORS headers. Mock/proxy routes and `/moxywsmock` accept loopback origins, same-origin and `cors.allowedOrigins` from `config.yml`; other origins get 403. The UI must escape anything it inserts as HTML with `escapeHTML`. Mock payloads are kept as JSON source text (`JSONText` in `utils.js`), never parsed and re-serialized, and every save first re-reads the file from disk (`syncWithDisk`) so stale tabs and external edits aren't overwritten. Communicates with backend via WebSocket at `/moxyws`. UI tests use Playwright with a static server on port 9099 that mocks backend endpoints.

**WebSocket mock** at `/moxywsmock`: Broadcasts messages to all connected clients.

**Config files** (auto-generated on first run from `templates/`):
- `config.yml` — Default route and payload archive path
- `mockdef.json` — Mock definitions (matched by URL substring/regex)
- `proxydef.json` — Proxy definitions (forward to target servers)

These JSON files are cached with a 1.5s TTL — changes take effect without restart.

## Dependencies

Only two external dependencies: `gorilla/websocket` and `gopkg.in/yaml.v3`.

## Pre-commit Hooks

Configured in `.pre-commit-config.yaml`: runs golangci-lint, Go tests, and UI tests.
