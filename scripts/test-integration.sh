#!/bin/bash

# Backend-Frontend E2E Integration Test Script
# Tests all integrated API endpoints and functionality

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
BACKEND_URL=${BACKEND_URL:-"http://localhost:8080"}
FRONTEND_URL=${FRONTEND_URL:-"http://localhost:5173"}
TEST_USER_EMAIL="test-integration@example.com"
TEST_USER_PASSWORD="TestPassword123!"

# Test results tracking
TESTS_PASSED=0
TESTS_FAILED=0
TEST_OUTPUT=""

# Helper functions
log() {
    echo -e "${BLUE}[TEST]${NC} $1"
}

success() {
    echo -e "${GREEN}✅ $1${NC}"
    ((TESTS_PASSED++))
    TEST_OUTPUT+="\n✅ $1"
}

error() {
    echo -e "${RED}❌ $1${NC}"
    ((TESTS_FAILED++))
    TEST_OUTPUT+="\n❌ $1"
}

warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

# Test helper function
test_endpoint() {
    local name="$1"
    local method="$2"
    local endpoint="$3"
    local headers="$4"
    local data="$5"
    local expected_status="${6:-200}"
    
    log "Testing $name: $method $endpoint"
    
    local response
    local status_code
    
    if [ -n "$data" ]; then
        response=$(curl -s -w "\n%{http_code}" -X "$method" "$BACKEND_URL$endpoint" \
            -H "Content-Type: application/json" \
            $headers \
            -d "$data" 2>/dev/null || echo -e "\n000")
    else
        response=$(curl -s -w "\n%{http_code}" -X "$method" "$BACKEND_URL$endpoint" \
            -H "Content-Type: application/json" \
            $headers 2>/dev/null || echo -e "\n000")
    fi
    
    status_code=$(echo "$response" | tail -n1)
    response_body=$(echo "$response" | head -n -1)
    
    if [ "$status_code" = "$expected_status" ]; then
        success "$name - Status: $status_code"
        return 0
    else
        error "$name - Expected: $expected_status, Got: $status_code"
        if [ "$status_code" != "000" ]; then
            echo "Response: $response_body"
        fi
        return 1
    fi
}

# Main test functions
test_backend_health() {
    log "Testing Backend Health & Connectivity"
    
    if test_endpoint "Health Check" "GET" "/api/health" "" "" "200"; then
        success "Backend is responding"
    else
        error "Backend health check failed"
        return 1
    fi
}

test_authentication_flow() {
    log "Testing Authentication Flow"
    
    # Test user registration
    local register_data="{\"username\":\"$TEST_USER_EMAIL\",\"email\":\"$TEST_USER_EMAIL\",\"password\":\"$TEST_USER_PASSWORD\"}"
    
    # Try to register (might fail if user exists)
    test_endpoint "User Registration" "POST" "/api/auth/register" "" "$register_data" "201" || \
    test_endpoint "User Registration (existing)" "POST" "/api/auth/register" "" "$register_data" "400"
    
    # Test user login
    local login_data="{\"username\":\"$TEST_USER_EMAIL\",\"password\":\"$TEST_USER_PASSWORD\"}"
    local login_response
    login_response=$(curl -s -X POST "$BACKEND_URL/api/auth/login" \
        -H "Content-Type: application/json" \
        -d "$login_data" 2>/dev/null)
    
    if echo "$login_response" | grep -q "token"; then
        success "User login successful"
        
        # Extract token for subsequent tests
        ACCESS_TOKEN=$(echo "$login_response" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
        REFRESH_TOKEN=$(echo "$login_response" | grep -o '"refresh_token":"[^"]*"' | cut -d'"' -f4)
        
        if [ -n "$ACCESS_TOKEN" ]; then
            success "Access token extracted"
            
            # Test authenticated endpoint
            test_endpoint "Get User Info" "GET" "/api/auth/me" "-H \"Authorization: Bearer $ACCESS_TOKEN\"" "" "200"
            
            # Test token refresh
            local refresh_data="{\"refresh_token\":\"$REFRESH_TOKEN\"}"
            test_endpoint "Token Refresh" "POST" "/api/auth/refresh" "" "$refresh_data" "200"
            
        else
            error "Failed to extract access token"
        fi
    else
        error "User login failed"
    fi
}

test_court_data_endpoints() {
    log "Testing Court Data Endpoints"
    
    if [ -z "$ACCESS_TOKEN" ]; then
        warning "Skipping court data tests - no access token available"
        return
    fi
    
    local auth_header="-H \"Authorization: Bearer $ACCESS_TOKEN\""
    
    # Test venues endpoint
    test_endpoint "Get Venues" "GET" "/api/venues" "$auth_header" "" "200"
    
    # Test court slots endpoint
    test_endpoint "Get Court Slots" "GET" "/api/court-slots" "$auth_header" "" "200"
    
    # Test dashboard stats (if available)
    test_endpoint "Get Dashboard Stats" "GET" "/api/dashboard/stats" "$auth_header" "" "200" || \
    warning "Dashboard stats endpoint not available"
}

test_system_control() {
    log "Testing System Control Endpoints"
    
    if [ -z "$ACCESS_TOKEN" ]; then
        warning "Skipping system control tests - no access token available"
        return
    fi
    
    local auth_header="-H \"Authorization: Bearer $ACCESS_TOKEN\""
    
    # Test system status
    test_endpoint "Get System Status" "GET" "/api/system/status" "$auth_header" "" "200"
    
    # Test system pause
    test_endpoint "Pause System" "POST" "/api/system/pause" "$auth_header" "" "200" || \
    test_endpoint "Pause System (already paused)" "POST" "/api/system/pause" "$auth_header" "" "400"
    
    # Test system resume
    test_endpoint "Resume System" "POST" "/api/system/resume" "$auth_header" "" "200" || \
    test_endpoint "Resume System (already running)" "POST" "/api/system/resume" "$auth_header" "" "400"
}

test_user_preferences() {
    log "Testing User Preferences Endpoints"
    
    if [ -z "$ACCESS_TOKEN" ]; then
        warning "Skipping user preferences tests - no access token available"
        return
    fi
    
    local auth_header="-H \"Authorization: Bearer $ACCESS_TOKEN\""
    
    # Test get user preferences
    test_endpoint "Get User Preferences" "GET" "/api/user/preferences" "$auth_header" "" "200"
    
    # Test update user preferences
    local prefs_data="{\"notification_preferences\":{\"email_enabled\":true,\"push_enabled\":false},\"court_preferences\":{\"preferred_venues\":[\"test-venue\"],\"preferred_times\":[\"morning\"]}}"
    test_endpoint "Update User Preferences" "PUT" "/api/user/preferences" "$auth_header" "$prefs_data" "200"
}

test_frontend_accessibility() {
    log "Testing Frontend Accessibility"
    
    # Check if frontend is running
    if curl -s -f "$FRONTEND_URL" > /dev/null 2>&1; then
        success "Frontend is accessible at $FRONTEND_URL"
    else
        error "Frontend is not accessible at $FRONTEND_URL"
        warning "Make sure frontend is running: cd apps/frontend && npm run dev"
    fi
    
    # Check if frontend can reach backend (CORS test)
    # This would need to be done in a browser context, so we'll skip for now
    warning "CORS testing requires browser environment - manual testing recommended"
}

print_summary() {
    echo ""
    echo "=================================="
    echo "INTEGRATION TEST SUMMARY"
    echo "=================================="
    echo -e "Tests Passed: ${GREEN}$TESTS_PASSED${NC}"
    echo -e "Tests Failed: ${RED}$TESTS_FAILED${NC}"
    echo ""
    
    if [ $TESTS_FAILED -eq 0 ]; then
        echo -e "${GREEN}🎉 All tests passed! Backend-Frontend integration is working correctly.${NC}"
    else
        echo -e "${RED}⚠️  Some tests failed. Please check the issues above.${NC}"
    fi
    
    echo ""
    echo "Test Results:"
    echo -e "$TEST_OUTPUT"
    echo ""
    
    if [ $TESTS_FAILED -eq 0 ]; then
        echo "✅ Integration is ready for development!"
        echo ""
        echo "Next steps:"
        echo "1. Start frontend: cd apps/frontend && npm run dev"
        echo "2. Start backend: cd apps/backend && make run"
        echo "3. Open http://localhost:5173 in your browser"
        echo "4. Register/login and test the UI"
    else
        echo "❌ Fix the failing tests before proceeding"
        echo ""
        echo "Common fixes:"
        echo "1. Ensure backend is running: cd apps/backend && make run"
        echo "2. Check environment variables are set correctly"
        echo "3. Verify database connections (if using external DB)"
        echo "4. Check JWT secret is configured"
    fi
}

# Main execution
main() {
    echo "🧪 Backend-Frontend E2E Integration Test"
    echo "========================================="
    echo "Backend URL: $BACKEND_URL"
    echo "Frontend URL: $FRONTEND_URL"
    echo ""
    
    # Initialize variables
    ACCESS_TOKEN=""
    REFRESH_TOKEN=""
    
    # Run all tests
    test_backend_health
    test_authentication_flow
    test_court_data_endpoints
    test_system_control
    test_user_preferences
    test_frontend_accessibility
    
    # Print summary
    print_summary
    
    # Exit with appropriate code
    if [ $TESTS_FAILED -eq 0 ]; then
        exit 0
    else
        exit 1
    fi
}

# Handle script arguments
case "${1:-}" in
    --help|-h)
        echo "Usage: $0 [--backend-url URL] [--frontend-url URL]"
        echo ""
        echo "Options:"
        echo "  --backend-url URL   Backend URL (default: http://localhost:8080)"
        echo "  --frontend-url URL  Frontend URL (default: http://localhost:5173)"
        echo "  --help, -h          Show this help"
        echo ""
        echo "Environment Variables:"
        echo "  BACKEND_URL         Backend URL"
        echo "  FRONTEND_URL        Frontend URL"
        exit 0
        ;;
    --backend-url)
        BACKEND_URL="$2"
        shift 2
        ;;
    --frontend-url)
        FRONTEND_URL="$2"
        shift 2
        ;;
    "")
        # No arguments, run normally
        ;;
    *)
        echo "Unknown argument: $1"
        echo "Use --help for usage information"
        exit 1
        ;;
esac

# Run main function
main "$@" 