package handlers

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"solana-balance-api/models"
	"solana-balance-api/services"
)

func GetBalanceHandler(balanceService *services.BalanceService) gin.HandlerFunc {
	return func(c *gin.Context) {

		var req models.GetBalanceRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Invalid request format: " + err.Error(),
			})
			return
		}
		if len(req.Wallets) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "At least one address is required",
			})
			return
		}
		if len(req.Wallets) > 100 {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Maximum 100 Wallets allowed per request",
			})
			return
		}
		results := make([]models.BalanceItem, len(req.Wallets))
		var wg sync.WaitGroup
		for i, address := range req.Wallets {
			wg.Add(1)
			go func(index int, addr string) {
				defer wg.Done()
				results[index] = balanceService.ProcessAddress(addr)
			}(i, address)
		}
		wg.Wait()
		c.JSON(http.StatusOK, models.GetBalanceResponse{
			Success: true,
			Results: results,
		})
	}
}

func HealthHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": 0, 
			"service":   "solana-balance-api",
		})
	}
} 