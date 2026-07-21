package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/ulule/limiter/v3"
	ginlimiter "github.com/ulule/limiter/v3/drivers/middleware/gin"
	"github.com/ulule/limiter/v3/drivers/store/memory"
	limiterredis "github.com/ulule/limiter/v3/drivers/store/redis"
)

// NewLimiterStore builds the store backing a rate limiter middleware.
// When rdb is non-nil it returns a Redis-backed store, so counters are
// shared across replicas and survive deploys/restarts. When rdb is nil
// (e.g. local dev without Redis, or tests) it falls back to an in-memory
// store, which is per-instance and resets on restart.
//
// prefix namespaces the store's keys in Redis. Callers that mount more
// than one rate limiter (e.g. a global limiter and a stricter per-route
// one) must use distinct prefixes — both key solely by client IP, so
// sharing a prefix would make them double-count the same requests.
func NewLimiterStore(rdb *redis.Client, prefix string) (limiter.Store, error) {
	if rdb == nil {
		return memory.NewStore(), nil
	}
	return limiterredis.NewStoreWithOptions(rdb, limiter.StoreOptions{
		Prefix:          prefix,
		MaxRetry:        limiter.DefaultMaxRetry,
		CleanUpInterval: limiter.DefaultCleanUpInterval,
	})
}

// RateLimiter returns a rate limiter middleware backed by the given store.
func RateLimiter(store limiter.Store, requestsPerMinute int64) gin.HandlerFunc {
	rate := limiter.Rate{
		Period: time.Minute,
		Limit:  requestsPerMinute,
	}
	instance := limiter.New(store, rate)

	mw := ginlimiter.NewMiddleware(instance, ginlimiter.WithLimitReachedHandler(func(c *gin.Context) {
		c.JSON(http.StatusTooManyRequests, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "RATE_LIMIT_EXCEEDED",
				"message": "too many requests, please slow down",
			},
		})
		c.Abort()
	}))

	return mw
}

// StrictRateLimiter returns a tighter rate limiter for sensitive auth endpoints.
func StrictRateLimiter(store limiter.Store, requestsPerMinute int64) gin.HandlerFunc {
	return RateLimiter(store, requestsPerMinute)
}
