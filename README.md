#### AVOXI Geo-Fence Service

A high-performance microservice designed to determine if a specific IP address is allowed access based on a country whitelist. Built with Go 1.25+ and the MaxMind GeoLite2 database.

#### Prerequisites

- **Go:** 1.25 or higher (for local development)
- **MaxMind database:** Ensure `GeoLite2-Country.mmdb` is placed in the `data/` directory. See [data/README.md](data/README.md) for download instructions.

#### Running with Docker

1. Place `GeoLite2-Country.mmdb` in the `data/` directory.
2. Build and run:

```bash
docker build -t avoxi-geo-fence .
docker run -p 8080:8080 avoxi-geo-fence
```

To use a different host port:

```bash
docker run -p 3000:8080 -e PORT=8080 avoxi-geo-fence
```

#### Running Locally

1. Initialize the project (if needed):

```bash
go mod tidy
```

2. Build and run:

```bash
go build -o avoxi-geo-fence ./cmd/server
./avoxi-geo-fence
```

#### Configure Environment

- **PORT:** The port the server listens on (default: 8080).

#### Test the Endpoint

```bash
curl -X POST http://localhost:8080/v1/check \
  -H "Content-Type: application/json" \
  -d '{"ip_address": "8.8.8.8", "allowed_countries": ["US", "CA"]}'
```
