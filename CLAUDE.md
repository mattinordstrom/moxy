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
- `models/` — Mock and Proxy struct definitions
- `config/` — YAML config loader (global `AppConfig`)
- `utils/` — JSON file reading with 1.5s TTL cache, console formatting, file helpers
- `testhelpers/` — Shared test setup utilities

**Admin UI** (`ui/`): Vanilla JS served at `/moxyadminui`. Communicates with backend via WebSocket at `/moxyws`. UI tests use Playwright with a static server on port 9099 that mocks backend endpoints.

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
