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
- **NATS**: localhost:4222

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
  plugins/          # Plugin directory (empty for now)
  web/              # Web UI directory (empty for now)
  deployments/      # Deployment configurations
    docker-compose.yml
    docker/
      core.Dockerfile
  scripts/          # Utility scripts
  Makefile          # Build automation
```

## Environment Variables

All secrets are configured via environment variables with sensible defaults for development:

- `POSTGRES_USER`: PostgreSQL username (default: pixie)
- `POSTGRES_PASSWORD`: PostgreSQL password (default: pixiepass)
- `POSTGRES_DB`: PostgreSQL database name (default: pixiedb)
- `MINIO_ROOT_USER`: MinIO root username (default: minioadmin)
- `MINIO_ROOT_PASSWORD`: MinIO root password (default: miniopass)
