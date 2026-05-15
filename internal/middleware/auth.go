package middleware

import (
	"context"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nodus-protocol/backend/internal/utils"
	"github.com/redis/go-redis/v9"
)

const (
	ContextKeyUserID = "user_id"
	ContextKeyEmail  = "user_email"
	ContextKeyRole   = "user_role"
)

// Auth returns a Gin middleware that validates Bearer JWT access tokens.
// It also checks the Redis blacklist for logged-out tokens.
func Auth(jwtManager *utils.JWTManager, rdb *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.Unauthorized(c, "authorization header is required")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
			utils.Unauthorized(c, "authorization header must be 'Bearer <token>'")
			c.Abort()
			return
		}

		tokenStr := parts[1]

		// Check if token is blacklisted (logged out)
		jti, err := jwtManager.ExtractJTI(tokenStr)
		if err == nil && jti != "" {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			blacklisted, _ := rdb.Exists(ctx, "blacklist:"+jti).Result()
			if blacklisted > 0 {
				utils.Unauthorized(c, "token has been revoked")
				c.Abort()
				return
			}
		}

		claims, err := jwtManager.ValidateAccessToken(tokenStr)
		if err != nil {
			if err == utils.ErrTokenExpired {
				utils.Unauthorized(c, "access token has expired")
			} else {
				utils.Unauthorized(c, "invalid access token")
			}
			c.Abort()
			return
		}

		// Inject claims into context
		c.Set(ContextKeyUserID, claims.UserID)
		c.Set(ContextKeyEmail, claims.Email)
		c.Set(ContextKeyRole, claims.Role)
		c.Next()
	}
}

// RequireRole returns a middleware that restricts access to users with a specific role.
func RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get(ContextKeyRole)
		if !exists {
			utils.Unauthorized(c, "not authenticated")
			c.Abort()
			return
		}
		if userRole != role {
			utils.Forbidden(c, "insufficient permissions")
			c.Abort()
			return
		}
		c.Next()
	}
}
