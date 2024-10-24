
FROM golang:1.23.0-alpine AS builder


WORKDIR /app


COPY go.mod ./
COPY go.sum ./
RUN go mod download


COPY . ./


RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o stop-loss-trading


FROM alpine:latest


RUN apk --no-cache add ca-certificates curl


WORKDIR /


COPY --from=builder /app/stop-loss-trading .


EXPOSE 8080 9090


CMD ["./stop-loss-trading"]
