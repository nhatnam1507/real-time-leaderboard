#!/bin/bash

# Initialize development environment
# Installs all required tools and libraries for linting, testing, building, and local full run with docker compose
# Only installs tools that are missing

set -e

# Check if SILENT mode is enabled (when called as dependency)
SILENT=${SILENT:-0}

# Add Go bin directory to PATH for checking
export PATH="$(go env GOPATH)/bin:${PATH}"

# Function to echo only if not silent
echo_if_verbose() {
    if [ "$SILENT" != "1" ]; then
        echo "$@"
    fi
}

# Function to check if a command exists (including in Go bin)
command_exists() {
    command -v "$1" &> /dev/null
}

# Check if Go is installed
if ! command_exists go; then
    echo "Error: Go is not installed. Please install Go first."
    exit 1
fi

TOOLS_INSTALLED=0
TOOLS_MISSING=0

# Check and install golangci-lint
if ! command_exists golangci-lint; then
    echo_if_verbose "Installing golangci-lint..."
    curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin latest > /dev/null 2>&1
    echo_if_verbose "✓ golangci-lint installed"
    TOOLS_INSTALLED=$((TOOLS_INSTALLED + 1))
else
    echo_if_verbose "✓ golangci-lint already installed"
fi

# Check and install migrate tool
if ! command_exists migrate; then
    echo_if_verbose "Installing migrate tool..."
    go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest > /dev/null 2>&1
    echo_if_verbose "✓ migrate tool installed"
    TOOLS_INSTALLED=$((TOOLS_INSTALLED + 1))
else
    echo_if_verbose "✓ migrate tool already installed"
fi

# Check and install air for hot reload
if ! command_exists air; then
    echo_if_verbose "Installing air..."
    go install github.com/air-verse/air@latest > /dev/null 2>&1
    echo_if_verbose "✓ air installed"
    TOOLS_INSTALLED=$((TOOLS_INSTALLED + 1))
else
    echo_if_verbose "✓ air already installed"
fi

# Check and install wait4x for service health checking
if ! command_exists wait4x; then
    echo_if_verbose "Installing wait4x..."
    go install wait4x.dev/v3/cmd/wait4x@latest > /dev/null 2>&1
    echo_if_verbose "✓ wait4x installed"
    TOOLS_INSTALLED=$((TOOLS_INSTALLED + 1))
else
    echo_if_verbose "✓ wait4x already installed"
fi

# Verify docker is installed
if ! command_exists docker; then
    echo_if_verbose "⚠ Warning: docker is not installed. Docker is required for local full run with docker compose."
    TOOLS_MISSING=$((TOOLS_MISSING + 1))
else
    echo_if_verbose "✓ docker is installed"
fi

# Verify docker-compose is installed
if ! command_exists docker-compose && ! docker compose version &> /dev/null 2>&1; then
    echo_if_verbose "⚠ Warning: docker-compose is not installed. Docker Compose is required for local full run with docker compose."
    TOOLS_MISSING=$((TOOLS_MISSING + 1))
else
    echo_if_verbose "✓ docker-compose is available"
fi

# Download Go dependencies
echo_if_verbose "Downloading Go dependencies..."
go mod download > /dev/null 2>&1
go mod tidy > /dev/null 2>&1
echo_if_verbose "✓ Go dependencies downloaded"

if [ "$SILENT" != "1" ]; then
    echo ""
    if [ $TOOLS_INSTALLED -gt 0 ]; then
        echo "Development environment initialized successfully! ($TOOLS_INSTALLED tool(s) installed)"
    else
        echo "Development environment is ready! (all tools already installed)"
    fi

    if [ $TOOLS_MISSING -gt 0 ]; then
        echo "⚠ Warning: $TOOLS_MISSING required tool(s) are missing. Some targets may not work."
    fi
fi

