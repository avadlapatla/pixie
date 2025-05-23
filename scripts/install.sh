#!/bin/bash
set -e

# Print colored output
print_info() {
  echo -e "\033[0;34m[INFO]\033[0m $1"
}

print_success() {
  echo -e "\033[0;32m[SUCCESS]\033[0m $1"
}

print_error() {
  echo -e "\033[0;31m[ERROR]\033[0m $1"
}

# Check if Docker is installed
check_docker() {
  print_info "Checking if Docker is installed..."
  if ! command -v docker &> /dev/null; then
    print_error "Docker is not installed. Please install Docker first."
    print_info "Visit https://docs.docker.com/get-docker/ for installation instructions."
    exit 1
  fi
  print_success "Docker is installed."
}

# Check if Docker Compose is installed
check_docker_compose() {
  print_info "Checking if Docker Compose is installed..."
  if ! docker compose version &> /dev/null; then
    print_error "Docker Compose is not installed or not in PATH."
    print_info "For Docker Desktop, Compose should be included."
    print_info "For standalone Docker, visit https://docs.docker.com/compose/install/"
    exit 1
  fi
  print_success "Docker Compose is installed."
}

# Check if Docker daemon is running
check_docker_running() {
  print_info "Checking if Docker daemon is running..."
  if ! docker info &> /dev/null; then
    print_error "Docker daemon is not running. Please start Docker and try again."
    exit 1
  fi
  print_success "Docker daemon is running."
}

# Main function
main() {
  print_info "Starting Pixie installation..."
  
  # Check prerequisites
  check_docker
  check_docker_compose
  check_docker_running
  
  # Navigate to project root (assuming script is run from project root or scripts directory)
  if [[ $(basename "$PWD") == "scripts" ]]; then
    cd ..
  fi
  
  print_info "Starting Pixie services with Docker Compose..."
  make dev
  
  print_success "Pixie is now running!"
  print_info "Access the services at:"
  print_info "- Core API: http://localhost:8080/healthz"
  print_info "- MinIO Console: http://localhost:9001"
  print_info "- PostgreSQL: localhost:5432"
  print_info "- NATS: localhost:4222"
}

# Run the main function
main
