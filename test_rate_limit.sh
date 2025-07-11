#!/bin/bash

# Rate Limiting Test Script for Solana Balance API
# This script tests the IP-based rate limiting (10 requests per minute)

API_URL="http://localhost:8080/api/get-balance"
API_KEY="test-api-key-1"
TEST_ADDRESS="9WzDXwBbmkg8ZTbNMqUxvQRAyrZzDsGYdLVL9zYtAWWM"

echo "🧪 Testing IP Rate Limiting (10 requests per minute)"
echo "=================================================="

# Test 1: Make exactly 10 requests (should all succeed)
echo -e "\n1. Making 10 requests (should all succeed):"
success_count=0
for i in {1..10}; do
    response=$(curl -s -X POST "$API_URL" \
        -H "Content-Type: application/json" \
        -H "X-API-Key: $API_KEY" \
        -d "{\"wallets\": [\"$TEST_ADDRESS\"]}")
    
    if echo "$response" | grep -q '"success":true'; then
        echo "  ✓ Request $i: SUCCESS"
        ((success_count++))
    else
        echo "  ✗ Request $i: FAILED - $(echo "$response" | jq -r '.error // .message // "Unknown error"')"
    fi
done

echo -e "\n   Results: $success_count/10 requests succeeded"

# Test 2: Make 11th request (should be rate limited)
echo -e "\n2. Making 11th request (should be rate limited):"
response=$(curl -s -X POST "$API_URL" \
    -H "Content-Type: application/json" \
    -H "X-API-Key: $API_KEY" \
    -d "{\"wallets\": [\"$TEST_ADDRESS\"]}")

if echo "$response" | grep -q '"error":"Rate limit exceeded'; then
    echo "  ✓ 11th request: RATE LIMITED (Expected)"
else
    echo "  ✗ 11th request: NOT RATE LIMITED (Unexpected)"
    echo "    Response: $response"
fi

# Test 3: Test with different API key (should still be rate limited by IP)
echo -e "\n3. Testing with different API key (should still be rate limited by IP):"
response=$(curl -s -X POST "$API_URL" \
    -H "Content-Type: application/json" \
    -H "X-API-Key: test-api-key-2" \
    -d "{\"wallets\": [\"$TEST_ADDRESS\"]}")

if echo "$response" | grep -q '"error":"Rate limit exceeded'; then
    echo "  ✓ Different API key: RATE LIMITED (Expected - IP-based)"
else
    echo "  ✗ Different API key: NOT RATE LIMITED (Unexpected)"
    echo "    Response: $response"
fi

# Test 4: Wait and test again
echo -e "\n4. Waiting 5 seconds and testing again (should still be rate limited):"
sleep 5
response=$(curl -s -X POST "$API_URL" \
    -H "Content-Type: application/json" \
    -H "X-API-Key: $API_KEY" \
    -d "{\"wallets\": [\"$TEST_ADDRESS\"]}")

if echo "$response" | grep -q '"error":"Rate limit exceeded'; then
    echo "  ✓ After 5 seconds: RATE LIMITED (Expected)"
else
    echo "  ✗ After 5 seconds: NOT RATE LIMITED (Unexpected)"
    echo "    Response: $response"
fi

# Test 5: Test health endpoint (should not be rate limited)
echo -e "\n5. Testing health endpoint (should not be rate limited):"
for i in {1..5}; do
    response=$(curl -s http://localhost:8080/health)
    if echo "$response" | grep -q '"status":"healthy"'; then
        echo "  ✓ Health request $i: SUCCESS (No rate limiting)"
    else
        echo "  ✗ Health request $i: FAILED"
    fi
done

echo -e "\n✅ Rate limiting test completed!"
echo -e "\nNote: Rate limit resets after 1 minute. You can test again after waiting."
echo -e "To test reset, wait 60 seconds and run this script again." 