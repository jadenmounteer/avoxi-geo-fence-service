# Stage 1: Builder
FROM golang:1.25-alpine AS builder
WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o avoxi-geo-fence ./cmd/server

# Stage 2: Runner
FROM alpine:latest
RUN adduser -D -g '' appuser

WORKDIR /app
COPY --from=builder /build/avoxi-geo-fence .
COPY --from=builder /build/data ./data

USER appuser

EXPOSE 8080 9090
CMD ["./avoxi-geo-fence"]
