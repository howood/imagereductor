package custommiddleware

import (
	"net/http"
	"sync"

	"time"

	"github.com/labstack/echo/v4"
	"golang.org/x/time/rate"
)

type limiterEntry struct {
	limiter    *rate.Limiter
	lastAccess time.Time
}

type RateLimitConfig struct {
	Rate         rate.Limit
	Burst        int
	KeyFunc      func(c echo.Context) string
	ErrorMsg     string
	Skipper      func(c echo.Context) bool
	CleanupTTL   time.Duration // How long to keep unused limiters
	CleanupEvery time.Duration // How often to run cleanup
}

type RateLimiter struct {
	limiters map[string]*limiterEntry
	mutex    sync.RWMutex
	config   RateLimitConfig
}

func NewRateLimiter(config RateLimitConfig) *RateLimiter {
	if config.KeyFunc == nil {
		config.KeyFunc = func(c echo.Context) string {
			return c.RealIP()
		}
	}
	if config.ErrorMsg == "" {
		config.ErrorMsg = "Rate limit exceeded"
	}
	if config.Skipper == nil {
		config.Skipper = func(c echo.Context) bool {
			return false
		}
	}
	if config.CleanupTTL == 0 {
		config.CleanupTTL = 10 * time.Minute
	}
	if config.CleanupEvery == 0 {
		config.CleanupEvery = 5 * time.Minute
	}

	rl := &RateLimiter{
		limiters: make(map[string]*limiterEntry),
		config:   config,
	}

	go rl.cleanupLoop()

	return rl
}

func (rl *RateLimiter) getLimiter(key string) *rate.Limiter {
	now := time.Now()
	rl.mutex.RLock()
	entry, exists := rl.limiters[key]
	rl.mutex.RUnlock()

	if !exists {
		rl.mutex.Lock()
		entry, exists = rl.limiters[key]
		if !exists {
			entry = &limiterEntry{
				limiter:    rate.NewLimiter(rl.config.Rate, rl.config.Burst),
				lastAccess: now,
			}
			rl.limiters[key] = entry
		}
		rl.mutex.Unlock()
	}
	// Update last access time
	rl.mutex.Lock()
	entry.lastAccess = now
	rl.mutex.Unlock()

	return entry.limiter
}

func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(rl.config.CleanupEvery)
	defer ticker.Stop()
	for range ticker.C {
		rl.cleanup()
	}
}

func (rl *RateLimiter) cleanup() {
	threshold := time.Now().Add(-rl.config.CleanupTTL)
	rl.mutex.Lock()
	for key, entry := range rl.limiters {
		if entry.lastAccess.Before(threshold) {
			delete(rl.limiters, key)
		}
	}
	rl.mutex.Unlock()
}

func (rl *RateLimiter) Middleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if rl.config.Skipper(c) {
				return next(c)
			}

			key := rl.config.KeyFunc(c)
			limiter := rl.getLimiter(key)

			if !limiter.Allow() {
				return echo.NewHTTPError(http.StatusTooManyRequests, rl.config.ErrorMsg)
			}

			return next(c)
		}
	}
}

func IPKeyFunc(c echo.Context) string {
	return c.RealIP()
}

func APIKeyFunc(c echo.Context) string {
	apiKey := c.Request().Header.Get("Authorization")
	if apiKey != "" {
		return "api:" + apiKey
	}
	return c.RealIP()
}
