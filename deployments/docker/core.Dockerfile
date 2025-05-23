# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY core/go.mod ./

# Download dependencies
RUN go mod download

# Copy source code
COPY core/ ./

# Build the application with static linking
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-s -w" -o pixie-core .

# Final stage
FROM scratch

# Copy the binary from builder
COPY --from=builder /app/pixie-core /pixie-core

# Copy CA certificates for HTTPS requests
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Expose the application port
EXPOSE 8080

# Set the entrypoint
ENTRYPOINT ["/pixie-core"]
