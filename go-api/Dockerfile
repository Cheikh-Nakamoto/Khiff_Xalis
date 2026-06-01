FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go get github.com/ansrivas/fiberprometheus/v2 && go mod tidy
RUN CGO_ENABLED=0 go build -o /bin/api cmd/main.go

FROM alpine:3.19
RUN apk --no-cache add ca-certificates
COPY --from=builder /bin/api /bin/api
EXPOSE 8080
CMD ["/bin/api"]
