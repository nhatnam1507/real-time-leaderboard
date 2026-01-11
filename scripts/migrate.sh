#!/bin/bash

# Database migration tool
# Usage: ./scripts/migrate.sh [up|down] [migration_dir]
#   up           Apply all pending migrations
#   down         Rollback all migrations
#   migration_dir  Migration directory path (required, can be relative or absolute)

set -e

# Get script directory and project root
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Change to project root directory
cd "$PROJECT_ROOT"

# Configuration (paths relative to project root)
DB_URL="${DB_URL:-postgres://postgres:postgres@localhost:5432/leaderboard?sslmode=disable}"

# Add Go bin directory to PATH
export PATH="$(go env GOPATH)/bin:${PATH}"

# Function to show usage
show_usage() {
    echo "Usage: $0 [up|down] [migration_dir]"
    echo ""
    echo "Commands:"
    echo "  up           Apply all pending migrations"
    echo "  down         Rollback all migrations"
    echo ""
    echo "Arguments:"
    echo "  migration_dir  Migration directory path (required)"
    echo "                 Can be relative to project root or absolute path"
    echo ""
    echo "Environment variables:"
    echo "  DB_URL       Database connection string (default: postgres://postgres:postgres@localhost:5432/leaderboard?sslmode=disable)"
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

# Function to validate migration directory
validate_migration_dir() {
    local migration_dir=$1
    
    if [ -z "$migration_dir" ]; then
        echo "Error: Migration directory is required"
        show_usage
    fi
    
    # Convert relative path to absolute if needed
    if [[ "$migration_dir" != /* ]]; then
        migration_dir="$PROJECT_ROOT/$migration_dir"
    fi
    
    if [ ! -d "$migration_dir" ]; then
        echo "Error: Migration directory does not exist: $migration_dir"
        exit 1
    fi
    
    echo "$migration_dir"
}

# Function to get migration table name from directory path
# Schema uses default table, dev uses custom table with suffix
get_migration_table() {
    local migration_dir=$1
    # Extract the last directory name (schema or dev)
    local dir_name=$(basename "$migration_dir")
    
    # Schema uses default table (empty string means use default)
    if [ "$dir_name" = "schema" ]; then
        echo ""
    else
        # Other directories (like dev) use custom table with suffix
        echo "schema_migrations_${dir_name}"
    fi
}

# Function to add migration table parameter to database URL
add_migration_table_to_url() {
    local db_url=$1
    local table_name=$2
    
    # If table_name is empty, use default (don't add parameter)
    if [ -z "$table_name" ]; then
        echo "$db_url"
        return
    fi
    
    # Check if URL already has query parameters
    if [[ "$db_url" == *"?"* ]]; then
        echo "${db_url}&x-migrations-table=${table_name}"
    else
        echo "${db_url}?x-migrations-table=${table_name}"
    fi
}

# Function to run migration up
migrate_up() {
    local migration_dir=$1
    
    migration_dir=$(validate_migration_dir "$migration_dir")
    local table_name=$(get_migration_table "$migration_dir")
    local db_url_with_table=$(add_migration_table_to_url "$DB_URL" "$table_name")
    
    if [ -n "$table_name" ]; then
        echo "Applying migrations from: $migration_dir (using table: $table_name)"
    else
        echo "Applying migrations from: $migration_dir (using default table)"
    fi
    migrate -path "$migration_dir" -database "$db_url_with_table" up
    
    echo "✓ Migrations applied successfully"
}

# Function to run migration down
migrate_down() {
    local migration_dir=$1
    
    migration_dir=$(validate_migration_dir "$migration_dir")
    local table_name=$(get_migration_table "$migration_dir")
    local db_url_with_table=$(add_migration_table_to_url "$DB_URL" "$table_name")
    
    if [ -n "$table_name" ]; then
        echo "Rolling back migrations from: $migration_dir (using table: $table_name)"
    else
        echo "Rolling back migrations from: $migration_dir (using default table)"
    fi
    migrate -path "$migration_dir" -database "$db_url_with_table" down
    
    echo "✓ Migrations rolled back successfully"
}

# Main function - handles CLI input and execution
main() {
    # Validate parameter count
    if [ $# -eq 0 ]; then
        echo "Error: Missing command"
        show_usage
    elif [ $# -lt 2 ]; then
        echo "Error: Missing migration directory. Expected 2 parameters, got $#"
        show_usage
    elif [ $# -gt 2 ]; then
        echo "Error: Too many parameters. Expected 2 parameters, got $#"
        show_usage
    fi
    
    # Check if migrate tool is available
    check_migrate_tool
    
    # Parse arguments
    local direction=$1
    local migration_dir=$2
    
    # Execute migration command
    case "$direction" in
        up)
            migrate_up "$migration_dir"
            ;;
        down)
            migrate_down "$migration_dir"
            ;;
        *)
            echo "Error: Invalid command '$direction'"
            show_usage
            ;;
    esac
}

# Execute main function with all arguments
main "$@"
