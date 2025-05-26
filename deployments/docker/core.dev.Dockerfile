FROM golang:1.23-alpine

# Install dev dependencies
RUN apk add --no-cache git tzdata ca-certificates

# Set the working directory
WORKDIR /app

# Copy go.mod and go.sum files first (better layer caching)
COPY go.mod go.sum ./
COPY core/go.mod core/go.sum ./core/

# Download dependencies (will be cached)
RUN go mod download
RUN cd core && go mod download 

# Set environment for development
ENV CGO_ENABLED=0
ENV GO111MODULE=on
ENV GOOS=linux 
ENV GOARCH=amd64

# Install development tools
RUN go install github.com/cosmtrek/air@latest

# Working directory for the application
WORKDIR /app

# We'll mount the source code as a volume during development
# The .air.toml file should be in the project root
COPY .air.toml ./

# Create required directories
RUN mkdir -p /plugins /app/keys

# Default command - use air for hot reload
CMD ["air", "-c", ".air.toml"]
