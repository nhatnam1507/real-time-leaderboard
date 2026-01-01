#!/bin/bash

# Database migration tool
# Usage: ./scripts/migrate.sh [up|down] [version]
#   up      Apply all pending migrations (or up to version if specified)
#   down    Rollback migrations (or down to version if specified)
#   version Optional: specific version to migrate to

set -e

# Get script directory and project root
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Change to project root directory
cd "$PROJECT_ROOT"

# Configuration (paths relative to project root)
MIGRATIONS_DIR="$PROJECT_ROOT/internal/shared/database/migrations"
DB_URL="${DB_URL:-postgres://postgres:postgres@localhost:5432/leaderboard?sslmode=disable}"

# Add Go bin directory to PATH
export PATH="$(go env GOPATH)/bin:${PATH}"

# Function to show usage
show_usage() {
    echo "Usage: $0 [up|down] [version]"
    echo ""
    echo "Commands:"
    echo "  up      Apply all pending migrations (or up to version if specified)"
    echo "  down    Rollback migrations (or down to version if specified)"
    echo ""
    echo "Options:"
    echo "  version Optional: specific version to migrate to"
    echo ""
    echo "Environment variables:"
    echo "  DB_URL  Database connection string (default: postgres://postgres:postgres@localhost:5432/leaderboard?sslmode=disable)"
    exit 1
}

# Function to check if migrate tool is installed
check_migrate_tool() {
    if ! command -v migrate &> /dev/null; then
        echo "Error: migrate tool not found"
        echo "Install it with:"
        echo "  go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest"
        exit 1
    fi
}

# Function to run migration up
migrate_up() {
    local version=$1
    
    echo "Applying migrations..."
    if [ -n "$version" ]; then
        echo "  Migrating up to version: $version"
        migrate -path "$MIGRATIONS_DIR" -database "$DB_URL" up "$version"
    else
        echo "  Migrating to latest version"
        migrate -path "$MIGRATIONS_DIR" -database "$DB_URL" up
    fi
    
    echo "✓ Migrations applied successfully"
}

# Function to run migration down
migrate_down() {
    local version=$1
    
    echo "Rolling back migrations..."
    if [ -n "$version" ]; then
        echo "  Rolling back to version: $version"
        migrate -path "$MIGRATIONS_DIR" -database "$DB_URL" down "$version"
    else
        echo "  Rolling back one migration"
        migrate -path "$MIGRATIONS_DIR" -database "$DB_URL" down 1
    fi
    
    echo "✓ Migrations rolled back successfully"
}

# Main function - handles CLI input and execution
main() {
    # Validate parameter count
    if [ $# -eq 0 ]; then
        echo "Error: Missing command"
        show_usage
    elif [ $# -gt 2 ]; then
        echo "Error: Too many parameters. Expected 1-2 parameters, got $#"
        show_usage
    fi
    
    # Check if migrate tool is available
    check_migrate_tool
    
    # Parse arguments
    local direction=$1
    local version=${2:-}
    
    # Execute migration command
    case "$direction" in
        up)
            migrate_up "$version"
            ;;
        down)
            migrate_down "$version"
            ;;
        *)
            echo "Error: Invalid command '$direction'"
            show_usage
            ;;
    esac
}

# Execute main function with all arguments
main "$@"
