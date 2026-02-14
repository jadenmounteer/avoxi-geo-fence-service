# Dynamic data lyfecycle and resiliency

The database file changes every 2 weeks. After this epic, a CRON job (not included in this epic) would be able to update the .db file every 24 hours
and it wouldn't reboot the server. Users wouldn't know anything changed.

## Deliverable: A self-healing system that stays up-to-date without restarts.

### Task 4.1: Implement Thread-Safe In-Memory Data Swapping using sync.RWMutex

You will wrap your GeoStore or database reader in a protective layer using sync.RWMutex. This allows hundreds of goroutines to read the data simultaneously, but ensures that if a write (swap) happens, it waits for the reads to finish and then updates the pointer safely.

#### Key Actions:

- Introduce a Mutex: Add sync.RWMutex to your GeoStore struct.

- Thread-Safe Lookups: Wrap your IP lookup logic in a RLock() (Read Lock).

- The Reload Method: Create a Reload(newPath string) function that:

- Opens the new .mmdb file.

- Acquires a Lock() (Write Lock).

- Swaps the old pointer with the new one.

- Closes the old reader.

#### Acceptance Criteria (AC)

- [ ] No Race Conditions: Running go test -race passes while simulated traffic is hitting the service.

- [ ] Read/Write Separation: The implementation uses RLock/RUnlock for lookups and Lock/Unlock for the reload.

- [ ] Resource Cleanup: The previous database file handle is closed properly after the swap to prevent memory leaks.

- [ ] Atomic Swap: The pointer swap happens instantly, ensuring no request ever sees a nil database.

### Task 4.2: Integrate File System Watcher (fsnotify) for Automated Database Reloads

You will use the fsnotify library to monitor the database file for WRITE or CREATE events. When an event occurs, you'll trigger the Reload() method.

#### Key Actions:

- Initialize Watcher: Create a new fsnotify.Watcher and point it to the directory containing your MaxMind database.

- Event Loop: Run a background goroutine that listens to the watcher.Events channel.

- Debouncing (Optional but Pro): Sometimes a file write takes a few milliseconds. A senior approach is to wait a brief moment after the first event to ensure the file is fully written before attempting to open it.

- Integration: Link the watcher event directly to your GeoStore.Reload() function.

- Error Handling: Listen to the watcher.Errors channel and log issues using slog.

#### Acceptance Criteria (AC)

- [ ] Automatic Detection: Replacing the .mmdb file while the server is running triggers a "Database Reloaded" log entry.

- [ ] Non-Blocking: The watcher runs in its own goroutine and does not interfere with HTTP/gRPC traffic.

- [ ] Resilience: The application does not crash if the watcher encounters a temporary OS error.

- [ ] Clean Exit: The watcher is closed during the graceful shutdown phase (from Task 1.5).

### Task 4.3: Design Kubernetes Sidecar Pattern for Automated MaxMind Updates

You will modify your Kubernetes deployment.yaml to include a second container. This container's sole job is to check for a new .mmdb file from MaxMind (or a private S3 bucket/URL) and save it to a shared volume.

#### Key Actions:

- Shared Volume: Define an emptyDir volume in the Pod spec.

- Volume Mounts: Mount this volume to both the main container and the sidecar container at the same path (e.g., /data).

- Sidecar Image: Use a lightweight image (like curlimages/curl or a custom script) that runs on a schedule to download the latest database.

- InitContainer (Optional but Recommended): Use an initContainer to ensure a version of the database exists before the Go app even tries to start.

#### Acceptance Criteria (AC)

- [ ] Multi-Container Pod: The Deployment manifest now defines two containers: geo-service and db-updater.

- [ ] Volume Sharing: Both containers reference the same volumeMount.

- [ ] Decoupled Logic: The main Go application code remains unaware of how the file is updated; it just reacts to the file change.

- [ ] Security: The sidecar is configured with the necessary environment variables (like MAXMIND_LICENSE_KEY) stored as K8s Secrets.

### Task 4.4: Implement Metrics Instrumentation (Prometheus) for Lookup Latency and Success Rates

You will integrate the Prometheus Go client library to track the RED (Rate, Errors, Duration) metrics for your lookups. You will also expose a dedicated /metrics endpoint that a Prometheus server can "scrape" to collect this data.

#### Key Actions:

- Define Metrics: Create global (or injected) Prometheus collectors:

- Counter: geofence_lookups_total (labeled by status and protocol).

- Histogram: geofence_lookup_duration_seconds (to track latency distribution).

- Instrument Handlers: Add logic to your HTTP and gRPC handlers to record the start time, run the lookup, and then Observe the duration and Inc the counter.

- Expose Endpoint: Mount the promhttp.Handler() on a new route (usually /metrics).

- Labeling: Use labels like result="allowed" or result="denied" so you can filter your data later in Grafana.

#### Acceptance Criteria (AC)

- [ ] New Endpoint: GET /metrics returns standard Prometheus text format (e.g., geofence_lookups_total 42).

- [ ] Latency Tracking: The histogram successfully captures durations and places them into buckets (e.g., 0.01s, 0.05s).

- [ ] Success/Failure Counters: Metrics differentiate between successful lookups and errors (like malformed IPs).

- [ ] Shared Registry: Both HTTP and gRPC servers contribute to the same global metrics registry.

- [ ] Performance: Instrumentation adds negligible overhead (sub-millisecond) to the request path.

### Task 4.5: Produce "Future-Proofing" Maintenance Documentation

You will create a MAINTENANCE.md file (or a significant section in your README.md) that acts as a runbook for the service. It should focus on operational tasks and architectural guardrails.

#### Key Actions:

- Database Update Procedure: Document how to obtain a new MaxMind license key and where to drop the .mmdb file for the sidecar (Task 4.3) to pick up.

- Scaling Guide: Explain the resource limits (Task 2.4). If traffic triples, should they increase CPU or Memory first? (Hint: Since it's an in-memory DB, Memory is the bottleneck).

- Extending the API: Provide a "Quick Start" for adding a new gRPC method, referencing the Makefile (Task 3.2) and the internal package structure.

- Troubleshooting RED Metrics: List the Prometheus alerts (Task 4.4) an engineer should look at if the service starts slowing down.

#### Acceptance Criteria (AC)

- [ ] Clarity: A junior developer can follow the guide to update the database without asking for help.

- [ ] Dependency List: Clearly lists required tools (protoc, docker, kubectl, make).

- [ ] Architectural Decision Records (ADR): Briefly explains why you chose slog over zap or logrus (standard library focus).

- [ ] Environment Variable Map: A complete table of all ENV variables and their effects.
