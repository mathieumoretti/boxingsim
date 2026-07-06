#!/bin/bash

# Stop and remove test environment services
echo "Stopping test environment..."
docker-compose -f docker-compose.test.yml down

echo "Test environment stopped."