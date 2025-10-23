# Stage 1: Build the Go binary
FROM golang:1.24.1-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum to download dependencies
COPY go.mod .
RUN go mod download

# Copy the source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -o /fila-service

# Stage 2: Create the final lightweight image
FROM alpine:latest

# Copy the binary from the builder stage
COPY --from=builder /fila-service /fila-service

# Expose the port the service runs on
EXPOSE 8083

# Command to run the service
CMD ["/fila-service"]
