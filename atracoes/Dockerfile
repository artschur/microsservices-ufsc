# Stage 1: Build the Go binary
FROM golang:1.24.1-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum to download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -o /atracoes-service

# Stage 2: Create the final lightweight image
FROM alpine:latest

# Copy the binary from the builder stage
COPY --from=builder /atracoes-service /atracoes-service

# Expose the port the service runs on
EXPOSE 8081

# Command to run the service
CMD ["/atracoes-service"]
