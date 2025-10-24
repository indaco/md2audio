#!/bin/bash
set -e

# Parse flags
VERBOSE=""
if [[ "$1" == "-v" ]]; then
    VERBOSE="-v"
fi

# Clean up old coverage files
rm -f coverage.txt coverage-*.txt

echo "Running tests..."

# Get list of packages excluding testhelpers and internal/tts (interface-only package)
packages=$(go list ./... | grep -Ev 'internal/testhelpers$|internal/tts$')

# Run tests for each package and generate individual coverage files
for pkg in $packages; do
    safe_pkg=$(echo "$pkg" | tr '/' '-')
    go test $VERBOSE -count=1 -timeout 30s -covermode=atomic -coverprofile="coverage-${safe_pkg}.txt" "$pkg"
done

# Merge all coverage files into one
echo "mode: atomic" > coverage.txt
for file in coverage-*.txt; do
    if [ -f "$file" ]; then
        tail -n +2 "$file" >> coverage.txt
    fi
done

# Clean up individual coverage files
rm -f coverage-*.txt

echo ""
echo "Total Coverage:"
go tool cover -func=coverage.txt | grep total | awk '{print "  " $3}'
