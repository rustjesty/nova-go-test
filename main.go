package main

import (
	"log"

	"github.com/gagliardetto/solana-go/rpc"
	"solana-balance-api/cache"
	"solana-balance-api/config"
	"solana-balance-api/database"
	"solana-balance-api/rate_limiter"
	"solana-balance-api/routes"
	"solana-balance-api/services"
)

func main() {
	// Load environment variables from .env file
	config.LoadEnv()
	
	// Initialize MongoDB
	db, err := database.ConnectMongo()
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	// Initialize Solana RPC client
	client := rpc.New(config.RPC_ENDPOINT)

	// Initialize cache
	c := cache.NewCache()
	c.Cleanup()

	// Initialize rate limiter (if you want to use it in middleware)
	_ = rate_limiter.NewRateLimiter()

	// Initialize balance service
	balanceService := services.NewBalanceService(client, c)

	// Setup routes
	router := routes.SetupRoutes(db, balanceService)

	// Start server
	log.Println("Starting Solana Balance API server on :8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
} 