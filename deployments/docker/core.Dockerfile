FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

# Set the working directory
WORKDIR /app

# Copy go.mod and go.sum files first (better layer caching)
COPY go.mod go.sum ./
COPY core/go.mod core/go.sum ./core/

# Download dependencies (this layer will be cached unless dependencies change)
RUN go mod download
RUN cd core && go mod download

# Copy only necessary files for the build
COPY core/ ./core/
COPY proto/ ./proto/

# Build with optimizations
WORKDIR /app/core
RUN go mod tidy && \
    GOPROXY=direct \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    go build -ldflags="-w -s" \
    -trimpath \
    -a \
    -o ../pixie-core .

# Create a minimal image
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Set the working directory
WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/pixie-core .

# Copy the UI React plugin - this should happen in a separate step
COPY plugins/ui-react/dist /plugins/ui-react/dist

# Create directories for plugins and keys
RUN mkdir -p /plugins /app/keys
COPY keys/ /app/keys/

# Expose the port
EXPOSE 8080

# Set environment variables
ENV JWT_ALGO=HS256
ENV JWT_SECRET=supersecret123

# Run the application
CMD ["/app/pixie-core"]
