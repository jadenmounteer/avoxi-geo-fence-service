# AVOXI Geo-Fence Service Standards

## Go Coding Standards

- Use idiomatic Go (Uber style guide preference).
- Favor the Standard Library over heavy frameworks (use `net/http`).
- All business logic must reside in `internal/` to prevent external leakage.
- Use structured logging with `slog` (Standard Library) instead of `fmt.Println`.

## Project Architecture

- **Transport Layer**: Handlers in `cmd/server/` or `internal/api/` handle HTTP/gRPC specifics.
- **Service Layer**: Logic in `internal/geofence/` handles IP validation.
- **Dependency Injection**: Always pass dependencies (like the MaxMind Reader) via constructors (`New...` functions).

## Error Handling

- Wrap errors using `fmt.Errorf("context: %w", err)`.
- Handlers should return clear JSON error messages to the client but log the full technical error to the console.

## Testing Requirement

- Every logic change must include a corresponding `_test.go` file.
- Use table-driven tests for IP validation logic.
