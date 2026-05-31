.PHONY: proto deps build test clean

# Install gRPC dependencies and protoc plugins
deps:
	go get google.golang.org/grpc
	go get google.golang.org/protobuf
	go mod tidy

# Install protoc-gen-go and protoc-gen-go-grpc (one-time setup)
install-protoc-plugins:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Proto generation — regenerate Go code from signals.proto
proto:
	protoc --go_out=. --go_opt=paths=source_relative \
	       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
	       proto/signals.proto
	go mod tidy

# Build the Go API binary
build:
	go build -o bin/api ./cmd/main.go

# Run tests
test:
	go test ./... -v

# Clean build artifacts
clean:
	rm -rf bin/
