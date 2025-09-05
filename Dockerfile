# Stage 1: Build the Go binary
FROM golang:1.24-alpine AS builder

# Install git for dependency fetching if needed
RUN apk add --no-cache git

# Set working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum first to cache dependencies
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy all source code
COPY . .

# Build the binary for Linux
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o image-resizer ./cmd/main.go

# Stage 2: Create a minimal image to run the binary
FROM alpine:latest

# Install ca-certificates for HTTPS support
RUN apk add --no-cache ca-certificates

# Copy the binary from builder
COPY --from=builder /app/image-resizer /usr/local/bin/image-resizer

# Set working directory
WORKDIR /usr/local/bin

# Expose port 8080
EXPOSE 8080

# Run the binary
CMD ["./image-resizer"]
