#!/bin/bash

# Validate GitHub Actions workflow YAML syntax
# Uses act --list to check syntax without running workflows

set -e

# Check if act is available
if ! command -v act &> /dev/null; then
    echo "Warning: act not found. Skipping workflow validation."
    echo "Install act by running: make init"
    exit 0
fi

echo "Validating GitHub Actions workflows..."

# Use act --list to validate workflow syntax
# This parses all workflows and lists jobs without actually running them
# If there are syntax errors, act will fail
if act --list > /dev/null 2>&1; then
    echo "✓ All workflows validated successfully"
    exit 0
else
    echo "✗ Workflow syntax errors detected"
    echo ""
    # Show the actual errors
    act --list 2>&1 || true
    exit 1
fi
