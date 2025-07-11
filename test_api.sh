#!/bin/bash

# Test script for Solana Balance API
# Make sure the API is running on localhost:8080

API_URL="http://localhost:8080/api/get-balance"
API_KEY="test-api-key-1"

echo "🧪 Testing Solana Balance API"
echo "=============================="

# Test 1: Valid request
echo -e "\n1. Testing valid request..."
curl -s -X POST "$API_URL" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: $API_KEY" \
  -d '{
    "wallets": ["9WzDXwBbmkg8ZTbNMqUxvQRAyrZzDsGYdLVL9zYtAWWM"]
  }' | jq '.'

# Test 2: Missing API key
echo -e "\n2. Testing missing API key..."
curl -s -X POST "$API_URL" \
  -H "Content-Type: application/json" \
  -d '{
    "wallets": ["9WzDXwBbmkg8ZTbNMqUxvQRAyrZzDsGYdLVL9zYtAWWM"]
  }' | jq '.'

# Test 3: Invalid API key
echo -e "\n3. Testing invalid API key..."
curl -s -X POST "$API_URL" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: invalid-key" \
  -d '{
    "wallets": ["9WzDXwBbmkg8ZTbNMqUxvQRAyrZzDsGYdLVL9zYtAWWM"]
  }' | jq '.'

# Test 4: Invalid address
echo -e "\n4. Testing invalid address..."
curl -s -X POST "$API_URL" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: $API_KEY" \
  -d '{
    "wallets": ["invalid-address"]
  }' | jq '.'

# Test 5: Missing address
echo -e "\n5. Testing missing address..."
curl -s -X POST "$API_URL" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: $API_KEY" \
  -d '{}' | jq '.'

# Test 6: Cache test (same address twice)
echo -e "\n6. Testing cache functionality (same address twice)..."
echo "First request:"
curl -s -X POST "$API_URL" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: $API_KEY" \
  -d '{
    "wallets": ["11111111111111111111111111111112"]
  }' | jq '.'

echo -e "\nSecond request (should be cached):"
curl -s -X POST "$API_URL" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: $API_KEY" \
  -d '{
    "wallets": ["11111111111111111111111111111112"]
  }' | jq '.'

# Test 7: Different API key
echo -e "\n7. Testing different API key..."
curl -s -X POST "$API_URL" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: test-api-key-2" \
  -d '{
    "wallets": ["9WzDXwBbmkg8ZTbNMqUxvQRAyrZzDsGYdLVL9zYtAWWM"]
  }' | jq '.'

echo -e "\n✅ Testing completed!"
echo -e "\nNote: If you see rate limit errors, wait 1 minute and try again." 