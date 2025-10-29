# Stage 1: Build
FROM golang:1.24.1-alpine AS builder
WORKDIR /app
COPY go.mod .
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /fila-service

# Stage 2: Final
FROM alpine:latest
WORKDIR /app
COPY --from=builder /fila-service ./fila-service
EXPOSE 8083
CMD ["./fila-service"]
