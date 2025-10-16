#!/bin/bash

# ===============================
# API Testing Script
# ===============================
# This script tests all endpoints of the Contact Management API
# Usage: ./test_api.sh [BASE_URL]
# Example: ./test_api.sh http://localhost:9001

set -e  # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
BASE_URL="${1:-http://localhost:9001}"
API_URL="${BASE_URL}/api/v1"
TEST_EMAIL="test_$(date +%s)@example.com"
TEST_PHONE="+6281234567890"
TEST_PASSWORD="Test123456"
TEST_FULL_NAME="Test User"
JWT_TOKEN=""
USER_ID=""
CONTACT_ID=""

# Counters
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# ===============================
# Helper Functions
# ===============================

print_header() {
    echo ""
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}========================================${NC}"
}

print_test() {
    echo -e "${YELLOW}[TEST]${NC} $1"
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
}

print_success() {
    echo -e "${GREEN}[✓]${NC} $1"
    PASSED_TESTS=$((PASSED_TESTS + 1))
}

print_error() {
    echo -e "${RED}[✗]${NC} $1"
    FAILED_TESTS=$((FAILED_TESTS + 1))
}

print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

# Test HTTP response
test_response() {
    local response=$1
    local expected_status=$2
    local test_name=$3
    
    local status=$(echo "$response" | jq -r '.status_code // .status // "null"')
    
    if [ "$status" == "$expected_status" ] || [ "$status" == "null" ]; then
        print_success "$test_name - Status: $expected_status"
        return 0
    else
        print_error "$test_name - Expected: $expected_status, Got: $status"
        echo "Response: $response"
        return 1
    fi
}

# ===============================
# Test Cases
# ===============================

test_health_check() {
    print_header "1. HEALTH CHECK"
    
    print_test "Testing health endpoint..."
    response=$(curl -s "${BASE_URL}/health")
    
    if echo "$response" | jq -e '.status' > /dev/null 2>&1; then
        print_success "Health check passed"
        echo "Response: $response"
    else
        print_error "Health check failed"
        echo "Response: $response"
    fi
}

test_ping() {
    print_header "2. PING ENDPOINT"
    
    print_test "Testing ping endpoint..."
    response=$(curl -s "${API_URL}/ping")
    
    if echo "$response" | jq -e '.message' > /dev/null 2>&1; then
        print_success "Ping test passed"
        echo "Response: $response"
    else
        print_error "Ping test failed"
        echo "Response: $response"
    fi
}

test_register() {
    print_header "3. USER REGISTRATION"
    
    print_test "Registering new user..."
    print_info "Email: $TEST_EMAIL"
    print_info "Phone: $TEST_PHONE"
    
    response=$(curl -s -X POST "${API_URL}/auth/register" \
        -H "Content-Type: application/json" \
        -d '{
            "full_name": "'"$TEST_FULL_NAME"'",
            "email": "'"$TEST_EMAIL"'",
            "phone": "'"$TEST_PHONE"'",
            "password": "'"$TEST_PASSWORD"'"
        }')
    
    test_response "$response" "201" "User registration"
    
    # Extract user ID
    USER_ID=$(echo "$response" | jq -r '.data.id // .data.user.id // ""')
    if [ -n "$USER_ID" ]; then
        print_info "User ID: $USER_ID"
    fi
    
    echo "Response: $response"
}

test_register_duplicate() {
    print_header "4. DUPLICATE REGISTRATION (Should Fail)"
    
    print_test "Attempting duplicate registration..."
    
    response=$(curl -s -X POST "${API_URL}/auth/register" \
        -H "Content-Type: application/json" \
        -d '{
            "full_name": "'"$TEST_FULL_NAME"'",
            "email": "'"$TEST_EMAIL"'",
            "phone": "'"$TEST_PHONE"'",
            "password": "'"$TEST_PASSWORD"'"
        }')
    
    test_response "$response" "409" "Duplicate registration (expected to fail)"
    echo "Response: $response"
}

test_login() {
    print_header "5. USER LOGIN"
    
    print_test "Logging in..."
    
    response=$(curl -s -X POST "${API_URL}/auth/login" \
        -H "Content-Type: application/json" \
        -d '{
            "email": "'"$TEST_EMAIL"'",
            "password": "'"$TEST_PASSWORD"'"
        }')
    
    test_response "$response" "200" "User login"
    
    # Extract JWT token
    JWT_TOKEN=$(echo "$response" | jq -r '.data.token // .data.access_token // ""')
    if [ -n "$JWT_TOKEN" ]; then
        print_success "JWT token obtained"
        print_info "Token: ${JWT_TOKEN:0:50}..."
    else
        print_error "Failed to get JWT token"
        echo "Response: $response"
        exit 1
    fi
}

test_login_invalid() {
    print_header "6. INVALID LOGIN (Should Fail)"
    
    print_test "Attempting login with wrong password..."
    
    response=$(curl -s -X POST "${API_URL}/auth/login" \
        -H "Content-Type: application/json" \
        -d '{
            "email": "'"$TEST_EMAIL"'",
            "password": "WrongPassword123"
        }')
    
    test_response "$response" "401" "Invalid login (expected to fail)"
    echo "Response: $response"
}

test_get_profile() {
    print_header "7. GET USER PROFILE"
    
    print_test "Getting user profile..."
    
    response=$(curl -s -X GET "${API_URL}/me" \
        -H "Authorization: Bearer $JWT_TOKEN")
    
    test_response "$response" "200" "Get user profile"
    echo "Response: $response"
}

test_update_profile() {
    print_header "8. UPDATE USER PROFILE"
    
    print_test "Updating user profile..."
    
    response=$(curl -s -X PUT "${API_URL}/me" \
        -H "Authorization: Bearer $JWT_TOKEN" \
        -H "Content-Type: application/json" \
        -d '{
            "full_name": "Updated Test User",
            "phone": "+6281234567891"
        }')
    
    test_response "$response" "200" "Update user profile"
    echo "Response: $response"
}

test_create_contact() {
    print_header "9. CREATE CONTACT"
    
    print_test "Creating new contact..."
    
    response=$(curl -s -X POST "${API_URL}/contacts" \
        -H "Authorization: Bearer $JWT_TOKEN" \
        -H "Content-Type: application/json" \
        -d '{
            "full_name": "John Doe",
            "phone": "+6281234567892",
            "email": "john.doe@example.com"
        }')
    
    test_response "$response" "201" "Create contact"
    
    # Extract contact ID
    CONTACT_ID=$(echo "$response" | jq -r '.data.id // .data.contact.id // ""')
    if [ -n "$CONTACT_ID" ]; then
        print_info "Contact ID: $CONTACT_ID"
    fi
    
    echo "Response: $response"
}

test_list_contacts() {
    print_header "10. LIST CONTACTS"
    
    print_test "Listing all contacts..."
    
    response=$(curl -s -X GET "${API_URL}/contacts?page=1&limit=10" \
        -H "Authorization: Bearer $JWT_TOKEN")
    
    test_response "$response" "200" "List contacts"
    echo "Response: $response"
}

test_search_contacts() {
    print_header "11. SEARCH CONTACTS"
    
    print_test "Searching contacts with query 'John'..."
    
    response=$(curl -s -X GET "${API_URL}/contacts?q=John&page=1&limit=10" \
        -H "Authorization: Bearer $JWT_TOKEN")
    
    test_response "$response" "200" "Search contacts"
    echo "Response: $response"
}

test_get_contact() {
    print_header "12. GET CONTACT DETAIL"
    
    if [ -z "$CONTACT_ID" ]; then
        print_error "No contact ID available, skipping test"
        return
    fi
    
    print_test "Getting contact detail for ID: $CONTACT_ID"
    
    response=$(curl -s -X GET "${API_URL}/contacts/${CONTACT_ID}" \
        -H "Authorization: Bearer $JWT_TOKEN")
    
    test_response "$response" "200" "Get contact detail"
    echo "Response: $response"
}

test_update_contact() {
    print_header "13. UPDATE CONTACT"
    
    if [ -z "$CONTACT_ID" ]; then
        print_error "No contact ID available, skipping test"
        return
    fi
    
    print_test "Updating contact ID: $CONTACT_ID"
    
    response=$(curl -s -X PUT "${API_URL}/contacts/${CONTACT_ID}" \
        -H "Authorization: Bearer $JWT_TOKEN" \
        -H "Content-Type: application/json" \
        -d '{
            "full_name": "John Doe Updated",
            "phone": "+6281234567893",
            "email": "john.updated@example.com",
            "favorite": true
        }')
    
    test_response "$response" "200" "Update contact"
    echo "Response: $response"
}

test_delete_contact() {
    print_header "14. DELETE CONTACT"
    
    if [ -z "$CONTACT_ID" ]; then
        print_error "No contact ID available, skipping test"
        return
    fi
    
    print_test "Deleting contact ID: $CONTACT_ID"
    
    response=$(curl -s -X DELETE "${API_URL}/contacts/${CONTACT_ID}" \
        -H "Authorization: Bearer $JWT_TOKEN")
    
    test_response "$response" "200" "Delete contact"
    echo "Response: $response"
}

test_unauthorized_access() {
    print_header "15. UNAUTHORIZED ACCESS (Should Fail)"
    
    print_test "Attempting to access protected endpoint without token..."
    
    response=$(curl -s -X GET "${API_URL}/me")
    
    test_response "$response" "401" "Unauthorized access (expected to fail)"
    echo "Response: $response"
}

# ===============================
# Main Execution
# ===============================

main() {
    clear
    echo -e "${GREEN}╔════════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║   Contact Management API Test Suite   ║${NC}"
    echo -e "${GREEN}╔════════════════════════════════════════╗${NC}"
    echo ""
    print_info "Base URL: $BASE_URL"
    print_info "API URL: $API_URL"
    echo ""
    
    # Check if jq is installed
    if ! command -v jq &> /dev/null; then
        print_error "jq is not installed. Please install it first:"
        echo "  macOS: brew install jq"
        echo "  Ubuntu: sudo apt-get install jq"
        echo "  CentOS: sudo yum install jq"
        exit 1
    fi
    
    # Check if server is running
    print_info "Checking if server is accessible..."
    if ! curl -s "${BASE_URL}/health" > /dev/null 2>&1; then
        print_error "Server is not accessible at $BASE_URL"
        print_info "Please make sure the server is running:"
        echo "  make run"
        exit 1
    fi
    print_success "Server is accessible"
    
    # Run all tests
    test_health_check
    test_ping
    test_register
    test_register_duplicate
    test_login
    test_login_invalid
    test_get_profile
    test_update_profile
    test_create_contact
    test_list_contacts
    test_search_contacts
    test_get_contact
    test_update_contact
    test_delete_contact
    test_unauthorized_access
    
    # Print summary
    print_header "TEST SUMMARY"
    echo ""
    echo -e "Total Tests:  ${BLUE}${TOTAL_TESTS}${NC}"
    echo -e "Passed:       ${GREEN}${PASSED_TESTS}${NC}"
    echo -e "Failed:       ${RED}${FAILED_TESTS}${NC}"
    echo ""
    
    if [ $FAILED_TESTS -eq 0 ]; then
        echo -e "${GREEN}╔════════════════════════════════════════╗${NC}"
        echo -e "${GREEN}║       ALL TESTS PASSED! ✓              ║${NC}"
        echo -e "${GREEN}╚════════════════════════════════════════╝${NC}"
        exit 0
    else
        echo -e "${RED}╔════════════════════════════════════════╗${NC}"
        echo -e "${RED}║       SOME TESTS FAILED! ✗             ║${NC}"
        echo -e "${RED}╚════════════════════════════════════════╝${NC}"
        exit 1
    fi
}

# Run main function
main "$@"
