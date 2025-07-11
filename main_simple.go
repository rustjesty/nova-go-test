package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"golang.org/x/time/rate"
)

// Configuration
var (
	HeliusRPCURL = "https://pomaded-lithotomies-xfbhnqagbt-dedicated.helius-rpc.com/?api-key=37ba4475-8fa3-4491-875f-758894981943"
	CacheTTL     = 10 * time.Second
	RateLimit    = 10 // requests per minute
)

// Response structures

type GetBalanceRequest struct {
	Wallets []string `json:"wallets" binding:"required"`
}

type WalletBalance struct {
	Address string  `json:"address"`
	Balance float64 `json:"balance"`
	Error   string  `json:"error,omitempty"`
}

type BalanceResponse struct {
	Success bool            `json:"success"`
	Wallets []WalletBalance `json:"wallets"`
	Error   string          `json:"error,omitempty"`
}

// Cache structure
type CacheEntry struct {
	Balance   float64
	Timestamp time.Time
}

// Application state
type App struct {
	client          *rpc.Client
	cache           map[string]CacheEntry
	cacheMutex      sync.RWMutex
	rateLimiters    map[string]*rate.Limiter
	rateMutex       sync.RWMutex
	requestMutex    map[string]*sync.Mutex
	requestMutexMap sync.RWMutex
	validAPIKeys    map[string]bool
}

// Initialize the application
func NewApp() (*App, error) {
	// Initialize Solana RPC client
	client := rpc.New(HeliusRPCURL)

	// Initialize valid API keys (for testing)
	validAPIKeys := map[string]bool{
		"test-api-key-1": true,
		"test-api-key-2": true,
		"demo-api-key-123": true,
	}

	return &App{
		client:       client,
		cache:        make(map[string]CacheEntry),
		rateLimiters: make(map[string]*rate.Limiter),
		requestMutex: make(map[string]*sync.Mutex),
		validAPIKeys: validAPIKeys,
	}, nil
}

// Validate API key (in-memory for testing)
func (app *App) validateAPIKey(apiKey string) bool {
	return app.validAPIKeys[apiKey]
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

	// Validate wallets array
	if len(req.Wallets) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "At least one wallet address is required",
		})
		return
	}

	// Limit number of wallets per request (optional, for performance)
	if len(req.Wallets) > 100 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Maximum 100 wallet addresses per request",
		})
		return
	}

	var walletBalances []WalletBalance

	// Process each wallet address
	for _, address := range req.Wallets {
		// Get request mutex for this address
		mutex := app.getRequestMutex(address)
		mutex.Lock()

		// Check cache first
		if balance, cached := app.getCachedBalance(address); cached {
			walletBalances = append(walletBalances, WalletBalance{
				Address: address,
				Balance: balance,
			})
			mutex.Unlock()
			continue
		}

		// Fetch from Solana
		balance, err := app.fetchBalanceFromSolana(address)
		mutex.Unlock()

		if err != nil {
			walletBalances = append(walletBalances, WalletBalance{
				Address: address,
				Error:   err.Error(),
			})
		} else {
			// Cache the result
			app.setCachedBalance(address, balance)
			walletBalances = append(walletBalances, WalletBalance{
				Address: address,
				Balance: balance,
			})
		}
	}

	// Return response
	c.JSON(http.StatusOK, BalanceResponse{
		Success: true,
		Wallets: walletBalances,
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
		"service":   "solana-balance-api-simple",
		"note":      "Running in simple mode without MongoDB",
	})
}

// Setup routes
func (app *App) setupRoutes() *gin.Engine {
	router := gin.Default()

	// Health check endpoint (no authentication required)
	router.GET("/", app.healthHandler)

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
	log.Println("Starting Solana Balance API server (Simple Mode) on http://localhost:8080")
	log.Println("Available API keys: test-api-key-1, test-api-key-2, demo-api-key-123")
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
} 