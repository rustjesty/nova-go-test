package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/time/rate"
)

// Configuration
var (
	HeliusRPCURL = getEnv("HELIUS_RPC_URL", "https://pomaded-lithotomies-xfbhnqagbt-dedicated.helius-rpc.com/?api-key=37ba4475-8fa3-4491-875f-758894981943")
	MongoURI      = getEnv("MONGO_URI", "mongodb://localhost:27017")
	DatabaseName  = getEnv("DB_NAME", "solana_api")
	CollectionName = getEnv("COLLECTION_NAME", "api_keys")
	CacheTTL      = 10 * time.Second
	RateLimit     = 10 // requests per minute
)

// Helper function to get environment variables with defaults
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Response structures
type BalanceResponse struct {
	Success bool    `json:"success"`
	Balance float64 `json:"balance"`
	Address string  `json:"address"`
	Error   string  `json:"error,omitempty"`
}

type GetBalanceRequest struct {
	Address string `json:"address" binding:"required"`
}

// Cache structure
type CacheEntry struct {
	Balance   float64
	Timestamp time.Time
}

// Application state
type App struct {
	client     *rpc.Client
	db         *mongo.Database
	cache      map[string]CacheEntry
	cacheMutex sync.RWMutex
	rateLimiters map[string]*rate.Limiter
	rateMutex    sync.RWMutex
	requestMutex map[string]*sync.Mutex
	requestMutexMap sync.RWMutex
}

// Initialize the application
func NewApp() (*App, error) {
	// Initialize Solana RPC client
	client := rpc.New(HeliusRPCURL)

	// Initialize MongoDB connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(MongoURI))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %v", err)
	}

	// Ping the database
	err = mongoClient.Ping(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %v", err)
	}

	db := mongoClient.Database(DatabaseName)

	// Initialize collections and indexes
	collection := db.Collection(CollectionName)
	_, err = collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "api_key", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		log.Printf("Warning: failed to create index: %v", err)
	}

	return &App{
		client:        client,
		db:            db,
		cache:         make(map[string]CacheEntry),
		rateLimiters:  make(map[string]*rate.Limiter),
		requestMutex:  make(map[string]*sync.Mutex),
	}, nil
}

// Validate API key against MongoDB
func (app *App) validateAPIKey(apiKey string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := app.db.Collection(CollectionName)
	var result bson.M
	err := collection.FindOne(ctx, bson.M{"api_key": apiKey}).Decode(&result)
	return err == nil
}

// Get rate limiter for IP
func (app *App) getRateLimiter(ip string) *rate.Limiter {
	app.rateMutex.Lock()
	defer app.rateMutex.Unlock()

	if limiter, exists := app.rateLimiters[ip]; exists {
		return limiter
	}

	limiter := rate.NewLimiter(rate.Every(time.Minute/time.Duration(RateLimit)), RateLimit)
	app.rateLimiters[ip] = limiter
	return limiter
}

// Get request mutex for wallet address
func (app *App) getRequestMutex(address string) *sync.Mutex {
	app.requestMutexMap.Lock()
	defer app.requestMutexMap.Unlock()

	if mutex, exists := app.requestMutex[address]; exists {
		return mutex
	}

	mutex := &sync.Mutex{}
	app.requestMutex[address] = mutex
	return mutex
}

// Get cached balance
func (app *App) getCachedBalance(address string) (float64, bool) {
	app.cacheMutex.RLock()
	defer app.cacheMutex.RUnlock()

	if entry, exists := app.cache[address]; exists {
		if time.Since(entry.Timestamp) < CacheTTL {
			return entry.Balance, true
		}
	}
	return 0, false
}

// Set cached balance
func (app *App) setCachedBalance(address string, balance float64) {
	app.cacheMutex.Lock()
	defer app.cacheMutex.Unlock()

	app.cache[address] = CacheEntry{
		Balance:   balance,
		Timestamp: time.Now(),
	}
}

// Fetch balance from Solana
func (app *App) fetchBalanceFromSolana(address string) (float64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Parse the address
	pubKey, err := solana.PublicKeyFromBase58(address)
	if err != nil {
		return 0, fmt.Errorf("invalid address: %v", err)
	}

	// Get balance
	balance, err := app.client.GetBalance(ctx, pubKey, rpc.CommitmentConfirmed)
	if err != nil {
		return 0, fmt.Errorf("failed to get balance: %v", err)
	}

	// Convert lamports to SOL (1 SOL = 1,000,000,000 lamports)
	return float64(balance.Value) / 1_000_000_000, nil
}

// Get balance handler
func (app *App) getBalanceHandler(c *gin.Context) {
	// Get client IP for rate limiting
	clientIP := c.ClientIP()

	// Rate limiting
	limiter := app.getRateLimiter(clientIP)
	if !limiter.Allow() {
		c.JSON(http.StatusTooManyRequests, gin.H{
			"success": false,
			"error":   "Rate limit exceeded. Maximum 10 requests per minute.",
		})
		return
	}

	// Parse request
	var req GetBalanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request format: " + err.Error(),
		})
		return
	}

	// Get request mutex for this address
	mutex := app.getRequestMutex(req.Address)
	mutex.Lock()
	defer mutex.Unlock()

	// Check cache first
	if balance, cached := app.getCachedBalance(req.Address); cached {
		c.JSON(http.StatusOK, BalanceResponse{
			Success: true,
			Balance: balance,
			Address: req.Address,
		})
		return
	}

	// Fetch from Solana
	balance, err := app.fetchBalanceFromSolana(req.Address)
	if err != nil {
		c.JSON(http.StatusInternalServerError, BalanceResponse{
			Success: false,
			Address: req.Address,
			Error:   err.Error(),
		})
		return
	}

	// Cache the result
	app.setCachedBalance(req.Address, balance)

	// Return response
	c.JSON(http.StatusOK, BalanceResponse{
		Success: true,
		Balance: balance,
		Address: req.Address,
	})
}

// Middleware for API key validation
func (app *App) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "API key required",
			})
			c.Abort()
			return
		}

		if !app.validateAPIKey(apiKey) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Invalid API key",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// Health check handler
func (app *App) healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
		"service":   "solana-balance-api",
	})
}

// Setup routes
func (app *App) setupRoutes() *gin.Engine {
	router := gin.Default()

	// Health check endpoint (no authentication required)
	router.GET("/health", app.healthHandler)

	// API routes with authentication
	api := router.Group("/api")
	api.Use(app.authMiddleware())
	{
		api.POST("/get-balance", app.getBalanceHandler)
	}

	return router
}

// Cleanup old cache entries
func (app *App) cleanupCache() {
	ticker := time.NewTicker(time.Minute)
	go func() {
		for range ticker.C {
			app.cacheMutex.Lock()
			now := time.Now()
			for address, entry := range app.cache {
				if now.Sub(entry.Timestamp) > CacheTTL {
					delete(app.cache, address)
				}
			}
			app.cacheMutex.Unlock()
		}
	}()
}

func main() {
	// Initialize application
	app, err := NewApp()
	if err != nil {
		log.Fatalf("Failed to initialize app: %v", err)
	}

	// Start cache cleanup
	app.cleanupCache()

	// Setup routes
	router := app.setupRoutes()

	// Start server
	log.Println("Starting Solana Balance API server on :8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
} 