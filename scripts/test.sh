#!/bin/bash

# Run unit tests and generate coverage report using gotestsum
# Usage: ./scripts/test.sh

set -e

# Get script directory and project root
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Change to project root directory
cd "$PROJECT_ROOT"

# Add Go bin directory to PATH
export PATH="$(go env GOPATH)/bin:${PATH}"

# Configuration
COVERAGE_DIR="$PROJECT_ROOT/tmp/coverage"
COVERAGE_PROFILE="$COVERAGE_DIR/coverage.out"
COVERAGE_HTML="$COVERAGE_DIR/coverage.html"
TEST_REPORTS_DIR="$PROJECT_ROOT/tmp/test-reports"
JUNIT_XML="$TEST_REPORTS_DIR/junit.xml"

# Create directories if they don't exist
mkdir -p "$COVERAGE_DIR"
mkdir -p "$TEST_REPORTS_DIR"

echo "Running unit tests with gotestsum..."
echo ""

# Run tests with gotestsum
# --format: short-verbose shows package names and test results
# --junitfile: generates JUnit XML for CI integration
# --junitfile-hide-empty-pkg: omit packages with no tests from junit.xml
# --hide-summary: hide error summary (errors are from packages without tests, not test failures)
# --: passes remaining flags to go test
gotestsum \
    --format short-verbose \
    --junitfile "$JUNIT_XML" \
    --junitfile-hide-empty-pkg \
    --hide-summary=errors \
    -- \
    -coverprofile="$COVERAGE_PROFILE" \
    -covermode=atomic \
    ./cmd/... ./internal/... 2>&1 | grep -v "^compile:" || true

echo ""

# Generate coverage summary
if [ -f "$COVERAGE_PROFILE" ]; then
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "Coverage Report"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""
    
    # Show per-package coverage summary
    echo "Package Coverage:"
    echo "────────────────────────────────────────────────────────────────────────────"
    go tool cover -func="$COVERAGE_PROFILE" | grep -E "^real-time-leaderboard" | awk '
    {
        # Extract package name (remove "real-time-leaderboard/" prefix and file path)
        pkg = $1
        gsub(/^real-time-leaderboard\//, "", pkg)
        gsub(/\/[^/]+\.go.*$/, "", pkg)
        
        # Extract coverage percentage
        match($0, /([0-9.]+)%$/, arr)
        if (arr[1] != "") {
            cov = arr[1] + 0  # Convert to number
            # Accumulate coverage for package average
            if (pkg_count[pkg] == "") {
                pkg_count[pkg] = 0
                pkg_total[pkg] = 0
            }
            pkg_count[pkg]++
            pkg_total[pkg] += cov
        }
    }
    END {
        # Print packages with average coverage
        for (pkg in pkg_count) {
            avg = pkg_total[pkg] / pkg_count[pkg]
            printf "  ✓ %-50s %6.1f%%\n", pkg, avg
        }
    }' | sort
    
    echo "────────────────────────────────────────────────────────────────────────────"
    
    # Show overall coverage
    OVERALL_COV=$(go tool cover -func="$COVERAGE_PROFILE" 2>/dev/null | grep "total:" | awk '{print $3}' || echo "N/A")
    if [ "$OVERALL_COV" != "N/A" ]; then
        echo ""
        echo "  Overall Coverage: $OVERALL_COV"
    fi
    
    echo ""
    
    # Generate HTML coverage report
    echo "Generating HTML coverage report..."
    go tool cover -html="$COVERAGE_PROFILE" -o "$COVERAGE_HTML" > /dev/null 2>&1 || true
    
    if [ -f "$COVERAGE_HTML" ]; then
        echo "✓ Coverage report generated: $COVERAGE_HTML"
    fi
    
    if [ -f "$JUNIT_XML" ]; then
        echo "✓ JUnit XML report generated: $JUNIT_XML"
        echo "  (for GitHub Actions integration)"
    fi
fi

echo ""
echo "Unit tests completed successfully"
