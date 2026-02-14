#### AVOXI Geo-Fence Service

A high-performance microservice designed to determine if a specific IP address is allowed access based on a country whitelist. Built with Go 1.21+ and the MaxMind GeoLite2 database.

#### 🛠 Prerequisites

Go: 1.21 or higher

MaxMind Database: Ensure GeoLite2-Country.mmdb is placed in the /data directory.

#### Getting Started

1. Intitialize the project.

```
go mod init github.com/yourusername/avoxi-geo-fence
go mod tidy
```

#### Configure Environment

The server uses environment variables for configuration. You can export them in your terminal or use a .env file:

APP_PORT: The port the server will listen on (default: 8080).

DB_PATH: The relative path to your MaxMind database (default: ./data/GeoLite2-Country.mmdb).

#### Run the server

`go run cmd/server/main.go`

#### Test the endpoint

````curl -X POST http://localhost:8080/check \
-H "Content-Type: application/json" \
-d '{
  "ip_address": "8.8.8.8",
  "allowed_countries": ["US", "CA"]
}'```
````
