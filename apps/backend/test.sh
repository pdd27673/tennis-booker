#!/bin/bash

# Test script for tennis-booker backend
# This script runs all tests while skipping MongoDB integration tests to prevent hanging

echo "Running backend tests (skipping MongoDB integration tests)..."
echo "=================================================="

# Set environment variable to skip MongoDB tests
export SKIP_MONGODB_TESTS=true

# Run tests with verbose output
go test ./internal/... -v

echo ""
echo "=================================================="
echo "Test run complete!"
echo ""
echo "Note: MongoDB integration tests were skipped."
echo "To run MongoDB tests, ensure MongoDB is running and use:"
echo "  go test ./internal/database -v"
echo ""
echo "To run all tests including integration tests:"
echo "  unset SKIP_MONGODB_TESTS && go test ./internal/... -v" 