#### AVOXI Geo-Fence Service

A high-performance microservice designed to determine if a specific IP address is allowed access based on a country whitelist. Built with Go 1.25+ and the MaxMind GeoLite2 database.

#### Prerequisites

- **Go:** 1.25 or higher (for local development)
- **MaxMind database:** Place `GeoLite2-Country.mmdb` in the `data/` directory. See [data/README.md](data/README.md) for download instructions.

#### Quick Start in isolation (quick tests on your local machine).

```bash
make run          # Build and run locally
make docker-run   # Build and run in Docker
```

For Kubernetes (Kind), create a cluster once, then deploy (good for production environments or microservices):

```bash
make kind-cluster   # One-time: create Kind cluster
make k8s-up         # Build, load image, deploy
make k8s-forward    # Port-forward to localhost:8080
```

Run `make help` for all available commands.

#### Environment Variables

| Variable  | Default                      | Description                             |
| --------- | ---------------------------- | --------------------------------------- |
| APP_PORT  | 8080                         | Port the server listens on              |
| PORT      | (fallback if APP_PORT unset) | Alternative for Heroku, Cloud Run, etc. |
| DB_PATH   | data/GeoLite2-Country.mmdb   | Path to GeoLite2-Country.mmdb           |
| LOG_LEVEL | info                         | Log level: debug, info, warn, error     |

#### Test the Endpoint

```bash
# Good request
curl -X POST http://localhost:8080/v1/check \
  -H "Content-Type: application/json" \
  -d '{"ip_address": "8.8.8.8", "allowed_countries": ["US", "CA"]}'

# Bad request (invalid IP) - returns 400
curl -X POST http://localhost:8080/v1/check \
  -H "Content-Type: application/json" \
  -d '{"ip_address": "not-an-ip", "allowed_countries": ["US"]}'

# Bad request (empty country list) - returns 400
curl -X POST http://localhost:8080/v1/check \
  -H "Content-Type: application/json" \
  -d '{"ip_address": "8.8.8.8", "allowed_countries": []}'
```
