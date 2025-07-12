package config

import (
	"os"
	"time"
	"path/filepath"
	"log"
	"github.com/joho/godotenv"
)

var (
	RPC_ENDPOINT   string
	MongoURI       string
	DatabaseName   string
	CollectionName string
	CacheTTL       = 10 * time.Second
	RateLimit      = 10 // requests per minute
)

// LoadEnv loads environment variables from .env file
func LoadEnv() {
	// load .env from the root directory
	rootDir := filepath.Join("..", ".env")
	if err := godotenv.Load(rootDir); err != nil {
		log.Printf("Warning: Could not load .env file from %s: %v", rootDir, err)
		// Try to load from current directory as fallback
		if err := godotenv.Load(); err != nil {
			log.Printf("Warning: Could not load .env file from current directory: %v", err)
		}
	}
	
	// Initialize variables after loading .env
	RPC_ENDPOINT = GetEnv("RPC_ENDPOINT", "https://api.mainnet-beta.solana.com")
	MongoURI = GetEnv("MONGO_URI", "mongodb://localhost:27017")
	DatabaseName = GetEnv("DB_NAME", "test")
	CollectionName = GetEnv("COLLECTION_NAME", "test")
}

// GetEnv returns the value of the environment variable or a default value
func GetEnv(key, defaultValue string) string {

	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
} 