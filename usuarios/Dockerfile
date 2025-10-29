# Stage 1: Build
FROM golang:1.24.1-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /usuarios-service

# Stage 2: Final
FROM alpine:latest
WORKDIR /app
COPY --from=builder /usuarios-service ./usuarios-service
EXPOSE 8082
CMD ["./usuarios-service"]
