#!/bin/bash

# mini-build.sh - Simplified build system for boxing simulator

set -e  # Exit on any error

# Simple build function
build() {
    echo "Building application..."
    go build -o boxing-server ./cmd/server
    echo "Build completed successfully"
}

# Simple test function
test() {
    echo "Running tests..."
    go test -v ./...
    echo "Tests completed successfully"
}

# Main CI-like function
ci_pipeline() {
    echo "Starting CI pipeline..."

    # Run linters first
    echo "Running linters..."
    go vet ./...
    golangci-lint run ./...

    # Build and test
    build
    test

    echo "CI pipeline completed successfully"
}

# Command line interface
case "${1:-build}" in
    "build")
        build
        ;;
    "test")
        test
        ;;
    "ci")
        ci_pipeline
        ;;
    "clean")
        rm -f boxing-server
        echo "Cleaned build artifacts"
        ;;
    *)
        echo "Usage: $0 {build|test|ci|clean}"
        exit 1
        ;;
esac

echo "Process completed"