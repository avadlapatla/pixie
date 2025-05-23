# Pixie - Lightweight Plugin-based Photo Hosting Service

Pixie is a lightweight, plugin-based photo hosting service built with Go, MinIO (S3), PostgreSQL (with PostGIS), and NATS JetStream.

## Quick Start

### Prerequisites

- [Docker](https://docs.docker.com/get-docker/) and Docker Compose
- [Go 1.22+](https://golang.org/dl/) (for local development)
- [Make](https://www.gnu.org/software/make/) (usually pre-installed on macOS/Linux)

### Installation

The easiest way to get started is to run the installation script:

```bash
./scripts/install.sh
```

Alternatively, you can start the services manually:

```bash
make dev
```

### Accessing Services

Once the services are running, you can access them at:

- **Core API**: [http://localhost:8080/healthz](http://localhost:8080/healthz)
- **MinIO Console**: [http://localhost:9001](http://localhost:9001)
- **PostgreSQL**: localhost:5432
- **NATS**: localhost:4222 (client), localhost:8222 (monitoring)

### API Usage Examples

#### Upload a photo

```bash
# Upload a photo and get back the ID
curl -X POST -F "file=@/path/to/your/photo.jpg" http://localhost:8080/upload
# Response: {"id":"123e4567-e89b-12d3-a456-426614174000"}
```

#### Download a photo

```bash
# Download a photo by ID
curl -o downloaded_photo.jpg http://localhost:8080/photo/123e4567-e89b-12d3-a456-426614174000
```

#### Delete a photo

```bash
# Delete a photo by ID
curl -X DELETE http://localhost:8080/photo/123e4567-e89b-12d3-a456-426614174000
```

#### Watch events

```bash
# Watch events in another terminal
docker compose exec nats nats sub "photo.*"
```

### Default Credentials

#### MinIO
- Username: `minioadmin`
- Password: `miniopass`

#### PostgreSQL
- Username: `pixie`
- Password: `pixiepass`
- Database: `pixiedb`

#### NATS
- No authentication by default in development mode

### Stopping Services

To stop and remove all containers:

```bash
make down
```

## Development

### Linting

To run linters on the Go code:

```bash
make lint
```

## Project Structure

```
pixie/
  core/             # Core service (Go)
  plugins/          # Plugin directory
    noop/           # Example "noop" plugin
  web/              # Web UI directory (empty for now)
  deployments/      # Deployment configurations
    docker-compose.yml
    docker/
      core.Dockerfile
  proto/            # Protocol buffer definitions
    plugin/         # Plugin API definitions
  scripts/          # Utility scripts
  Makefile          # Build automation
```

## Writing a Plugin

Pixie supports dynamic loading of plugins to extend its functionality. Here's how to create a new plugin:

1. **Copy the proto file**
   
   The plugin API is defined in `proto/plugin/v1/plugin.proto`. Copy this file to your plugin project.

2. **Generate stubs**
   
   Use `buf generate` to generate Go stubs from the proto file.

3. **Implement the server**
   
   Implement the `PhotoPlugin` service defined in the proto file. Your plugin must:
   - Accept a `--port=0` flag to let the OS choose an available port
   - Print `PORT=<n>` to stdout after starting
   - Implement the gRPC Health Check service
   - Implement the `PhotoPlugin` service methods

4. **Build the binary**
   
   Build your plugin and place the binary in the `plugins/` directory. The binary must be executable.

5. **Run Pixie**
   
   Run `make dev` to start Pixie with your plugin loaded.

## Authentication

Pixie supports JWT authentication via the auth-jwt plugin. The plugin validates JWT tokens using either HS256 or RS256 algorithms.

```bash
# Get a test token
TOKEN=$(docker run --rm alpine sh -c 'apk add --no-cache jq openssl > /dev/null &&   H=$(printf %s "$JWT_SECRET" | xxd -p -c 256);   HEADER=$(echo -n "{"alg":"HS256","typ":"JWT"}" | base64 | tr -d = | tr +/ -_);   PAYLOAD=$(echo -n "{"sub":"demo","exp":$(( $(date +%s) + 3600 ))}" | base64 | tr -d = | tr +/ -_);   SIGN=$(printf "%s.%s" "$HEADER" "$PAYLOAD" | openssl dgst -sha256 -mac HMAC -macopt hexkey:$H -binary | base64 | tr -d = | tr +/ -_);   echo "$HEADER.$PAYLOAD.$SIGN"')
curl -X POST -H "Authorization: Bearer $TOKEN" -F file=@example.jpg http://localhost:8080/upload
```

### JWT Configuration

Configure the JWT authentication plugin using the following environment variables:

- `JWT_ALGO`: JWT algorithm to use, either `HS256` or `RS256` (default: `HS256`)
- `JWT_SECRET`: Secret key for HS256 algorithm (required when using HS256)
- `JWT_PUBLIC_KEY_FILE`: Path to the public key file for RS256 algorithm (required when using RS256)

### Testing JWT Authentication

Several test scripts are provided to verify the JWT authentication plugin is working correctly:

```bash
# Run all tests
./run_all_tests.sh

# Or run individual tests
./test_jwt_auth.sh       # Test valid JWT token
./test_invalid_jwt.sh    # Test invalid JWT token
./test_user_id_header.sh # Test X-User-Id header
```

For more detailed information about testing the JWT authentication, see [JWT_TESTING.md](JWT_TESTING.md).

## Environment Variables

All secrets are configured via environment variables with sensible defaults for development:

- `S3_ENDPOINT`: MinIO/S3 endpoint URL (default: http://minio:9000)
- `S3_ACCESS_KEY`: MinIO/S3 access key (default: minio)
- `S3_SECRET_KEY`: MinIO/S3 secret key (default: minio123)
- `S3_BUCKET`: MinIO/S3 bucket name (default: pixie)
- `DATABASE_URL`: PostgreSQL connection string (default: postgres://pixie:pixiepass@postgres:5432/pixiedb?sslmode=disable)
- `POSTGRES_USER`: PostgreSQL username (default: pixie)
- `POSTGRES_PASSWORD`: PostgreSQL password (default: pixiepass)
- `POSTGRES_DB`: PostgreSQL database name (default: pixiedb)
- `MINIO_ROOT_USER`: MinIO root username (default: minioadmin)
- `MINIO_ROOT_PASSWORD`: MinIO root password (default: miniopass)
- `NATS_URL`: NATS server URL (default: nats://nats:4222)
