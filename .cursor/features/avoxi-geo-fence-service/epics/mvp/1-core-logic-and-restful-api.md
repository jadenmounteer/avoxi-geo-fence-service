# Core Logic & RESTful API

## Deliverable: A functional Go service capable of determining geo-fencing permissions.

### Task 1.1: Initialize Go Project Structure and Dependency Management ✅

#### Key Actions:

- Initialize the module: Use a naming convention that matches a remote repository (e.g., github.com/username/repo-name). (`go mod init`) ✅

- Create the Directory Tree:

  ` /cmd/server/`: The entry point of the application. ✅

  `/internal/`: For code you don't want other projects to import (the core logic). ✅

  `/data/`: To hold the MaxMind .mmdb file. ✅

  `/pkg/`: (Optional) for shared utility code. ✅

- Dependency Setup: Install the geoip2-golang library. ✅

- Ignore Files: Create .gitignore to keep binary files and local secrets out of Git, and .cursorignore to prevent Cursor/AI tools from indexing heavy binary data like the MaxMind database.

#### Acceptance Criteria:

- [x] go.mod and go.sum files exist and the module name is correct.

- [x] The project structure follows the /cmd/ and /internal/ pattern.

- [x] go mod verify passes without errors.

- [x] .gitignore includes entries for:
  - The compiled binary (e.g., avoxi-service)
  - The /data/\*.mmdb file (don't commit large binaries to Git!)
  - OS-specific files (e.g., .DS_Store)

- [x] .cursorignore is present and excludes:
  - data/ (to keep the AI from trying to "read" a binary database)
  - vendor/ (if present)

### Task 1.2: Integrate MaxMind GeoLite2 Reader and Binary Database ✅

You need to bridge the gap between your Go code and the binary .mmdb file. This involves establishing a "Store" or "Service" in Go that opens the database and provides a safe, clean interface for other parts of your app to query.

#### Key Actions:

- Procure the Database: Download the GeoLite2-Country.mmdb file from MaxMind (or use a placeholder if you're in a restricted environment) and place it in your /data folder. ✅

- Initialize the Reader: Write the logic to open the file using the oschwald/geoip2-golang library.

- Encapsulation: Create a struct (e.g., GeoStore) so you aren't passing the raw library objects around your code. This makes testing and future updates (like Epic 4's hot-reloading) much easier.

- Error Handling: Implement checks to ensure the service fails fast if the database file is missing or corrupted.

#### Acceptance Criteria (AC)

- [x] The project can successfully open a .mmdb file located in the /data directory at startup.

- [x] A Lookup method exists that accepts a net.IP and returns a standard ISO Country Code (e.g., "US", "FR").

- [x] The code handles "Unknown" IPs gracefully (e.g., if an IP isn't in the database, it returns a specific error or an empty string, not a crash).

- [x] The database reader is properly closed using defer or a dedicated Close() method to prevent memory leaks.

- [x] Self-Documentation: A small README.md entry exists in the /data folder explaining where to get the MaxMind file.

### Task 1.3: Implement Geo-Fencing Business Logic (IP-to-Country Validation) ✅

You need to create the logic that ties the IP lookup to the "White List" of countries provided in the request. This should be a clean, reusable function or method that doesn't care about HTTP or gRPC; it just cares about the "Yes/No" decision.

#### Key Actions:

- Input Parsing: Ensure the logic can handle different IP formats (IPv4 vs IPv6).

- Comparison Logic: Implement a high-performance check to see if the detected country exists in the allowed_countries slice.

- Edge Case Handling: Decide what happens if the IP is internal (private ranges like 192.168.x.x), if the country code is empty, or if the "allowed" list is empty.

- Result Object: Return a structured result that includes the decision (bool) and the metadata (the actual country found), which helps with logging later.

- Tip: Efficiency matters in Go. Instead of looping through the allowed_countries slice for every request, a better approach for a large list of countries would be to convert the slice into a map[string]struct{} for $O(1)$ lookups. However, since the list of allowed countries for a single customer is likely small (usually 1–5), a simple loop is often faster due to CPU cache locality.

#### Acceptance Criteria (AC)

- [x] Core Logic: A function exists that takes a string (IP) and a []string (Allowed Countries) and returns a bool.

- [x] Accuracy: \* If IP=8.8.8.8 and Allowed=["US", "CA"], result is true.

- If IP=8.8.8.8 and Allowed=["GB"], result is false.

- [x] Case Sensitivity: The logic should be resilient to casing (e.g., "us" vs "US") if possible, or strictly documented as ISO-2 uppercase.

- [x] Invalid Input: The function returns a clear error if the IP string provided is malformed (e.g., "not-an-ip").

- [x] Unit Tested: At least 3 unit tests exist covering: a successful match, a blocked match, and an invalid IP.

### Task 1.4: Develop HTTP/JSON Handler and API Routing ✅

You will create the web server layer. This layer is responsible for "unmarshalling" (parsing) the incoming JSON, calling your GeoStore logic, and "marshalling" (encoding) the result back into a JSON response.

#### Key Actions:

- Define DTOs (Data Transfer Objects): Create Go structs with JSON tags that match the expected request/response format.

- Routing: Set up the URL path (e.g., POST /v1/check).

- Handler Logic: \* Validate the HTTP method (only allow POST).

- Parse the request body.

- Call the logic from Task 1.3.

- Write the appropriate HTTP status codes (200 for success, 400 for bad input, 500 for server errors).

- Middleware (Optional but recommended): Add a basic logger to print the incoming requests to the console.

#### Acceptance Criteria (AC)

- [x] Endpoint exists: The server listens on a configurable port (defaulting to :8080) and responds to POST /v1/check.

- [x] Request Validation: Returns 405 Method Not Allowed for GET requests.

- [x] JSON Schema: Accurately maps ip_address and allowed_countries from the JSON body.

- [x] Error Handling: Returns a clear JSON error message and 400 Bad Request if the JSON is malformed or the IP is invalid.

- [x] Content-Type: Correctly sets Content-Type: application/json in the response header.

### Task 1.5: Implement Graceful Shutdown and Resource Cleanup

You need to instruct your Go application to listen for "Termination Signals" from the Operating System (like SIGINT or SIGTERM, which Kubernetes sends when it wants to stop a container). Instead of exiting instantly, the app should stop accepting new requests, finish the ones it's currently processing, and then close the database reader cleanly.

#### Key Actions:

- Signal Channel: Create a channel to listen for os.Interrupt and syscall.SIGTERM.

- Context Timeout: Use context.WithTimeout to give the server a window (e.g., 5–10 seconds) to finish active work.

- Cleanup Logic: Explicitly call .Close() on the MaxMind reader to ensure memory-mapped files are released properly.

- Shutdown Log: Add a log entry so you can see the service shutting down gracefully in your terminal.

#### Acceptance Criteria (AC)

- [ ] Signal Detection: The application does not exit immediately when Ctrl+C is pressed.

- [ ] Server Shutdown: The http.Server.Shutdown() method is called correctly using a background context.

- [ ] Resource Release: The MaxMind database reader is closed after the server stops accepting new requests.

- [ ] Clean Exit: The program exits with status code 0 after a successful graceful shutdown.
