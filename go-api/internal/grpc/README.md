# gRPC Client — Rust Engine

This package wraps the gRPC client for calling the Rust Engine.

## Usage

```go
import brvmgrpc "github.com/brvm/go-api/internal/grpc"

// Connect to the Rust Engine
client, err := brvmgrpc.NewEngineClient("localhost:50051")
if err != nil {
    log.Fatal(err)
}
defer client.Close()

// Generate a signal for a single ticker
resp, err := client.GenerateSignal(ctx, "SNTS", prices, volumes, 2500.0, nil)

// Scan all tickers
signals, err := client.ScanAllTickers(ctx, tickerDataList)
```

## Code Generation

The generated Go code lives in `proto/gen/`. To regenerate:

```bash
make proto
```

Or from the `go-api/` directory:

```bash
protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       proto/signals.proto
```

## Dependencies

Add the required gRPC dependencies to `go.mod`:

```bash
go get google.golang.org/grpc
go get google.golang.org/protobuf
```
