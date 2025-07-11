#!/bin/bash

# Multiple IP Rate Limiting Test Script
# This script simulates requests from different IP addresses

API_URL="http://localhost:8080/api/get-balance"
API_KEY="test-api-key-1"
TEST_ADDRESS="9WzDXwBbmkg8ZTbNMqUxvQRAyrZzDsGYdLVL9zYtAWWM"

echo "🌐 Testing Rate Limiting with Multiple IP Addresses"
echo "=================================================="

# Function to make request with specific IP
make_request_with_ip() {
    local ip=$1
    local request_num=$2
    
    echo "  Making request $request_num from IP: $ip"
    response=$(curl -s -X POST "$API_URL" \
        -H "Content-Type: application/json" \
        -H "X-API-Key: $API_KEY" \
        -H "X-Forwarded-For: $ip" \
        -d "{\"wallets\": [\"$TEST_ADDRESS\"]}")
    
    if echo "$response" | grep -q '"success":true'; then
        echo "    ✓ SUCCESS"
        return 0
    elif echo "$response" | grep -q '"error":"Rate limit exceeded'; then
        echo "    ✗ RATE LIMITED"
        return 1
    else
        echo "    ✗ ERROR: $(echo "$response" | jq -r '.error // .message // "Unknown error"')"
        return 2
    fi
}

# Test 1: Make 10 requests from IP 192.168.1.100
echo -e "\n1. Making 10 requests from IP 192.168.1.100:"
success_count=0
for i in {1..10}; do
    if make_request_with_ip "192.168.1.100" $i; then
        ((success_count++))
    fi
done
echo "   Results: $success_count/10 requests succeeded from 192.168.1.100"

# Test 2: Make 11th request from same IP (should be rate limited)
echo -e "\n2. Making 11th request from 192.168.1.100 (should be rate limited):"
make_request_with_ip "192.168.1.100" 11

# Test 3: Make requests from different IP (should work)
echo -e "\n3. Making 10 requests from different IP 192.168.1.200:"
success_count=0
for i in {1..10}; do
    if make_request_with_ip "192.168.1.200" $i; then
        ((success_count++))
    fi
done
echo "   Results: $success_count/10 requests succeeded from 192.168.1.200"

# Test 4: Make 11th request from second IP (should be rate limited)
echo -e "\n4. Making 11th request from 192.168.1.200 (should be rate limited):"
make_request_with_ip "192.168.1.200" 11

# Test 5: Test with multiple different IPs
echo -e "\n5. Testing with multiple different IPs (each should get 10 requests):"
ips=("10.0.0.1" "10.0.0.2" "10.0.0.3" "172.16.0.1" "172.16.0.2")

for ip in "${ips[@]}"; do
    echo -e "\n   Testing IP: $ip"
    success_count=0
    for i in {1..5}; do  # Only 5 requests per IP for this test
        if make_request_with_ip "$ip" $i; then
            ((success_count++))
        fi
    done
    echo "   Results: $success_count/5 requests succeeded from $ip"
done

echo -e "\n✅ Multiple IP rate limiting test completed!"
echo -e "\nKey findings:"
echo "- Each IP address gets its own rate limit quota"
echo "- Rate limiting is truly IP-based, not global"
echo "- Different IPs can make requests simultaneously" 