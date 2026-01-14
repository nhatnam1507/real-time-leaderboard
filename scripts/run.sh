#!/bin/bash

# Run script for starting the application
# Usage: ./scripts/run.sh [dev|prod-like]
#   dev:       Start dependency services and run app with air (hot reload)
#   prod-like: Start full Docker Swarm stack (app + deps in containers)

set -e

# Get script directory and project root
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Change to project root directory
cd "$PROJECT_ROOT"

# Add Go bin directory to PATH
export PATH="$(go env GOPATH)/bin:${PATH}"

# Configuration (paths relative to project root)
COMPOSE_DEV_FILE="$PROJECT_ROOT/docker/docker-compose.dev.yml"
COMPOSE_SWARM_FILE="$PROJECT_ROOT/docker/docker-compose.swarm.yml"
MIGRATION_SCRIPT="$SCRIPT_DIR/migrate.sh"
STACK_NAME="leaderboard"
# Migrations run on host, so always use localhost (ports are exposed)
DATABASE_URL="postgres://postgres:postgres@localhost:5432/leaderboard?sslmode=disable"

# Function to show usage
show_usage() {
    echo "Usage: $0 [dev|prod-like]"
    echo ""
    echo "Modes:"
    echo "  dev       Start dependency services and run app with air (hot reload)"
    echo "  prod-like Start full Docker Swarm stack (app + deps in containers)"
    exit 1
}

# Function to check if a container is running
is_container_running() {
    local container_name=$1
    docker ps --format '{{.Names}}' | grep -q "^${container_name}$"
}

# Function to wait for stack to be removed
wait_for_stack_removal() {
    local stack_name=$1
    local timeout=${2:-60}
    local interval=${3:-2}
    local elapsed=0
    
    echo "Waiting for stack '$stack_name' to be removed..."
    while [ $elapsed -lt $timeout ]; do
        if ! docker stack ls | grep -q "^${stack_name} "; then
            echo "✓ Stack removed"
            return 0
        fi
        sleep $interval
        elapsed=$((elapsed + interval))
    done
    
    echo "Warning: Stack removal timeout after ${timeout}s"
    return 1
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
    docker compose -f "$COMPOSE_DEV_FILE" up -d
}

# Function to wait for services to be ready
# Note: wait4x runs on host, so we check localhost (ports are exposed)
# In prod-like mode, app container uses service names (postgres/redis) from Docker Swarm network
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
    docker compose -f "$COMPOSE_DEV_FILE" down 2>/dev/null || true
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

# Function to start prod-like mode (Docker Swarm stack)
start_prod_like_mode() {
    echo "Starting Docker Swarm stack..."
    
    # Check if Docker Swarm is initialized
    if ! docker info | grep -q "Swarm: active"; then
        echo "Initializing Docker Swarm..."
        docker swarm init 2>/dev/null || {
            echo "Warning: Swarm may already be initialized or failed to initialize"
        }
    fi
    
    # Check if stack already exists and remove it first
    if docker stack ls | grep -q "^${STACK_NAME} "; then
        echo "Removing existing stack..."
        docker stack rm "$STACK_NAME" 2>/dev/null || true
        wait_for_stack_removal "$STACK_NAME" 60 2
    fi
    
    # Deploy stack to Swarm first
    # Swarm uses leaderboard_prod network, dev uses leaderboard_dev network (no conflict)
    echo "Deploying stack to Docker Swarm..."
    docker stack deploy -c "$COMPOSE_SWARM_FILE" "$STACK_NAME"
    
    # Wait for swarm postgres to be ready
    echo "Waiting for swarm postgres to be ready..."
    wait4x postgresql "$DATABASE_URL" \
        --timeout 120s \
        --interval 2s || {
        echo "Error: Failed to connect to swarm postgres"
        exit 1
    }
    
    # Run schema migrations against swarm postgres (no dev seed in production-like mode)
    echo "Running migrations against swarm postgres..."
    run_migrations "migrations/schema"
    
    echo "Stack deployed. Use 'docker stack services $STACK_NAME' to view services."
    echo "Use 'docker stack logs -f $STACK_NAME' to view logs."
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
        prod-like)
            start_prod_like_mode
            ;;
        *)
            echo "Error: Invalid mode '$mode'"
            show_usage
            ;;
    esac
}

# Execute main function with all arguments
main "$@"
