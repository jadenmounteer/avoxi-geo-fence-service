# Code Review Feedback

**Reviewer perspective:** Senior/Staff Engineer conducting a technical interview code review.

**Scope:** AVOXI Geo-Fence Service – Go microservice with dual HTTP/gRPC transport, MaxMind GeoLite2 integration, and Kubernetes deployment.

---

## Summary

The codebase demonstrates solid engineering fundamentals: clear separation of concerns, idiomatic Go, and good test coverage. The dual-transport design (HTTP + gRPC) with shared business logic is well executed. A few production-readiness gaps and minor consistency issues remain.

**Overall assessment:** Strong hire with targeted improvements.

---

## Strengths

### Architecture and Design

- **Layered structure:** Transport (`internal/api`), business logic (`internal/geofence`), and generated protos (`internal/pb`) are clearly separated. `internal/` correctly prevents external leakage.
- **Dependency injection:** Dependencies are passed via constructors (`NewChecker`, `NewCheckHandler`, etc.). The `CountryLookuper` interface allows testing without the real GeoIP database.
- **Shared validation:** Typed errors (`ErrEmptyAllowedCountries`, `ErrInvalidIP`, `ErrUnknownIP`) centralize validation in the checker; HTTP and gRPC handlers map them to transport-specific responses. This avoids duplication and keeps behavior consistent.
- **Dual transport:** HTTP and gRPC share the same `Checker` and `GeoStore`. `HealthHandler` correctly implements both HTTP probes and gRPC `CheckHealth` via a shared `isReady()`.

### Code Quality

- **Table-driven tests:** `check_handler_test.go`, `grpc_server_test.go`, `checker_test.go` use table-driven tests with good coverage of success, validation, and error cases.
- **Error handling:** Errors are wrapped with `fmt.Errorf` and `%w`; `errors.Is` is used correctly for sentinel checks.
- **Logging:** Structured logging with `slog`; appropriate levels (Info, Warn, Error).

### DevOps and Tooling

- **Docker:** Multi-stage build, non-root user, minimal Alpine image.
- **Kubernetes:** Probes, resource limits/requests, and a dedicated Service manifest.
- **Makefile:** Clear targets for build, test, proto generation, Docker, and K8s.
- **Proto:** Proto3, reflection enabled, consistent with grpc-api rules.

---

## Areas for Improvement

### 1. Production Readiness

| Issue                     | Location                 | Recommendation                                                                                                                                                                               |
| ------------------------- | ------------------------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **No request size limit** | `CheckHandler.ServeHTTP` | Use `http.MaxBytesReader` to cap JSON body size and reduce DoS risk.                                                                                                                         |
| **HTTP server timeouts**  | `main.go`                | Set `ReadHeaderTimeout`, `ReadTimeout`, `WriteTimeout`, and `IdleTimeout` on `http.Server` to protect against slow clients.                                                                  |
| **Shutdown order**        | `main.go`                | Call `httpServer.Shutdown` first (or both in parallel with a context); then `grpcServer.GracefulStop`. Both block until connections drain; parallel shutdown can reduce total shutdown time. |

### 2. Deployment and Configuration

| Issue                       | Location                                  | Recommendation                                                                                                                             |
| --------------------------- | ----------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------ |
| **gRPC not exposed in K8s** | `k8s/deployment.yaml`, `k8s/service.yaml` | Add `containerPort: 9090` and a second Service port (e.g. `port: 9090`, `targetPort: 9090`, `name: grpc`) so gRPC is reachable in-cluster. |
| **Docker ports**            | `Dockerfile`, `Makefile`                  | Add `EXPOSE 9090` and `-p 9090:9090` in `docker-run` so gRPC is accessible in local Docker runs.                                           |
| **K8s env vars**            | `k8s/deployment.yaml`                     | Add `HTTP_PORT` and `GRPC_PORT` to align with main.go’s config and document port usage.                                                    |

### 3. Dependencies and Module Hygiene ✅

| Issue                   | Location | Recommendation                                                                                                                                                        |
| ----------------------- | -------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Missing direct deps** | `go.mod` | `golang.org/x/sync`, `google.golang.org/grpc`, and `google.golang.org/protobuf` are used directly; move them to `require` (non-indirect) or run `go mod tidy` to fix. |
| **Indirect flags**      | `go.mod` | Ensure `// indirect` is only on true transitive dependencies.                                                                                                         |

### 4. Consistency and Polish

| Issue                       | Location                 | Recommendation                                                                                                                                       |
| --------------------------- | ------------------------ | ---------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Liveness 405 response**   | `HealthHandler.Liveness` | For non-GET requests, return JSON (e.g. `ErrorResponse`) for consistency with `Ready` and other handlers.                                            |
| **Unused errgroup context** | `main.go:109`            | `g, ctx := errgroup.WithContext(...)` – `ctx` is unused. Either use it for propagation/cancellation or drop it: `g, _ := errgroup.WithContext(...)`. |

### 5. Testing✅

| Issue                 | Location                     | Recommendation                                                                                                                                  |
| --------------------- | ---------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------- |
| **LoggingMiddleware** | `internal/api/middleware.go` | Add `middleware_test.go` to cover logging behavior and status code recording.                                                                   |
| **Integration tests** | N/A                          | Consider a small integration test that starts the server and exercises HTTP and gRPC together (optional, but useful for deployment validation). |

### 6. Security and Validation

| Issue                            | Location        | Recommendation                                                                                          |
| -------------------------------- | --------------- | ------------------------------------------------------------------------------------------------------- |
| **Allowed countries validation** | `Checker.Check` | Consider validating country codes (e.g. ISO 3166-1 alpha-2) if the system must strictly enforce format. |
| **IP normalization**             | `Checker.Check` | IPv4-mapped IPv6 addresses could be normalized for consistency; document behavior if not.               |

---

## Minor Notes

- **mockLookuper:** Defined in both `api` and `geofence` test files. Reasonable given package boundaries; could be extracted to a shared `testutil` package if it grows.
- **GeoStore vs Checker IP validation:** `Checker` validates IP format before lookup; `GeoStore.Lookup` also checks for nil and invalid conversion. Slight overlap but defensible.
- **GeoStore in NewGeoStore:** Logs at Info level on success. Consider whether this should be Debug in production to reduce log volume.

---

## Suggested Priorities

1. **P0:** Fix gRPC exposure in Docker and K8s.✅
2. **P1:** Add request size limit and HTTP server timeouts.
3. **P2:** Fix go.mod dependencies and add LoggingMiddleware tests.✅
4. **P3:** Align Liveness 405 response and shutdown ordering.

---

## Closing

The project shows good understanding of Go idioms, clean architecture, and production concerns (health checks, structured logging, typed errors). The main gaps are around deployment configuration for gRPC and basic production hardening (limits, timeouts). Addressing these would bring it in line with production-grade standards.
