BINARY := avoxi-geo-fence
IMAGE := avoxi-geo-fence:latest

.PHONY: build run test test-integration clean proto docker-build docker-run kind-cluster kind-load k8s-deploy k8s-up k8s-forward help

help:
	@echo "Local:"
	@echo "  make build           - Build the binary"
	@echo "  make run             - Build and run locally"
	@echo "  make test            - Run all tests"
	@echo "  make test-integration - Run integration tests (requires GeoIP DB)"
	@echo "  make proto           - Generate Go code from proto files"
	@echo ""
	@echo "Docker:"
	@echo "  make docker-build - Build Docker image"
	@echo "  make docker-run   - Run container (port 8080)"
	@echo ""
	@echo "Kubernetes (Kind):"
	@echo "  make kind-cluster - Create Kind cluster (one-time)"
	@echo "  make kind-load    - Build image and load into Kind"
	@echo "  make k8s-deploy   - Apply k8s manifests"
	@echo "  make k8s-up       - Load image + deploy (requires Kind cluster)"
	@echo "  make k8s-forward  - Port-forward service to localhost:8080"

build:
	go build -o $(BINARY) ./cmd/server

run: build
	./$(BINARY)

test:
	go test ./...

test-integration:
	go test -tags=integration ./cmd/server -v

proto:
	PATH="$$PATH:$$(go env GOPATH)/bin" protoc -I. \
		--go_out=. --go_opt=module=github.com/jadenmounteer/avoxi-geo-fence \
		--go-grpc_out=. --go-grpc_opt=module=github.com/jadenmounteer/avoxi-geo-fence \
		proto/geofence.proto
	@echo "Generated internal/pb/geofence.pb.go and internal/pb/geofence_grpc.pb.go"

clean:
	rm -f $(BINARY)

docker-build:
	docker build -t $(IMAGE) .

docker-run: docker-build
	docker run -p 8080:8080 $(IMAGE)

kind-cluster:
	kind create cluster

kind-load: docker-build
	kind load docker-image $(IMAGE)

k8s-deploy:
	kubectl apply -f k8s/

k8s-up: kind-load k8s-deploy

k8s-forward:
	kubectl port-forward svc/geo-fence-service 8080:80
