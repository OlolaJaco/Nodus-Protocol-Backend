package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// RequestTimeout aborts requests that exceed the given duration with 503.
// Use for protecting against slow clients or runaway handlers.
func RequestTimeout(d time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		done := make(chan struct{})

		go func() {
			c.Next()
			close(done)
		}()

		select {
		case <-done:
			// handler completed normally
		case <-time.After(d):
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
				"code":    "REQUEST_TIMEOUT",
				"message": "request exceeded maximum allowed duration",
			})
		}
	}
}
