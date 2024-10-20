
FROM golang:1.23.0-alpine AS builder
WORKDIR /app
COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . ./
RUN go build -o /stop-loss-trading


FROM alpine:latest
WORKDIR /
COPY --from=builder /stop-loss-trading .
EXPOSE 8080 9090  # Expose both HTTP API and Prometheus metrics ports

CMD ["./stop-loss-trading"]
