#!/bin/bash

# Multiple Wallets Test Script for Solana Balance API
# This script tests the new functionality to get balances for multiple wallets in one request

API_URL="http://localhost:8080/api/get-balance"
API_KEY="test-api-key-1"

echo "👛 Testing Multiple Wallets Functionality"
echo "=========================================="

# Test 1: Single wallet (should work as before)
echo -e "\n1. Testing single wallet:"
curl -s -X POST "$API_URL" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: $API_KEY" \
  -d '{
    "wallets": ["9WzDXwBbmkg8ZTbNMqUxvQRAyrZzDsGYdLVL9zYtAWWM"]
  }' | jq '.'

# Test 2: Multiple wallets
echo -e "\n2. Testing multiple wallets:"
curl -s -X POST "$API_URL" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: $API_KEY" \
  -d '{
    "wallets": [
      "9WzDXwBbmkg8ZTbNMqUxvQRAyrZzDsGYdLVL9zYtAWWM",
      "11111111111111111111111111111112",
      "So11111111111111111111111111111111111111112"
    ]
  }' | jq '.'

# Test 3: Mix of valid and invalid addresses
echo -e "\n3. Testing mix of valid and invalid addresses:"
curl -s -X POST "$API_URL" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: $API_KEY" \
  -d '{
    "wallets": [
      "9WzDXwBbmkg8ZTbNMqUxvQRAyrZzDsGYdLVL9zYtAWWM",
      "invalid-address",
      "11111111111111111111111111111112"
    ]
  }' | jq '.'

# Test 4: Empty wallets array
echo -e "\n4. Testing empty wallets array:"
curl -s -X POST "$API_URL" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: $API_KEY" \
  -d '{
    "wallets": []
  }' | jq '.'

# Test 5: Missing wallets field
echo -e "\n5. Testing missing wallets field:"
curl -s -X POST "$API_URL" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: $API_KEY" \
  -d '{
    "address": "9WzDXwBbmkg8ZTbNMqUxvQRAyrZzDsGYdLVL9zYtAWWM"
  }' | jq '.'

# Test 6: Large number of wallets (test limit)
echo -e "\n6. Testing large number of wallets (should be limited to 100):"
# Create an array with 101 wallet addresses
wallets_json="["
for i in {1..101}; do
    if [ $i -eq 1 ]; then
        wallets_json="${wallets_json}\"9WzDXwBbmkg8ZTbNMqUxvQRAyrZzDsGYdLVL9zYtAWWM\""
    else
        wallets_json="${wallets_json},\"9WzDXwBbmkg8ZTbNMqUxvQRAyrZzDsGYdLVL9zYtAWWM\""
    fi
done
wallets_json="${wallets_json}]"

curl -s -X POST "$API_URL" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: $API_KEY" \
  -d "{\"wallets\": $wallets_json}" | jq '.'

# Test 7: Cache test with multiple wallets
echo -e "\n7. Testing cache functionality with multiple wallets:"
echo "First request:"
curl -s -X POST "$API_URL" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: $API_KEY" \
  -d '{
    "wallets": [
      "9WzDXwBbmkg8ZTbNMqUxvQRAyrZzDsGYdLVL9zYtAWWM",
      "11111111111111111111111111111112"
    ]
  }' | jq '.'

echo -e "\nSecond request (should use cache):"
curl -s -X POST "$API_URL" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: $API_KEY" \
  -d '{
    "wallets": [
      "9WzDXwBbmkg8ZTbNMqUxvQRAyrZzDsGYdLVL9zYtAWWM",
      "11111111111111111111111111111112"
    ]
  }' | jq '.'

echo -e "\n✅ Multiple wallets test completed!" 