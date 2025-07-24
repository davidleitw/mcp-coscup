# Use golang alpine image for building
FROM golang:1.23-alpine AS builder

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/server

# Use a minimal alpine image for runtime
FROM alpine:latest

# Install ca-certificates for TLS
RUN apk --no-cache add ca-certificates

# Set working directory
WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/main .

# Copy data directory
COPY --from=builder /app/data ./data

# Expose port (Cloud Run will set $PORT)
EXPOSE 8080

# Run the binary
CMD ["./main"]