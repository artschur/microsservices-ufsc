# Stage 1: Build
FROM golang:1.24.1-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /atracoes-service

# Stage 2: Final
FROM alpine:latest
WORKDIR /app
COPY --from=builder /atracoes-service ./atracoes-service
EXPOSE 8081
CMD ["./atracoes-service"]
