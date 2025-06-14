#!/bin/bash

# Test Rate Limiting Script
# This script demonstrates the rate limiting functionality

echo "🔒 Tennis Booker API Rate Limiting Test"
echo "========================================"

# Check if the test auth server is running
echo "Checking if test auth server is available..."
if ! curl -s http://localhost:8080/health > /dev/null 2>&1; then
    echo "❌ Test auth server is not running on localhost:8080"
    echo "Please start the test auth server first:"
    echo "  cd apps/backend"
    echo "  go run cmd/test-auth-server/main.go"
    exit 1
fi

echo "✅ Test auth server is available"
echo ""

# Test 1: Basic load test
echo "📊 Test 1: Basic Load Test (Health Endpoint)"
echo "--------------------------------------------"
go run scripts/load-test/main.go \
    -endpoint=/health \
    -requests=50 \
    -concurrent=5 \
    -base-url=http://localhost:8080

echo ""

# Test 2: Rate limit test
echo "🚫 Test 2: Rate Limit Test (Auth Endpoint)"
echo "------------------------------------------"
go run scripts/load-test/main.go \
    -endpoint=/login \
    -method=POST \
    -body='{"username":"testuser","password":"testpass"}' \
    -requests=20 \
    -concurrent=3 \
    -test-rate-limit=true \
    -base-url=http://localhost:8080

echo ""

# Test 3: Burst traffic test
echo "💥 Test 3: Burst Traffic Test"
echo "-----------------------------"
go run scripts/load-test/main.go \
    -endpoint=/health \
    -requests=30 \
    -concurrent=30 \
    -base-url=http://localhost:8080

echo ""

# Test 4: Duration-based test
echo "⏱️  Test 4: Duration-Based Test (30 seconds)"
echo "--------------------------------------------"
go run scripts/load-test/main.go \
    -endpoint=/health \
    -duration=30s \
    -concurrent=5 \
    -base-url=http://localhost:8080

echo ""
echo "✅ Rate limiting tests completed!"
echo ""
echo "📝 Check the test auth server logs to see rate limiting events:"
echo "   - [RATE_LIMIT_DEBUG] for successful requests"
echo "   - [RATE_LIMIT_BLOCKED] for rate-limited requests"
echo "   - [RATE_LIMIT_ERROR] for any errors" 