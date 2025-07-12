package cache

import (
	"sync"
	"time"

	"solana-balance-api/models"
	"solana-balance-api/config"
)

type Cache struct {
	data  map[string]models.CacheEntry
	mutex sync.RWMutex
}

func NewCache() *Cache {
	return &Cache{
		data: make(map[string]models.CacheEntry),
	}
}

func (c *Cache) Get(address string) (float64, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	if entry, exists := c.data[address]; exists {
		if time.Since(entry.Timestamp) < config.CacheTTL {
			return entry.Balance, true
		}
	}
	return 0, false
}

func (c *Cache) Set(address string, balance float64) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.data[address] = models.CacheEntry{
		Balance:   balance,
		Timestamp: time.Now(),
	}
}

func (c *Cache) Cleanup() {
	ticker := time.NewTicker(time.Minute)
	go func() {
		for range ticker.C {
			c.mutex.Lock()
			now := time.Now()
			for address, entry := range c.data {
				if now.Sub(entry.Timestamp) > config.CacheTTL {
					delete(c.data, address)
				}
			}
			c.mutex.Unlock()
		}
	}()
} 