# Proto — gRPC contract between Go API and Rust Engine

## Files

- `signals.proto` — the gRPC service definition
- `signals.pb.go` — generated message types (DO NOT EDIT)
- `signals_grpc.pb.go` — generated gRPC stubs (DO NOT EDIT)

## Quick start

From the `go-api/` root:

```bash
# 1. Add gRPC dependencies to go.mod
make deps

# 2. (Optional) Regenerate Go code from proto
make proto
```

## Regenerate Go code

Requires `protoc` installed on the system, plus the Go plugins:

```bash
make install-protoc-plugins  # one-time
make proto
```

## Contract summary

```
service SignalService {
  rpc GenerateSignal(SignalRequest) returns (SignalResponse);
  rpc ScanAllTickers(ScanRequest) returns (ScanResponse);
  rpc HealthCheck(HealthRequest) returns (HealthResponse);
}
```
