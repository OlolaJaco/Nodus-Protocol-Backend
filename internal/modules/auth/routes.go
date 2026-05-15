package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/nodus-protocol/backend/internal/middleware"
	"github.com/redis/go-redis/v9"
	"github.com/nodus-protocol/backend/internal/utils"
)

// RegisterRoutes mounts all auth routes onto the given router group.
func RegisterRoutes(rg *gin.RouterGroup, h *Handler, jwtManager *utils.JWTManager, rdb *redis.Client) {
	auth := rg.Group("/auth")

	// Public routes — rate-limited strictly
	public := auth.Group("")
	public.Use(middleware.StrictRateLimiter(20))
	{
		public.POST("/register", h.Register)
		public.POST("/login", h.Login)
		public.POST("/refresh", h.RefreshToken)
		public.GET("/verify-email", h.VerifyEmail)
		public.POST("/resend-verification", h.ResendVerification)
		public.POST("/forgot-password", h.ForgotPassword)
		public.POST("/reset-password", h.ResetPassword)
	}

	// Protected routes
	protected := auth.Group("")
	protected.Use(middleware.Auth(jwtManager, rdb))
	{
		protected.POST("/logout", h.Logout)
	}
}
