package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/nodus-protocol/backend/internal/middleware"
	"github.com/nodus-protocol/backend/internal/utils"
	"github.com/redis/go-redis/v9"
)

// RegisterRoutes mounts all auth routes onto the given router group.
func RegisterRoutes(rg *gin.RouterGroup, h *Handler, sep10h *Sep10Handler, jwtManager *utils.JWTManager, rdb *redis.Client) {
	auth := rg.Group("/auth")

	authLimiterStore, err := middleware.NewLimiterStore(rdb, "nodus_rl_auth")
	if err != nil {
		panic("failed to initialize auth rate limiter store: " + err.Error())
	}

	// Public routes — rate-limited strictly
	public := auth.Group("")
	public.Use(middleware.StrictRateLimiter(authLimiterStore, 20))
	{
		public.POST("/register", h.Register)
		public.POST("/login", h.Login)
		public.POST("/refresh", h.RefreshToken)
		public.GET("/verify-email", h.VerifyEmail)
		public.POST("/resend-verification", h.ResendVerification)
		public.POST("/forgot-password", h.ForgotPassword)
		public.POST("/reset-password", h.ResetPassword)

		// SEP-10 Stellar Web Authentication
		if sep10h != nil {
			public.GET("/stellar/challenge", sep10h.StellarChallenge)
			public.POST("/stellar/token", sep10h.StellarToken)
		}
	}

	// Protected routes
	protected := auth.Group("")
	protected.Use(middleware.Auth(jwtManager, rdb))
	{
		protected.POST("/logout", h.Logout)
	}
}
