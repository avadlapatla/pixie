# Docker Build Optimization Guide for Go Applications

This document explains the optimization strategies implemented to reduce Docker build times for the Go application, specifically addressing the slow step:
```
RUN go mod tidy && CGO_ENABLED=0 GOOS=linux go build -o ../pixie-core .
```

## Table of Contents
1. [Implemented Optimizations](#implemented-optimizations)
2. [Production vs. Development Workflow](#production-vs-development-workflow)
3. [Build Performance Comparison](#build-performance-comparison)
4. [Usage Instructions](#usage-instructions)
5. [Additional Optimization Tips](#additional-optimization-tips)

## Implemented Optimizations

### Multi-stage Builds
- **Builder Stage**: Compiles the Go application using the golang image
- **Final Stage**: Uses minimal alpine image to run the application
- **Benefits**: Smaller final image size, better security posture

### Layer Caching Strategy
- **Dependency Management**: `go.mod` and `go.sum` files are copied and dependencies are downloaded before copying source code
- **Selective Copying**: Only necessary directories (`core/` and `proto/`) are copied into the build context
- **Benefits**: Faster rebuilds as dependencies layer is cached unless dependencies change

### Build Flags Optimization
- **Strip Debug Information**: `-ldflags="-w -s"` removes debug information and symbol tables
- **Trim Path**: `-trimpath` removes file system paths from the binary
- **Static Linking**: `CGO_ENABLED=0` ensures static linking
- **Architecture Targeting**: Explicit `GOARCH=amd64` for optimized builds
- **Benefits**: Smaller binary size, faster build times

### Development Mode with Hot Reload
- Implemented using `air` for automatic rebuilds when files change
- Mounts source code as volumes instead of copying
- **Benefits**: Instant feedback during development without full rebuilds

## Production vs. Development Workflow

### Production (Optimized for Size and Performance)
- Uses `deployments/docker/core.Dockerfile`
- Produces smaller, optimized binaries
- Full rebuild required for each code change
- Best for deployment and CI/CD pipelines

### Development (Optimized for Iteration Speed)
- Uses `deployments/docker/core.dev.Dockerfile`
- Implements hot reloading with `air`
- Mounts source code as volumes for immediate changes
- Best for local development and testing

## Build Performance Comparison

| Aspect                        | Before               | After (Production)     | After (Development)   |
|-------------------------------|----------------------|------------------------|----------------------|
| Docker Build Time             | Slow (minutes)       | Faster (layer caching) | Very fast (seconds)  |
| Rebuild After Code Changes    | Full rebuild needed  | Full rebuild needed    | Hot reload (instant) |
| Binary Size                   | Larger               | Smaller (stripped)     | Dev mode (not optimized) |
| Container Image Size          | Larger               | Smaller                | Larger (includes dev tools) |

## Usage Instructions

### Building and Running

We've provided a `Makefile.optimize` with simplified commands:

```bash
# Build optimized production image
make -f Makefile.optimize build-prod

# Run production containers
make -f Makefile.optimize run-prod

# Build development image with hot reload
make -f Makefile.optimize build-dev

# Run development environment with hot reload
make -f Makefile.optimize run-dev
```

### Development Workflow

1. Start the development environment:
   ```bash
   make -f Makefile.optimize run-dev
   ```
   
2. Edit files locally - changes are detected automatically and the app will rebuild and restart

3. Use the development tools in the container as needed

## Additional Optimization Tips

1. **Keep .dockerignore Updated**: Exclude unnecessary files from the Docker build context
   ```
   .git/
   .github/
   **/*.md
   **/README*
   **/LICENSE
   **/*_test.go
   **/test
   ```

2. **Dependency Updates**: Consider using tools like `go mod tidy -v` periodically to clean up unused dependencies

3. **Build Cache**: Use Docker BuildKit for even better build cache:
   ```bash
   DOCKER_BUILDKIT=1 docker build -t pixie-core .
   ```

4. **Go Build Cache**: Consider mounting Go build cache as a volume in development:
   ```yaml
   volumes:
     - go-build-cache:/root/.cache/go-build
   ```

5. **Parallelism**: For multi-module projects, consider parallel dependency downloads with `-x` flag for go mod download

6. **Vendor Dependencies**: In some cases, vendoring dependencies can speed up builds:
   ```bash
   go mod vendor
   ```
   Then modify the Dockerfile to use vendored dependencies

7. **Tune Go GC**: For large builds, tuning Go's garbage collector can help:
   ```
   ENV GOGC=off
