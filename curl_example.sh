#!/bin/bash

# Example curl command to test the updated /get-balance endpoint
# Replace YOUR_API_KEY with your actual API key

curl -X POST http://localhost:8080/api/get-balance \
  -H "Content-Type: application/json" \
  -H "X-API-Key: YOUR_API_KEY" \
  -d '{
    "addresses": [
      "9WzDXwBbmkg8ZTbNMqUxvQRAyrZzDsGYdLVL9zYtAWWM",
      "11111111111111111111111111111112",
      "So11111111111111111111111111111111111111112"
    ]
  }'

echo ""
echo ""
echo "Expected response format:"
echo '{
  "success": true,
  "results": [
    {
      "address": "9WzDXwBbmkg8ZTbNMqUxvQRAyrZzDsGYdLVL9zYtAWWM",
      "balance": 1.23456789
    },
    {
      "address": "11111111111111111111111111111112",
      "balance": 0.0
    },
    {
      "address": "So11111111111111111111111111111111111111112",
      "balance": 0.0
    }
  ]
}' 