---
description: Rules for writing Go code in the webui backend
---

# Go Coding Rules — WebUI Backend

## File Size
- **Maximum 300 lines per file.** Split by domain (handlers, service, repository) or by feature when the limit is reached.

## Style & Formatting
- `gofmt` / `goimports` are mandatory.
- Import order: stdlib → third-party → project internal (`github.com/.../webui/backend/internal/...`).

## Error Handling
- Explicitly handle every error. Use `http.Error` or structured JSON errors for HTTP handlers.
- Log errors with context before returning them to the client.

## Architecture
- Keep `cmd/main.go` minimal: only wiring (DI, config, server start).
- Business logic lives in `internal/` packages, never in handlers directly.
- Handlers depend on interfaces, not concrete types.

## Naming
- Exported names need godoc comments.
- Handler functions use HTTP-method style where appropriate: `GetUser`, `PostServer`.
- DTO structs for requests/responses end in `Request` / `Response`.

## Testing
- Use `httptest.Server` for handler tests.
- Mock DB / external calls; do not hit real databases in unit tests.
- Run `go test ./...` in `backend/` before committing.
