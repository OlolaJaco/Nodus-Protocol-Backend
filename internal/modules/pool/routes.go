package pool

import (
	"github.com/gin-gonic/gin"
	"github.com/nodus-protocol/backend/internal/middleware"
	"github.com/nodus-protocol/backend/internal/utils"
	"github.com/redis/go-redis/v9"
)

func RegisterRoutes(rg *gin.RouterGroup, h *Handler, jwtManager *utils.JWTManager, rdb *redis.Client) {
	p := rg.Group("/pool")

	// Public — read-only pool state (used by frontend and mobile without auth)
	p.GET("/reserves",      h.GetReserves)
	p.GET("/quote",         h.GetQuote)
	p.GET("/lp-balance",    h.GetLPBalance)
	p.GET("/stats",         h.GetStats)
	p.GET("/snapshots",     h.GetSnapshots)
	p.GET("/tvl",           h.GetTVL)
	p.GET("/price-history", h.GetPriceHistory)
	p.GET("/overview",      h.GetOverview)

	// Protected — transaction building requires auth so we can log user context
	protected := p.Group("")
	protected.Use(middleware.Auth(jwtManager, rdb))
	{
		protected.POST("/build/swap",             h.BuildSwapParams)
		protected.POST("/build/add-liquidity",    h.BuildAddLiquidity)
		protected.POST("/build/remove-liquidity", h.BuildRemoveLiquidity)
	}
}
