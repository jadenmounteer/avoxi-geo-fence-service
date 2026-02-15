# High-Performance gRPC Interface

Note to self: gRPC (Remote Procedure Calls) essentially allows you to send a super fast message to another API. That API will then respond, possibly saying if a task was successfully completed or not. This is not asyncronous. It is useful in the context of this feature because it's a lot faster than JSON (5x, actually). This allows other services to check the geofence with extreme speed.

## Deliverable: A dual-protocol server responding to both HTTP and gRPC. Supporting both allows it to be highly accessible and high-performant. Not everything can talk to gRCP, but for the microservices that can, it is soooooo fast.

### Task 3.1: Define Protocol Buffer (.proto) Service Contract✅

You will design the interface for your gRPC service. Unlike REST, where the "contract" is often loose documentation, gRPC requires a strict definition of your data structures and your service methods before you write any implementation code.

#### Key Actions:

- Define the Syntax: Use proto3 (the modern standard).

- Define Messages: Create the CheckRequest (containing the IP and allowed countries) and CheckResponse (containing the boolean result and country code).

- Define the Service: Create a GeoFenceService with a CheckAccess RPC method.

- Package Naming: Follow Go conventions for the option go_package to ensure the generated code lands in the right directory (e.g., internal/pb).

#### Acceptance Criteria (AC)

- [x] Syntax: The file begins with syntax = "proto3";.

- [x] Data Types: Uses appropriate Protobuf types (e.g., string for IP, repeated string for the list of countries).

- [x] Naming Conventions: Follows Google’s Style Guide (CamelCase for Messages, snake_case for field names).

- [x] Field Tags: Every field in a message has a unique incrementing tag (e.g., string ip_address = 1;).

- [x] Inclusion of Health Check: (Optional but Pro) Include a simple CheckHealth RPC to mirror your K8s probes in the gRPC world.

### Task 3.2: Generate Go Code Bindings via protoc✅

You need to set up the tooling to compile your Protobuf file. This requires the protoc compiler and the Go-specific plugins. We will create a Makefile or a Go Generate script so any developer can recreate the code with one command.

Note: I already ran the following commands, ensuring I can generate gRPC code:

```
brew install protobuf
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

#### Key Actions:

- Tooling Installation: Ensure protoc, protoc-gen-go, and protoc-gen-go-grpc are installed in your environment. ✅

- Output Directory: Create the target folder (e.g., internal/pb) where the generated files will live.

- Compilation Command: Run the compiler, pointing it to your .proto file and specifying the Go and gRPC plugins.

- Automation: Add the command to a Makefile or use //go:generate comments in your code.

#### Acceptance Criteria (AC)

- [x] Successful Generation: Two files are created in your internal/pb (or similar) directory:
  - geofence.pb.go (Contains data structures/messages).

  - geofence_grpc.pb.go (Contains the gRPC service client and server interfaces).

- [x] No Manual Edits: The generated files should never be edited by hand (verify they have the "DO NOT EDIT" header).

- [x] Automated Script: A Makefile or scripts/gen-proto.sh exists so the team can re-generate code easily if the .proto changes.

- [x] Importable: The generated code compiles without errors when imported into your main application.

- [x] Make file is in project root.

### Task 3.3: Implement gRPC Server Interface and Method Logic✅

Now that you have the generated "stubs" from Task 3.2, you need to actually write the Go logic that satisfies the gRPC interface. You aren't rewriting your business logic; you are simply "wrapping" the logic you wrote in Task 1.3 inside a gRPC handler.

#### Key Actions:

- Define the Server Struct: Create a struct that "embeds" the generated UnimplementedGeoFenceServiceServer.

- Implement the Method: Write the CheckAccess function to match the signature in your generated code.

- Bridge to Logic: Inside that function, call your internal/geofence validation logic.

- Error Mapping: Map Go errors to gRPC status codes (e.g., use codes.InvalidArgument if the IP is malformed).

#### Acceptance Criteria (AC)

- [x] Interface Satisfaction: The server struct correctly implements the GeoFenceServiceServer interface.

- [x] Request Handling: The CheckAccess method correctly extracts the IP and Allowed List from the Protobuf request object.

- [x] Response Construction: The method returns a CheckAccessResponse containing the boolean result and the country code.

- [x] Graceful Errors: If the lookup fails, the server returns a proper gRPC error using status.Error.

### Task 3.4: Configure Concurrent HTTP and gRPC Server Listeners✅

You need to launch both servers simultaneously without one blocking the other. In Go, we do this using Goroutines and a sync.WaitGroup (or an errgroup) to manage their lifecycle.

#### Key Actions:

- Initialize both Servers: Set up your HTTP router (from Epic 1) and your gRPC server (from Task 3.3).

- Run in Goroutines: Wrap each .Serve() call in a go func(). This allows them to run in the background.

- Unified Error Handling: Use golang.org/x/sync/errgroup to catch errors from either server. If one fails to start, the whole app should shut down gracefully.

- Shared Logic: Ensure both servers are injected with the same GeoStore instance so they share the same database and business rules.

#### Acceptance Criteria (AC)

- [x] Dual Connectivity: I can successfully curl the HTTP port AND run a gRPC client call against the gRPC port.

- [x] Non-Blocking: Neither server prevents the other from starting.

- [x] Clean Shutdown: When a termination signal is received, both servers stop gracefully (using the logic from Task 1.5).

- [x] Configuration: Ports for both HTTP and gRPC are configurable via environment variables (e.g., HTTP_PORT and GRPC_PORT).

### Task 3.5: Implement gRPC Reflection for Easier Service Discovery and Testing✅

You will register the reflection service on your gRPC server. This is a one-liner in Go, but it fundamentally changes how easy your service is to debug in a Kubernetes environment.

Note: Use grpcurl to test.
I have already ran `go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest`, so we should be ready to test.

#### Key Actions:

- Import the Reflection Package: Add google.golang.org/grpc/reflection to your project.

- Register the Service: Call reflection.Register(grpcServer) after you've registered your GeoFenceService.

- Validation: Use a CLI tool to "list" the services running on your server.

#### Acceptance Criteria (AC)

- [x] Reflection Enabled: The gRPC server has reflection registered.

- [x] Discovery: Running grpcurl -plaintext localhost:9090 list returns geofence.v1.GeoFenceService.

- [x] Method Inspection: Running grpcurl can describe the CheckAccess method and its request/response types.

- [x] No Proto Required: An external developer can call your service without having the physical .proto file.
