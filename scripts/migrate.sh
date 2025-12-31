#!/bin/bash

# Database migration script
# Usage: ./scripts/migrate.sh [up|down] [version]

set -e

MIGRATIONS_DIR="./internal/infrastructure/database/migrations"
DB_URL="${DB_URL:-postgres://postgres:postgres@localhost:5432/leaderboard?sslmode=disable}"

if [ -z "$1" ]; then
    echo "Usage: $0 [up|down] [version]"
    exit 1
fi

DIRECTION=$1
VERSION=${2:-}

if command -v migrate &> /dev/null; then
    if [ -n "$VERSION" ]; then
        migrate -path "$MIGRATIONS_DIR" -database "$DB_URL" $DIRECTION $VERSION
    else
        migrate -path "$MIGRATIONS_DIR" -database "$DB_URL" $DIRECTION
    fi
else
    echo "migrate tool not found. Install it with:"
    echo "  go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest"
    exit 1
fi

