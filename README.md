# Pixie

<div align="center">

![Pixie Logo](https://via.placeholder.com/150x150.png?text=Pixie)

**A modern, extensible photo management system**

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)
[![Docker](https://img.shields.io/badge/Docker-Ready-blue)](https://www.docker.com/)
[![React](https://img.shields.io/badge/React-UI-61dafb)](https://reactjs.org/)
[![Go](https://img.shields.io/badge/Go-Backend-00ADD8)](https://golang.org/)

</div>

## Overview

Pixie is an open-source photo management system designed for self-hosting your personal photo collection. It provides a clean, intuitive interface for organizing, viewing, and sharing your photos with powerful search capabilities and an extensible plugin architecture.

## Features

- **Modern UI**: Clean, responsive interface built with React and Tailwind CSS
- **Secure Authentication**: JWT-based authentication with support for both HS256 and RS256 algorithms
- **Photo Management**: Upload, organize, view, and delete photos
- **Thumbnails**: Automatic thumbnail generation for faster browsing
- **Extensible**: Plugin architecture for adding custom functionality
- **Docker Support**: Easy deployment with Docker and Docker Compose
- **API-First Design**: Well-documented REST API for integration with other services

## Screenshots

<div align="center">
<img src="https://via.placeholder.com/800x450.png?text=Gallery+View" alt="Gallery View" width="80%"/>
<p><em>Gallery View</em></p>

<img src="https://via.placeholder.com/800x450.png?text=Photo+Detail" alt="Photo Detail" width="80%"/>
<p><em>Photo Detail View</em></p>
</div>

## Architecture

Pixie is built with a microservices architecture:

- **Core Service**: Go-based backend that handles HTTP requests and serves the API
- **Authentication**: Integrated JWT authentication system
- **Storage**: S3-compatible object storage for photos
- **Database**: PostgreSQL for metadata storage
- **Event System**: NATS for event-driven communication between services
- **Thumbnailer Plugin**: Generates thumbnails for uploaded photos
- **UI**: React-based frontend for user interaction

## Installation

### Prerequisites

- Docker and Docker Compose
- Make (for development)
- Node.js and npm (for UI development)
- Go 1.23+ (for backend development)

### Quick Start

1. Clone the repository:

```bash
git clone https://github.com/yourusername/pixie.git
cd pixie
```

2. Start the services:

```bash
make dev
```

3. Access the UI at http://localhost:8080

## Configuration

### Environment Variables

Create a `.env` file in the root directory based on `.env.example`:

```bash
# Copy example environment file
cp .env.example .env

# Edit with your preferred editor
nano .env
```

### Authentication Configuration

The authentication system can be configured using the following environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `JWT_ALGO` | JWT algorithm (HS256 or RS256) | HS256 |
| `JWT_SECRET` | Secret key for HS256 algorithm | supersecret123 |
| `JWT_PUBLIC_KEY_FILE` | Path to public key for RS256 | |
| `JWT_PRIVATE_KEY_FILE` | Path to private key for RS256 | |

### Switching JWT Algorithms

```bash
# Switch to HS256
make switch-hs256

# Switch to RS256 (will generate keys if they don't exist)
make switch-rs256
```

## API Documentation

### Authentication Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/auth/health` | GET | Check authentication service health |
| `/api/auth/token` | POST | Generate a new JWT token |
| `/api/auth/revoke` | POST | Revoke a JWT token |

### Photo Management Endpoints

All endpoints require a valid JWT token in the Authorization header.

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/photos` | GET | List all photos |
| `/api/upload` | POST | Upload a new photo |
| `/api/photo/{id}` | GET | Get a specific photo |
| `/api/photo/{id}` | DELETE | Delete a specific photo |

### Example: Generating a Token

```bash
curl -X POST -H "Content-Type: application/json" \
  -d '{"subject":"demo"}' \
  http://localhost:8080/api/auth/token
```

### Example: Using a Token

```bash
curl -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  http://localhost:8080/api/photos
```

## Development

### Project Structure

```
pixie/
├── core/               # Core service (Go)
│   ├── auth/           # Authentication service
│   ├── db/             # Database access
│   ├── events/         # Event system
│   ├── http/           # HTTP handlers
│   ├── photo/          # Photo management
│   ├── plugin/         # Plugin system
│   └── storage/        # Storage service
├── plugins/            # Plugin implementations
│   ├── noop/           # Example no-op plugin
│   ├── thumbnailer/    # Thumbnail generation plugin
│   └── ui-react/       # React UI
├── proto/              # Protocol buffer definitions
├── deployments/        # Deployment configurations
│   ├── docker/         # Docker configurations
│   └── docker-compose.yml
└── scripts/            # Utility scripts
```

### Building and Running

```bash
# Install UI dependencies
make ui-deps

# Build UI
make ui-build

# Build plugins
make plugins

# Start all services
make dev

# Stop all services
make down
```

### Plugin Development

Pixie supports a plugin architecture for extending functionality. Plugins implement the `PhotoPlugin` gRPC service defined in `proto/plugin/v1/plugin.proto`.

## Troubleshooting

### Common Issues

1. **Authentication Issues**:
   - Check the health of the auth service: `curl http://localhost:8080/api/auth/health`
   - Verify your token is valid and not expired

2. **Service Connectivity**:
   - Check Docker container status: `docker ps`
   - View service logs: `docker logs pixie-core`

3. **Upload Problems**:
   - Ensure S3 storage is properly configured
   - Check file size limits in your configuration

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

Please make sure your code follows the existing style and includes appropriate tests.

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgments

- Built with [Go](https://golang.org/), [React](https://reactjs.org/), and [Docker](https://www.docker.com/)
- Uses [NATS](https://nats.io/) for event streaming
- Uses [PostgreSQL](https://www.postgresql.org/) for metadata storage
- Uses [MinIO](https://min.io/) for S3-compatible object storage
