# Production Grade Containerization

## Deliverable: Small, secure Docker image and K8s configuration files.

### Task 2.1: Design Multi-Stage Dockerfile for Minimal Image Footprint

You need to create a Dockerfile that separates the build environment from the production environment.

#### Key Actions:

- Stage 1 (The Builder): Use a heavy Go image (like golang:1.24-alpine) to compile your code. This stage includes the Go compiler, git, and your source code.

- The Compilation: Build a static binary. This is crucial for Go—it ensures the binary contains everything it needs to run without external dependencies.

- Stage 2 (The Runner): Start fresh with a tiny, secure image (like alpine:latest or scratch).

- The Transfer: Copy only the compiled binary and the /data folder (containing your .mmdb file) from the Builder to the Runner.

- Security: Ensure the application does not run as the "root" user inside the container.

#### Acceptance Criteria (AC)

- [ ] Multi-Stage Structure: The Dockerfile contains at least two FROM statements.

- [ ] Minimal Footprint: The final image size is under 50MB (excluding the .mmdb file size).

- [ ] Static Compilation: The Go build command includes flags to disable cgo (CGO_ENABLED=0) for maximum portability.

- [ ] Non-Root User: The Dockerfile creates a system user and uses the USER instruction to run the binary.

- [ ] Cleanliness: The image does not contain source code, Go toolchains, or shell history.

### Task 2.2: Configure Structured Logging for Observability

You will replace standard text logging with Go's native slog (introduced in Go 1.21). This allows you to output logs as JSON objects while maintaining high performance.

#### Key Actions:

- Initialize slog: Set up a global or injected logger that outputs to os.Stdout in JSON format.

- Contextual Metadata: Ensure logs include helpful keys like "service": "geo-fence-service" and "version": "1.0.0".

- HTTP Middleware: Implement (or use a small utility) to log every incoming request, including the duration it took to process and the resulting status code.

- Error Leveling: Use appropriate levels (Info, Warn, Error). For example, a blocked IP might be an Info level, but a missing database file is a Fatal/Error.

#### Acceptance Criteria (AC)

- [ ] JSON Output: All application logs are printed in valid JSON format.

- [ ] Standard Fields: Every log entry includes a time, level, and msg.

- [ ] Request Logging: Every API call produces a log line containing:
  - method (e.g., POST)

  - path (e.g., /v1/check)

  - status (e.g., 200)

  - duration_ms (how long the lookup took)

- [ ] No Secrets: Ensure logs do not leak sensitive info (though for this project, IP addresses are the primary data).

- [ ] Standard Library Only: Accomplished using log/slog to keep the binary lean (following your modular rules).

### Task 2.3: Implement Environment Variable Configuration for Data Paths and Ports

You will modify your application to look for environment variables at startup. If they aren't found, the app should "fail-safe" by using sensible defaults.

#### Key Actions:

- Identify Variables: Define keys for APP_PORT (default: 8080), DB_PATH (default: ./data/GeoLite2-Country.mmdb), and LOG_LEVEL (default: info).

- Fetcher Logic: Create a helper function to read these variables. Avoid heavy third-party libraries like Viper if you can—Go's os.Getenv is more than enough for a service of this scale and keeps the binary small.

- Initialization: Call these at the very top of main.go before any other service (like the GeoStore) starts.

#### Acceptance Criteria (AC)

- [ ] No Hardcoding: The port and database path are not hardcoded strings in the logic.

- [ ] Sensible Defaults: If I run the app without setting any variables, it still works (using local defaults).

- [ ] Validation: The app logs an error and exits (Fail-Fast) if a provided DB_PATH does not exist.

- [ ] Documentation: The README includes a section listing all available environment variables and their purposes.

### Task 2.4: Define Kubernetes Deployment and Service Manifests

You need to create two primary YAML files (usually combined into one or kept in a /k8s folder). These files serve as the blueprint for your service's life in the cloud.
Note: I have installed Kind (`brew install kind`) and kubectl (`kubectl`) in order to verify the manifests and probes.

#### Key Actions:

- Deployment Manifest: Define the Deployment object. This tells K8s which Docker image to use (from Task 2.1), how many replicas to run, and which environment variables to inject (from Task 2.3).

- Resource Constraints: Set limits and requests for CPU and Memory. This prevents your service from accidentally eating all the resources on a cluster node.

- Service Manifest: Define a Service object (usually type ClusterIP) to give your pods a stable internal IP address and a DNS name.

- Label Selectors: Ensure the Service "finds" the Deployment using correct app labels.

#### Acceptance Criteria (AC)

- [ ] Deployment Object: Successfully creates a Deployment that references your Docker image name.

- [ ] Service Object: Successfully creates a Service that maps port 80 (external) to your app port (e.g., 8080).

- [ ] Environment Injection: The manifest includes a env section that sets DB_PATH and APP_PORT.

- [ ] Resource Management: Includes resources: requests and limits (e.g., 100m CPU, 128Mi Memory).

- [ ] Clean Metadata: Uses standard labeling (e.g., app: geo-fence-service).

### Task 2.5: Define Kubernetes Readiness and Liveness Probes

You need to create a dedicated health-check mechanism within your Go application and then configure Kubernetes to monitor it.
Note: I have installed Kind (`brew install kind`) and kubectl (`kubectl`) in order to verify the manifests and probes.

#### Key Actions:

- Implement the Health Endpoint: In your Go API, add a GET /health or GET /ready route.

- Liveness: Should return a 200 OK as long as the web server is alive.

- Readiness: Should return 200 OK only if the MaxMind database is successfully loaded and ready for queries.

- K8s Configuration: Update your deployment.yaml to point to these endpoints.

- efine Thresholds: Set the initialDelaySeconds (how long to wait before checking) and periodSeconds (how often to check).

#### Acceptance Criteria (AC)

- [ ] Dedicated Endpoint: A GET /health route exists and returns JSON (e.g., {"status": "up"}).

- [ ] Liveness Probe: Added to the container spec; restarts the pod if the endpoint fails.

- [ ] Readiness Probe: Added to the container spec; prevents traffic from reaching the pod until the database is loaded.

- [ ] Non-Blocking: The health check logic is fast and doesn't put significant load on the CPU.

- [ ] Observability: If the health check fails, it should log the reason (e.g., "Database handle is nil") using your slog setup from Task 2.2.
