#!/bin/bash

# Run script for starting the application
# Usage: ./scripts/run.sh [dev|all]
#   dev:  Start dependency services and run app with air (hot reload)
#   all:  Start full docker compose environment (app + deps in containers)

set -e

# Get script directory and project root
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Change to project root directory
cd "$PROJECT_ROOT"

# Add Go bin directory to PATH
export PATH="$(go env GOPATH)/bin:${PATH}"

# Configuration (paths relative to project root)
COMPOSE_DEPENDENCIES_FILE="$PROJECT_ROOT/docker/docker-compose.deps.yml"
COMPOSE_FULL_STACK_FILE="$PROJECT_ROOT/docker/docker-compose.yml"
MIGRATION_SCRIPT="$SCRIPT_DIR/migrate.sh"
# Migrations run on host, so always use localhost (ports are exposed)
DATABASE_URL="postgres://postgres:postgres@localhost:5432/leaderboard?sslmode=disable"

# Function to show usage
show_usage() {
    echo "Usage: $0 [dev|all]"
    echo ""
    echo "Modes:"
    echo "  dev  Start dependency services and run app with air (hot reload)"
    echo "  all  Start full docker compose environment (app + deps in containers)"
    exit 1
}

# Function to check if a container is running
is_container_running() {
    local container_name=$1
    docker ps --format '{{.Names}}' | grep -q "^${container_name}$"
}

# Function to ensure dependency services are running
ensure_dependency_services_running() {
    local postgres_running=$(is_container_running "leaderboard-postgres" && echo "yes" || echo "no")
    local redis_running=$(is_container_running "leaderboard-redis" && echo "yes" || echo "no")
    
    if [ "$postgres_running" = "yes" ] && [ "$redis_running" = "yes" ]; then
        echo "Dependency services (postgres, redis) are already running"
        return 0
    fi
    
    echo "Starting dependency services..."
    docker compose -f "$COMPOSE_DEPENDENCIES_FILE" up -d
}

# Function to wait for services to be ready
# Note: wait4x runs on host, so we check localhost (ports are exposed)
# In all mode, app container uses service names (postgres/redis) which is configured in docker-compose.yml
wait_for_services() {
    echo "Waiting for services to be ready..."
    
    # Wait for PostgreSQL
    echo "  Waiting for PostgreSQL at localhost:5432..."
    wait4x postgresql "$DATABASE_URL" \
        --timeout 60s \
        --interval 2s || {
        echo "Error: Failed to connect to PostgreSQL"
        exit 1
    }
    
    # Wait for Redis
    echo "  Waiting for Redis at localhost:6379..."
    wait4x redis "redis://localhost:6379" \
        --timeout 60s \
        --interval 2s || {
        echo "Error: Failed to connect to Redis"
        exit 1
    }
    
    echo "✓ All services are ready"
}

# Function to run database migrations
run_migrations() {
    local migration_dir=$1
    echo "Running database migrations from: $migration_dir"
    DB_URL="$DATABASE_URL" "$MIGRATION_SCRIPT" up "$migration_dir"
}

# Cleanup function to stop dependency services
stop_dependency_services() {
    echo ""
    echo "Stopping dependency services..."
    docker compose -f "$COMPOSE_DEPENDENCIES_FILE" down 2>/dev/null || true
    echo "Cleanup complete"
    exit 0
}

# Function to start development mode
start_development_mode() {
    # Set up signal traps for cleanup
    trap stop_dependency_services SIGHUP SIGINT SIGTERM
    
    echo "Starting development environment..."
    
    # Ensure dependency services are running
    ensure_dependency_services_running
    
    # Wait for services to be ready
    wait_for_services
    
    # Run schema migrations
    run_migrations "migrations/schema"
    
    # Run dev seed migrations
    run_migrations "migrations/dev"
    
    # Start application with air (hot reload)
    echo "Starting application with air..."
    echo "Press Ctrl+C to stop and cleanup..."
    air
}

# Function to start full stack mode (all services in containers)
start_full_stack_mode() {
    echo "Starting full docker compose environment..."
    
    # Ensure dependency services are running
    ensure_dependency_services_running
    
    # Wait for services to be ready
    wait_for_services
    
    # Run schema migrations (no dev seed in production-like mode)
    run_migrations "migrations/schema"
    
    # Start app container
    # App container uses service names (postgres/redis) from docker network (configured in docker-compose.yml)
    if is_container_running "leaderboard-app"; then
        echo "Application container is already running"
    else
        echo "Starting application container..."
        # docker-compose.yml has depends_on with health checks, so it will wait for deps to be healthy
        docker compose -f "$COMPOSE_FULL_STACK_FILE" up -d app
    fi
    
    echo "Services are running. Use 'docker compose -f $COMPOSE_FULL_STACK_FILE logs -f' to view logs."
}

# Main function - handles CLI input and execution
main() {
    # Validate parameter count
    if [ $# -eq 0 ]; then
        echo "Error: Missing mode parameter"
        show_usage
    elif [ $# -gt 1 ]; then
        echo "Error: Too many parameters. Expected 1 parameter, got $#"
        show_usage
    fi
    
    # Get mode from first parameter
    local mode=$1
    
    # Execute based on mode
    case "$mode" in
        dev)
            start_development_mode
            ;;
        all)
            start_full_stack_mode
            ;;
        *)
            echo "Error: Invalid mode '$mode'"
            show_usage
            ;;
    esac
}

# Execute main function with all arguments
main "$@"
