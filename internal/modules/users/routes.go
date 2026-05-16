package users

import (
	"github.com/gin-gonic/gin"
	"github.com/nodus-protocol/backend/internal/middleware"
	"github.com/nodus-protocol/backend/internal/utils"
	"github.com/redis/go-redis/v9"
)

// RegisterRoutes mounts all user routes (all protected by JWT auth).
func RegisterRoutes(rg *gin.RouterGroup, h *Handler, jwtManager *utils.JWTManager, rdb *redis.Client) {
	users := rg.Group("/users")
	users.Use(middleware.Auth(jwtManager, rdb))
	{
		users.GET("/me", h.GetMe)
		users.PUT("/me", h.UpdateMe)
		users.PUT("/me/password", h.ChangePassword)
		users.DELETE("/me", h.DeleteMe)
	}
}
