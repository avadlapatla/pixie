FROM golang:1.22-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum
COPY go.mod ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application with CGO disabled for a static binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o plugin-thumbnailer .

# Use a minimal alpine image for the final stage
FROM alpine:3.19

# Install runtime dependencies (if any)
RUN apk add --no-cache ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/plugin-thumbnailer .

# Expose the port the service will run on
EXPOSE 8080

# Run the service
ENTRYPOINT ["/app/plugin-thumbnailer"]
