FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go get github.com/prometheus/client_golang/prometheus/promhttp && go mod tidy
RUN CGO_ENABLED=0 go build -o /bin/collector cmd/main.go

FROM alpine:3.19
RUN apk --no-cache add ca-certificates tzdata
COPY --from=builder /bin/collector /bin/collector
ENV TZ=Africa/Abidjan
CMD ["/bin/collector"]
