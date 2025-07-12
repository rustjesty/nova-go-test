package models

import "time"

// Response structures
type BalanceResponse struct {
	Success bool    `json:"success"`
	Balance float64 `json:"balance"`
	Address string  `json:"address"`
	Error   string  `json:"error,omitempty"`
}

type BalanceItem struct {
	Address string  `json:"address"`
	Balance float64 `json:"balance"`
	Error   string  `json:"error,omitempty"`
}

type GetBalanceRequest struct {
	Wallets []string `json:"Wallets" binding:"required"`
}

type GetBalanceResponse struct {
	Success bool          `json:"success"`
	Results []BalanceItem `json:"results"`
	Error   string        `json:"error,omitempty"`
}

type CacheEntry struct {
	Balance   float64
	Timestamp time.Time
} 