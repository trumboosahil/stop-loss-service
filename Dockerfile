# Build Stage
FROM golang:1.23.0-alpine AS builder

# Set the working directory
WORKDIR /app

# Copy go.mod and go.sum files and download dependencies
COPY go.mod ./
COPY go.sum ./
RUN go mod download

# Copy the source code
COPY . ./

# Build the Go application
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o stop-loss-trading

# Run Stage
FROM alpine:latest

# Install CA certificates to allow HTTPS connections and curl for testing
RUN apk --no-cache add ca-certificates curl

# Set the working directory
WORKDIR /

# Copy the compiled binary from the build stage
COPY --from=builder /app/stop-loss-trading .

# Expose ports for the HTTP server and Prometheus metrics server
EXPOSE 8080 9090

# Run the application
CMD ["./stop-loss-trading"]
