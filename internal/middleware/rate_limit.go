package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ulule/limiter/v3"
	ginlimiter "github.com/ulule/limiter/v3/drivers/middleware/gin"
	"github.com/ulule/limiter/v3/drivers/store/memory"
)

// RateLimiter returns an in-memory rate limiter middleware.
// For production with multiple instances, swap the memory store for a Redis store.
func RateLimiter(requestsPerMinute int64) gin.HandlerFunc {
	rate := limiter.Rate{
		Period: time.Minute,
		Limit:  requestsPerMinute,
	}
	store := memory.NewStore()
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
func StrictRateLimiter(requestsPerMinute int64) gin.HandlerFunc {
	return RateLimiter(requestsPerMinute)
}
