package users

import (
	"github.com/gin-gonic/gin"
	"github.com/nodus-protocol/backend/internal/middleware"
	"github.com/nodus-protocol/backend/internal/utils"
	"github.com/redis/go-redis/v9"
)

func RegisterRoutes(rg *gin.RouterGroup, h *Handler, jwtManager *utils.JWTManager, rdb *redis.Client) {
	authMW := middleware.Auth(jwtManager, rdb)
	adminMW := middleware.RequireRole("admin")

	// Self (authenticated user)
	me := rg.Group("/users/me")
	me.Use(authMW)
	{
		me.GET("", h.GetMe)
		me.PUT("", h.UpdateMe)
		me.DELETE("", h.DeleteMe)
		me.PUT("/password", h.ChangePassword)

		// Stellar wallet
		me.POST("/wallet", h.LinkWallet)
		me.DELETE("/wallet", h.UnlinkWallet)

		// Pool position
		me.GET("/lp-position", h.GetLPPosition)

		// Transaction history
		me.GET("/transactions", h.ListTransactions)
		me.GET("/transactions/:id", h.GetTransaction)

		// GDPR export
		me.GET("/export", h.ExportData)
	}

	// Admin (requires admin role, inherits auth)
	admin := rg.Group("/admin")
	admin.Use(authMW, adminMW)
	{
		admin.GET("/users", h.AdminListUsers)
		admin.GET("/users/:id", h.AdminGetUser)
		admin.PUT("/users/:id/role", h.AdminUpdateRole)
		admin.DELETE("/users/:id", h.AdminDeleteUser)
		admin.GET("/stats", h.AdminStats)
		admin.GET("/transactions", h.AdminListTransactions)
	}

	// Leaderboard (public)
	rg.GET("/leaderboard/traders", h.LeaderboardTraders)
}
