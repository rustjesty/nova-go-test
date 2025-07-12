package rate_limiter

import (
	"sync"
	"time"

	"golang.org/x/time/rate"
	"solana-balance-api/config"
)

type RateLimiter struct {
	limiters map[string]*rate.Limiter
	mutex    sync.RWMutex
}

func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		limiters: make(map[string]*rate.Limiter),
	}
}

func (r *RateLimiter) Get(ip string) *rate.Limiter {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if limiter, exists := r.limiters[ip]; exists {
		return limiter
	}
	limiter := rate.NewLimiter(rate.Every(time.Minute/time.Duration(config.RateLimit)), config.RateLimit)
	r.limiters[ip] = limiter
	return limiter
} 