package routes

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"solana-balance-api/handlers"
	"solana-balance-api/middleware"
	"solana-balance-api/services"
)

func SetupRoutes(db *mongo.Database, balanceService *services.BalanceService) *gin.Engine {
	router := gin.Default()

	// Health check endpoint (no authentication required)
	router.GET("/", handlers.HealthHandler())

	// API routes with authentication
	api := router.Group("/api")
	api.Use(middleware.AuthMiddleware(db))
	{
		api.POST("/get-balance", handlers.GetBalanceHandler(balanceService))
	}

	return router
} 