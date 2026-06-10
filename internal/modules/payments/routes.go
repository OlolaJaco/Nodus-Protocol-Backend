package payments

import (
	"github.com/gin-gonic/gin"
	"github.com/nodus-protocol/backend/internal/middleware"
	"github.com/nodus-protocol/backend/internal/utils"
	"github.com/redis/go-redis/v9"
)

func RegisterRoutes(rg *gin.RouterGroup, h *Handler, jwtManager *utils.JWTManager, rdb *redis.Client) {
	p := rg.Group("/payments")

	// Public — webhook receiver (secured by shared secret checked in handler)
	p.POST("/webhook", h.WebhookHandler)

	// Public — rates and fees (no auth needed, used by frontend)
	p.GET("/fees", h.GetFees)
	p.GET("/rates", h.GetRates)
	p.GET("/engine/health", h.EngineHealth)

	// Protected — all payment operations require JWT
	protected := p.Group("")
	protected.Use(middleware.Auth(jwtManager, rdb))
	{
		protected.POST("", h.InitiatePayment)
		protected.GET("", h.ListPayments)
		protected.POST("/simulate", h.SimulatePayment)
		protected.POST("/batch", h.BatchPayments)
		protected.GET("/:id", h.GetPayment)
		protected.GET("/:id/receipt", h.GetReceipt)
	}
}
