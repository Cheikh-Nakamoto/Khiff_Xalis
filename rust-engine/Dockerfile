FROM rust:1.77 AS builder
WORKDIR /app
RUN apt-get update && apt-get install -y protobuf-compiler && rm -rf /var/lib/apt/lists/*
COPY Cargo.toml Cargo.lock ./
COPY build.rs ./
COPY proto/ ../proto/
COPY src/ src/
COPY benches/ benches/
RUN cargo build --release

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y ca-certificates wget && \
    wget -qO/bin/grpc_health_probe https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/v0.4.24/grpc_health_probe-linux-amd64 && \
    chmod +x /bin/grpc_health_probe && \
    rm -rf /var/lib/apt/lists/*
COPY --from=builder /app/target/release/brvm-engine /usr/local/bin/brvm-engine
EXPOSE 50051
CMD ["brvm-engine"]
