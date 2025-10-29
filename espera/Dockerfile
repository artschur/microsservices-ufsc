# Stage 1: Build
FROM golang:1.24.1-alpine AS builder
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /espera-service

# Stage 2: Final
FROM alpine:latest
WORKDIR /app
COPY --from=builder /espera-service ./espera-service
EXPOSE 8084
CMD ["./espera-service"]
