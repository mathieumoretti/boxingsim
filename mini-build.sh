#!/bin/bash

# mini-build.sh - Incremental build system for boxing simulator

set -e  # Exit on any error

BUILD_DIR="build"
LOG_FILE="$BUILD_DIR/build.log"

# Create build directory
mkdir -p $BUILD_DIR

# Logging function
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a $LOG_FILE
}

# Function to check if a package is built
is_built() {
    local pkg=$1
    [[ -f "$BUILD_DIR/$pkg.built" ]]
}

# Function to mark package as built
mark_built() {
    local pkg=$1
    touch "$BUILD_DIR/$pkg.built"
}

# Function to build a package with dependencies
build_package() {
    local pkg=$1
    log "Building $pkg..."

    case $pkg in
        "model")
            go build -o "$BUILD_DIR/model" ./internal/model
            mark_built "model"
            ;;
        "service")
            if ! is_built "model"; then
                build_package "model"
            fi
            go build -o "$BUILD_DIR/service" ./internal/service
            mark_built "service"
            ;;
        "db")
            if ! is_built "model"; then
                build_package "model"
            fi
            go build -o "$BUILD_DIR/db" ./internal/db
            mark_built "db"
            ;;
        "platform")
            if ! is_built "model"; then
                build_package "model"
            fi
            go build -o "$BUILD_DIR/platform" ./internal/platform/config
            go build -o "$BUILD_DIR/platform" ./internal/platform/database
            go build -o "$BUILD_DIR/platform" ./internal/platform/logger
            go build -o "$BUILD_DIR/platform" ./internal/platform/redis
            mark_built "platform"
            ;;
        "auth")
            if ! is_built "model"; then
                build_package "model"
            fi
            if ! is_built "platform"; then
                build_package "platform"
            fi
            go build -o "$BUILD_DIR/auth" ./internal/auth
            mark_built "auth"
            ;;
        "handler")
            if ! is_built "service"; then
                build_package "service"
            fi
            if ! is_built "model"; then
                build_package "model"
            fi
            go build -o "$BUILD_DIR/handler" ./internal/handler
            mark_built "handler"
            ;;
        "server")
            if ! is_built "auth"; then
                build_package "auth"
            fi
            if ! is_built "db"; then
                build_package "db"
            fi
            if ! is_built "handler"; then
                build_package "handler"
            fi
            if ! is_built "platform"; then
                build_package "platform"
            fi
            go build -o "$BUILD_DIR/boxing-server" ./cmd/server
            mark_built "server"
            ;;
        *)
            log "Unknown package: $pkg"
            return 1
            ;;
    esac

    log "Successfully built $pkg"
}

# Run tests for a specific package
run_tests() {
    local pkg=$1
    log "Running tests for $pkg..."

    go test -v ./internal/$pkg
    log "Tests for $pkg completed successfully"
}

# Main build function
main_build() {
    log "Starting incremental build process"

    # Build in dependency order
    build_package "model"
    build_package "service"
    build_package "db"
    build_package "platform"
    build_package "auth"
    build_package "handler"
    build_package "server"

    log "All packages built successfully"
}

# Main test function
main_test() {
    log "Running comprehensive tests"

    # Run all package tests
    run_tests "model"
    run_tests "service"
    run_tests "db"
    run_tests "handler"
    run_tests "auth"

    log "All tests completed"
}

# Main CI-like function
ci_pipeline() {
    log "Starting CI pipeline"

    # Clean previous build
    rm -rf $BUILD_DIR
    mkdir -p $BUILD_DIR

    # Run linters first (optional but recommended)
    log "Running linters..."
    go vet ./...
    golangci-lint run ./...

    # Build and test
    main_build
    main_test

    log "CI pipeline completed successfully"
}

# Command line interface
case "${1:-build}" in
    "build")
        main_build
        ;;
    "test")
        main_test
        ;;
    "ci")
        ci_pipeline
        ;;
    "clean")
        rm -rf $BUILD_DIR
        log "Cleaned build directory"
        ;;
    *)
        echo "Usage: $0 {build|test|ci|clean}"
        exit 1
        ;;
esac

log "Process completed"