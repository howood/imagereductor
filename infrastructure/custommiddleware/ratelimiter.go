package custommiddleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/labstack/echo/v5"
	"golang.org/x/time/rate"
)

const (
	// DefaultCleanupTTL is the default cleanup duration for unused limiters.
	DefaultCleanupTTL = 10 * time.Second
	// DefaultCleanupEvery is the default interval for running the cleanup routine.
	DefaultCleanupEvery = 5 * time.Minute
	// DefaultMaxKeys is the default maximum number of keys to store.
	DefaultMaxKeys = 10000
)

type limiterEntry struct {
	limiter    *rate.Limiter
	lastAccess time.Time
}

type RateLimitConfig struct {
	Rate         rate.Limit
	Burst        int
	KeyFunc      func(c *echo.Context) string
	ErrorMsg     string
	Skipper      func(c *echo.Context) bool
	CleanupTTL   time.Duration // How long to keep unused limiters
	CleanupEvery time.Duration // How often to run cleanup
	MaxKeys      int           // Maximum number of keys to store (0 = unlimited)
}

type RateLimiter struct {
	limiters map[string]*limiterEntry
	mutex    sync.RWMutex
	config   RateLimitConfig
}

func NewRateLimiter(config RateLimitConfig) *RateLimiter {
	if config.KeyFunc == nil {
		config.KeyFunc = func(c *echo.Context) string {
			return c.RealIP()
		}
	}
	if config.ErrorMsg == "" {
		config.ErrorMsg = "Rate limit exceeded"
	}
	if config.Skipper == nil {
		config.Skipper = func(_ *echo.Context) bool {
			return false
		}
	}
	if config.CleanupTTL == 0 {
		config.CleanupTTL = DefaultCleanupTTL
	}
	if config.CleanupEvery == 0 {
		config.CleanupEvery = DefaultCleanupEvery
	}
	if config.MaxKeys == 0 {
		config.MaxKeys = DefaultMaxKeys
	}

	rl := &RateLimiter{
		limiters: make(map[string]*limiterEntry),
		config:   config,
	}

	go rl.cleanupLoop()

	return rl
}

func (rl *RateLimiter) Middleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
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

func (rl *RateLimiter) getLimiter(key string) *rate.Limiter {
	now := time.Now()
	rl.mutex.RLock()
	entry, exists := rl.limiters[key]
	rl.mutex.RUnlock()

	if !exists {
		rl.mutex.Lock()
		// Double-check after acquiring write lock
		entry, exists = rl.limiters[key]
		if !exists {
			// Check if we've reached the maximum number of keys
			if rl.config.MaxKeys > 0 && len(rl.limiters) >= rl.config.MaxKeys {
				// Evict oldest entry (simple LRU)
				rl.evictOldest()
			}

			entry = &limiterEntry{
				limiter:    rate.NewLimiter(rl.config.Rate, rl.config.Burst),
				lastAccess: now,
			}
			rl.limiters[key] = entry
		}
		rl.mutex.Unlock()
	} else {
		// Update last access time
		rl.mutex.Lock()
		entry.lastAccess = now
		rl.mutex.Unlock()
	}

	return entry.limiter
}

// evictOldest removes the oldest entry from the limiters map.
// Must be called with write lock held.
func (rl *RateLimiter) evictOldest() {
	var oldestKey string
	var oldestTime time.Time
	first := true

	for key, entry := range rl.limiters {
		if first || entry.lastAccess.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.lastAccess
			first = false
		}
	}

	if oldestKey != "" {
		delete(rl.limiters, oldestKey)
	}
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

func IPKeyFunc(c *echo.Context) string {
	return c.RealIP()
}

func APIKeyFunc(c *echo.Context) string {
	apiKey := c.Request().Header.Get("Authorization")
	if apiKey != "" {
		return "api:" + apiKey
	}
	return c.RealIP()
}
