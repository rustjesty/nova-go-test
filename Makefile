.PHONY: help build run test setup clean docker-build docker-run

# Default target
help:
	@echo "Available commands:"
	@echo "  setup      - Set up MongoDB and create API keys"
	@echo "  build      - Build the application"
	@echo "  run        - Run the API server"
	@echo "  test       - Run API tests"
	@echo "  clean      - Clean build artifacts"
	@echo "  docker-build - Build Docker image"
	@echo "  docker-run   - Run with Docker Compose"

# Set up the database and API keys
setup:
	@echo "Setting up MongoDB and API keys..."
	go run setup_db.go

# Build the application
build:
	@echo "Building the application..."
	go build -o solana-api main.go

# Run the API server
run:
	@echo "Starting Solana Balance API server..."
	go run main.go

# Run API tests
test:
	@echo "Running API tests..."
	./test_api.sh

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -f solana-api
	go clean

# Build Docker image
docker-build:
	@echo "Building Docker image..."
	docker build -t solana-balance-api .

# Run with Docker Compose
docker-run:
	@echo "Starting services with Docker Compose..."
	docker-compose up -d

# Stop Docker Compose services
docker-stop:
	@echo "Stopping Docker Compose services..."
	docker-compose down

# View Docker Compose logs
docker-logs:
	@echo "Viewing Docker Compose logs..."
	docker-compose logs -f

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod tidy
	go mod download 