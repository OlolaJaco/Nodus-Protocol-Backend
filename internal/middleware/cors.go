package middleware

import (
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/nodus-protocol/backend/internal/config"
)

// CORS returns a configured CORS middleware.
func CORS(cfg *config.Config) gin.HandlerFunc {
	allowedOrigins := []string{cfg.App.FrontendURL}
	if !cfg.IsProd() {
		allowedOrigins = append(allowedOrigins, "http://localhost:3000", "http://localhost:5173", "http://127.0.0.1:3000")
	}

	return cors.New(cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodOptions},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Accept", "X-Request-ID"},
		ExposeHeaders:    []string{"X-Request-ID", "X-RateLimit-Remaining"},
		AllowCredentials: true,
		MaxAge:           43200, // 12 hours preflight cache
	})
}
