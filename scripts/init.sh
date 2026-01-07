#!/bin/bash

# Initialize development environment
# Usage: ./scripts/init.sh [dev|ci]
#   dev:  Install all tools and check docker, go (default)
#   ci:   Only check lint and go installation

set -e

# Get script directory and project root
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Change to project root directory
cd "$PROJECT_ROOT"

# Add Go bin directory to PATH
export PATH="$(go env GOPATH)/bin:${PATH}"

# Configuration
GITHOOKS_DIR="$PROJECT_ROOT/.githooks"

# Function to show usage
show_usage() {
    echo "Usage: $0 [dev|ci]"
    echo ""
    echo "Modes:"
    echo "  dev  Install all tools and check docker, go (default)"
    echo "  ci   Only check lint and go installation"
    exit 1
}

# Function to check if a command exists (including in Go bin)
command_exists() {
    command -v "$1" &> /dev/null
}

# Function to check if Go is installed
check_go() {
    if ! command_exists go; then
        echo "Error: Go is not installed. Please install Go first."
        exit 1
    fi
    echo "✓ Go is installed"
}

# Function to check golangci-lint
check_golangci_lint() {
    if ! command_exists golangci-lint; then
        echo "Error: golangci-lint is not installed."
        return 1
    fi
    echo "✓ golangci-lint is installed"
    return 0
}

# Function to install golangci-lint
install_golangci_lint() {
    echo "Installing golangci-lint..."
    curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin latest > /dev/null 2>&1
    echo "✓ golangci-lint installed"
}

# Function to check migrate tool
check_migrate() {
    if ! command_exists migrate; then
        echo "Error: migrate tool is not installed."
        return 1
    fi
    echo "✓ migrate tool is installed"
    return 0
}

# Function to install migrate tool
install_migrate() {
    echo "Installing migrate tool..."
    go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest > /dev/null 2>&1
    echo "✓ migrate tool installed"
}

# Function to check air
check_air() {
    if ! command_exists air; then
        echo "Error: air is not installed."
        return 1
    fi
    echo "✓ air is installed"
    return 0
}

# Function to install air for hot reload
install_air() {
    echo "Installing air..."
    go install github.com/air-verse/air@latest > /dev/null 2>&1
    echo "✓ air installed"
}

# Function to check wait4x
check_wait4x() {
    if ! command_exists wait4x; then
        echo "Error: wait4x is not installed."
        return 1
    fi
    echo "✓ wait4x is installed"
    return 0
}

# Function to install wait4x for service health checking
install_wait4x() {
    echo "Installing wait4x..."
    go install wait4x.dev/v3/cmd/wait4x@latest > /dev/null 2>&1
    echo "✓ wait4x installed"
}

# Function to check act
check_act() {
    if ! command_exists act; then
        echo "Error: act is not installed."
        return 1
    fi
    echo "✓ act is installed"
    return 0
}

# Function to install act for local GitHub Actions testing
install_act() {
    echo "Installing act..."
    # Install act to Go bin directory (user-accessible, no sudo needed)
    # Download latest release binary for Linux
    ACT_VERSION=$(curl -s https://api.github.com/repos/nektos/act/releases/latest | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    if [ -z "$ACT_VERSION" ]; then
        ACT_VERSION="v0.2.60"  # Fallback version
    fi
    curl -sL "https://github.com/nektos/act/releases/download/${ACT_VERSION}/act_Linux_x86_64.tar.gz" | \
        tar -xz -C /tmp && \
        mv /tmp/act $(go env GOPATH)/bin/act && \
        chmod +x $(go env GOPATH)/bin/act
    echo "✓ act installed"
}

# Function to check docker
check_docker() {
    if ! command_exists docker; then
        echo "⚠ Warning: docker is not installed. Docker is required for local full run with docker compose."
        return 1
    else
        echo "✓ docker is installed"
        return 0
    fi
}

# Function to check docker-compose
check_docker_compose() {
    if ! command_exists docker-compose && ! docker compose version &> /dev/null 2>&1; then
        echo "⚠ Warning: docker-compose is not installed. Docker Compose is required for local full run with docker compose."
        return 1
    else
        echo "✓ docker-compose is available"
        return 0
    fi
}

# Function to download Go dependencies
download_go_dependencies() {
    echo "Downloading Go dependencies..."
    go mod download > /dev/null 2>&1
    go mod tidy > /dev/null 2>&1
    echo "✓ Go dependencies downloaded"
}

# Function to configure git hooks
configure_git_hooks() {
    if [ -d "$GITHOOKS_DIR" ]; then
        # Check if git hooks path is already configured
        CURRENT_HOOKS_PATH=$(git config --get core.hooksPath 2>/dev/null || echo "")
        if [ "$CURRENT_HOOKS_PATH" != "$GITHOOKS_DIR" ]; then
            git config core.hooksPath "$GITHOOKS_DIR"
            echo "✓ Git hooks configured to use .githooks directory"
        else
            echo "✓ Git hooks already configured"
        fi
    fi
}

# Function to initialize development mode
init_dev() {
    echo "Initializing development environment..."
    echo ""
    
    TOOLS_INSTALLED=0
    TOOLS_MISSING=0
    
    # Check Go
    check_go
    echo ""
    
    # Install all tools (check first, install if missing)
    if ! check_golangci_lint; then
        install_golangci_lint
        TOOLS_INSTALLED=$((TOOLS_INSTALLED + 1))
    fi
    
    if ! check_migrate; then
        install_migrate
        TOOLS_INSTALLED=$((TOOLS_INSTALLED + 1))
    fi
    
    if ! check_air; then
        install_air
        TOOLS_INSTALLED=$((TOOLS_INSTALLED + 1))
    fi
    
    if ! check_wait4x; then
        install_wait4x
        TOOLS_INSTALLED=$((TOOLS_INSTALLED + 1))
    fi
    
    if ! check_act; then
        install_act
        TOOLS_INSTALLED=$((TOOLS_INSTALLED + 1))
    fi
    echo ""
    
    # Check Docker
    if ! check_docker; then TOOLS_MISSING=$((TOOLS_MISSING + 1)); fi
    if ! check_docker_compose; then TOOLS_MISSING=$((TOOLS_MISSING + 1)); fi
    echo ""
    
    # Download Go dependencies
    download_go_dependencies
    echo ""
    
    # Configure git hooks
    configure_git_hooks
    echo ""
    
    # Summary
    if [ $TOOLS_INSTALLED -gt 0 ]; then
        echo "Development environment initialized successfully! ($TOOLS_INSTALLED tool(s) installed)"
    else
        echo "Development environment is ready! (all tools already installed)"
    fi
    
    if [ $TOOLS_MISSING -gt 0 ]; then
        echo "⚠ Warning: $TOOLS_MISSING required tool(s) are missing. Some targets may not work."
    fi
}

# Function to initialize CI mode
init_ci() {
    echo "Initializing CI environment..."
    echo ""
    
    # Check Go
    check_go
    echo ""
    
    # Check and install golangci-lint if needed
    if ! check_golangci_lint; then
        install_golangci_lint
    fi
    echo ""
    
    echo "CI environment is ready!"
}

# Main function - handles CLI input and execution
main() {
    # Get mode from first parameter (default to dev)
    local mode=${1:-dev}
    
    # Validate parameter count
    if [ $# -gt 1 ]; then
        echo "Error: Too many parameters. Expected 0-1 parameter, got $#"
        show_usage
    fi
    
    # Execute based on mode
    case "$mode" in
        dev)
            init_dev
            ;;
        ci)
            init_ci
            ;;
        *)
            echo "Error: Invalid mode '$mode'"
            show_usage
            ;;
    esac
}

# Execute main function with all arguments
main "$@"
