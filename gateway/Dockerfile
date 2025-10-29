# Stage 1: Build the Go binary
FROM golang:1.24.1-alpine AS builder

WORKDIR /app

COPY go.mod .
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /gateway-service

# Stage 2: Final image
FROM alpine:latest
WORKDIR /app
COPY --from=builder /gateway-service ./gateway-service
EXPOSE 8000
CMD ["./gateway-service"]
