# Quick Start Guide

Get the Solana Balance API running in minutes!

## Prerequisites

- Go 1.21+ installed
- MongoDB running (or Docker for containerized setup)

## Option 1: Local Setup (Recommended for Development)

### 1. Install Dependencies
```bash
cd go-test
go mod tidy
```

### 2. Start MongoDB
```bash
# Ubuntu/Debian
sudo systemctl start mongodb

# Or using Docker
docker run -d -p 27017:27017 --name mongodb mongo:6.0
```

### 3. Set Up Database and API Keys
```bash
go run setup_db.go
```

### 4. Start the API Server
```bash
go run main.go
```

The API will be available at `http://localhost:8080`

## Option 2: Docker Setup (Recommended for Production)

### 1. Build and Run with Docker Compose
```bash
cd go-test
docker-compose up -d
```

This will:
- Start MongoDB container
- Build and start the API container
- Set up the database automatically

### 2. Check Status
```bash
docker-compose ps
docker-compose logs -f
```

## Testing the API

### 1. Health Check
```bash
curl http://localhost:8080/health
```

### 2. Get Solana Balance
```bash
curl -X POST http://localhost:8080/api/get-balance \
  -H "Content-Type: application/json" \
  -H "X-API-Key: test-api-key-1" \
  -d '{
    "address": "9WzDXwBbmkg8ZTbNMqUxvQRAyrZzDsGYdLVL9zYtAWWM"
  }'
```

### 3. Run All Tests
```bash
./test_api.sh
```

## Available API Keys

The setup script creates these test API keys:
- `test-api-key-1`
- `test-api-key-2`
- `demo-api-key-123`

## Troubleshooting

### MongoDB Connection Issues
```bash
# Check if MongoDB is running
sudo systemctl status mongodb

# Check MongoDB logs
sudo journalctl -u mongodb -f
```

### Port Already in Use
```bash
# Check what's using port 8080
sudo lsof -i :8080

# Kill the process if needed
sudo kill -9 <PID>
```

### Docker Issues
```bash
# Clean up Docker containers
docker-compose down
docker system prune -f

# Rebuild and start
docker-compose up --build -d
```

## Next Steps

1. Read the full [README.md](README.md) for detailed documentation
2. Check the [Makefile](Makefile) for additional commands
3. Customize configuration in [.env.example](env.example)
4. Add your own API keys to the database

## Support

If you encounter any issues:
1. Check the logs: `docker-compose logs -f` or console output
2. Verify MongoDB is running and accessible
3. Ensure all dependencies are installed
4. Check the troubleshooting section in the main README 