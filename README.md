# Solana Balance API

A high-performance Golang REST API for fetching Solana wallet balances with rate limiting, caching, and MongoDB authentication.

## Features

- ✅ **Rate Limiting**: 10 requests per minute per IP address
- ✅ **Caching**: 10-second TTL cache for wallet addresses
- ✅ **Concurrent Request Handling**: Mutex-based synchronization for same wallet addresses
- ✅ **MongoDB Authentication**: API key validation against database
- ✅ **High Performance**: Optimized for fast response times
- ✅ **Solana Integration**: Uses Helius RPC endpoint for reliable balance fetching

## Prerequisites

- Go 1.21 or higher
- MongoDB (running on localhost:27017 by default)
- Internet connection for Solana RPC calls

## Installation

1. **Install dependencies**:
   ```bash
   go mod tidy
   ```

2. **Set up MongoDB** (if not already running):
   ```bash
   # Install MongoDB (Ubuntu/Debian)
   sudo apt update
   sudo apt install mongodb
   sudo systemctl start mongodb
   sudo systemctl enable mongodb
   ```

3. **Set up the database and API keys**:
   ```bash
   go run setup_db.go
   ```

4. **Run the API server**:
   ```bash
   go run main.go
   ```

The server will start on `http://localhost:8080`

## API Documentation

### POST /api/get-balance

Fetches the Solana balance for a given wallet address.

**Headers**:
- `Content-Type: application/json`
- `X-API-Key: your-api-key` (required)

**Request Body**:
```json
{
  "address": "9WzDXwBbmkg8ZTbNMqUxvQRAyrZzDsGYdLVL9zYtAWWM"
}
```

**Response**:
```json
{
  "success": true,
  "balance": 1.23456789,
  "address": "9WzDXwBbmkg8ZTbNMqUxvQRAyrZzDsGYdLVL9zYtAWWM"
}
```

**Error Response**:
```json
{
  "success": false,
  "error": "Invalid address: invalid base58 string"
}
```

## Usage Examples

### Using curl

```bash
# Get balance for a wallet
curl -X POST http://localhost:8080/api/get-balance \
  -H "Content-Type: application/json" \
  -H "X-API-Key: test-api-key-1" \
  -d '{
    "address": "9WzDXwBbmkg8ZTbNMqUxvQRAyrZzDsGYdLVL9zYtAWWM"
  }'
```

### Using JavaScript/Node.js

```javascript
const response = await fetch('http://localhost:8080/api/get-balance', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'X-API-Key': 'test-api-key-1'
  },
  body: JSON.stringify({
    address: '9WzDXwBbmkg8ZTbNMqUxvQRAyrZzDsGYdLVL9zYtAWWM'
  })
});

const data = await response.json();
console.log(data);
```

### Using Python

```python
import requests

response = requests.post(
    'http://localhost:8080/api/get-balance',
    headers={
        'Content-Type': 'application/json',
        'X-API-Key': 'test-api-key-1'
    },
    json={
        'address': '9WzDXwBbmkg8ZTbNMqUxvQRAyrZzDsGYdLVL9zYtAWWM'
    }
)

print(response.json())
```

## Configuration

You can modify the following constants in `main.go`:

- `HeliusRPCURL`: Solana RPC endpoint
- `MongoURI`: MongoDB connection string
- `DatabaseName`: MongoDB database name
- `CollectionName`: MongoDB collection name for API keys
- `CacheTTL`: Cache time-to-live (default: 10 seconds)
- `RateLimit`: Rate limit per minute (default: 10 requests)

## Architecture

### Components

1. **Rate Limiting**: Uses `golang.org/x/time/rate` for IP-based rate limiting
2. **Caching**: In-memory cache with TTL for wallet addresses
3. **Concurrency Control**: Mutex-based synchronization for same wallet addresses
4. **Authentication**: MongoDB-based API key validation
5. **Solana Integration**: Uses `gagliardetto/solana-go` for RPC calls

### Flow

1. Request comes in with API key
2. API key is validated against MongoDB
3. Rate limiting is checked for client IP
4. If same wallet address is being processed, wait for completion
5. Check cache for existing balance
6. If not cached, fetch from Solana RPC
7. Cache the result and return response

## Error Handling

The API returns appropriate HTTP status codes:

- `200 OK`: Successful balance fetch
- `400 Bad Request`: Invalid request format
- `401 Unauthorized`: Missing or invalid API key
- `429 Too Many Requests`: Rate limit exceeded
- `500 Internal Server Error`: Solana RPC error or other server issues

## Performance Optimizations

1. **Connection Pooling**: MongoDB connection reuse
2. **Read-Write Mutex**: Efficient cache access patterns
3. **Background Cleanup**: Automatic cache cleanup every minute
4. **Timeout Handling**: Context-based timeouts for external calls
5. **Memory Management**: Efficient data structures and cleanup

## Testing

### Test API Keys

The setup script creates these test API keys:
- `test-api-key-1`
- `test-api-key-2`
- `demo-api-key-123`

### Test Wallet Addresses

You can test with these Solana addresses:
- `9WzDXwBbmkg8ZTbNMqUxvQRAyrZzDsGYdLVL9zYtAWWM` (Solana Foundation)
- `11111111111111111111111111111112` (System Program)

## Monitoring

The API logs important events:
- Server startup
- MongoDB connection status
- Cache cleanup operations
- API key validation warnings

## Security Considerations

1. **API Key Storage**: API keys are stored securely in MongoDB
2. **Rate Limiting**: Prevents abuse and ensures fair usage
3. **Input Validation**: All inputs are validated before processing
4. **Error Handling**: Sensitive information is not exposed in error messages

## Troubleshooting

### Common Issues

1. **MongoDB Connection Failed**:
   - Ensure MongoDB is running: `sudo systemctl status mongodb`
   - Check connection string in `main.go`

2. **Rate Limit Exceeded**:
   - Wait for the rate limit window to reset (1 minute)
   - Use different IP addresses for testing

3. **Invalid Address Error**:
   - Ensure the Solana address is valid base58 format
   - Check for typos in the address

4. **API Key Not Found**:
   - Run `go run setup_db.go` to create test API keys
   - Check the `X-API-Key` header in your request

## License

This project is open source and available under the MIT License.
