package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nodus-protocol/backend/internal/config"
	"github.com/nodus-protocol/backend/internal/database"
	"github.com/nodus-protocol/backend/internal/middleware"
	"github.com/nodus-protocol/backend/internal/modules/auth"
	"github.com/nodus-protocol/backend/internal/modules/users"
	"github.com/nodus-protocol/backend/internal/utils"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// @title           Nodus Protocol API
// @version         1.0
// @description     Production-ready REST API with full user authentication.
// @termsOfService  https://nodus-protocol.io/terms

// @contact.name   Nodus Protocol Team
// @contact.url    https://nodus-protocol.io
// @contact.email  support@nodus-protocol.io

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

func main() {
	// ── Config ────────────────────────────────────────────────────────────────
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	// ── Logger ────────────────────────────────────────────────────────────────
	log := newLogger(cfg)
	defer log.Sync() //nolint:errcheck

	// ── Database ──────────────────────────────────────────────────────────────
	db, err := database.NewPostgres(cfg, log)
	if err != nil {
		log.Fatal("postgres connection failed", zap.Error(err))
	}

	// Run auto-migrations in non-production or when DB_AUTOMIGRATE=true
	if !cfg.IsProd() || os.Getenv("DB_AUTOMIGRATE") == "true" {
		if err := database.AutoMigrate(db); err != nil {
			log.Fatal("auto-migration failed", zap.Error(err))
		}
		log.Info("database migrations applied")
	}

	// ── Redis ─────────────────────────────────────────────────────────────────
	rdb, err := database.NewRedis(cfg, log)
	if err != nil {
		log.Fatal("redis connection failed", zap.Error(err))
	}
	defer rdb.Close()

	// ── JWT Manager ───────────────────────────────────────────────────────────
	jwtManager, err := utils.NewJWTManager(cfg.JWT)
	if err != nil {
		log.Fatal("failed to initialize JWT manager", zap.Error(err))
	}

	// ── Mailer ────────────────────────────────────────────────────────────────
	mailer := utils.NewMailer(cfg.Email, log)

	// ── Gin Engine ────────────────────────────────────────────────────────────
	if cfg.IsProd() {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(
		middleware.Recovery(),
		middleware.SecurityHeaders(),
		middleware.CORS(cfg),
		middleware.RequestLogger(log),
		middleware.RateLimiter(300), // global: 300 req/min per IP
	)

	// Health check
	router.GET("/health", func(c *gin.Context) {
		sqlDB, _ := db.DB()
		dbErr := sqlDB.Ping()
		redisErr := rdb.Ping(context.Background()).Err()

		status := http.StatusOK
		if dbErr != nil || redisErr != nil {
			status = http.StatusServiceUnavailable
		}

		c.JSON(status, gin.H{
			"status":   "ok",
			"database": dbErr == nil,
			"redis":    redisErr == nil,
		})
	})

	// ── API v1 ────────────────────────────────────────────────────────────────
	v1 := router.Group("/api/v1")

	// Wire up Auth module
	authRepo := auth.NewRepository(db)
	authSvc := auth.NewService(authRepo, jwtManager, rdb, mailer, cfg, log)
	authHandler := auth.NewHandler(authSvc, log)
	auth.RegisterRoutes(v1, authHandler, jwtManager, rdb)

	// Wire up Users module
	usersRepo := users.NewRepository(db)
	usersSvc := users.NewService(usersRepo, log)
	usersHandler := users.NewHandler(usersSvc, log)
	users.RegisterRoutes(v1, usersHandler, jwtManager, rdb)

	// ── HTTP Server ───────────────────────────────────────────────────────────
	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Start server in goroutine
	go func() {
		log.Info("server starting", zap.String("addr", srv.Addr), zap.String("env", cfg.App.Env))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("server failed", zap.Error(err))
		}
	}()

	// ── Graceful Shutdown ─────────────────────────────────────────────────────
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down server gracefully...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("server forced to shutdown", zap.Error(err))
	}

	log.Info("server stopped")
}

// newLogger creates a production or development Zap logger based on config.
func newLogger(cfg *config.Config) *zap.Logger {
	var zapCfg zap.Config
	if cfg.IsProd() {
		zapCfg = zap.NewProductionConfig()
		zapCfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	} else {
		zapCfg = zap.NewDevelopmentConfig()
		zapCfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	level, err := zapcore.ParseLevel(cfg.Log.Level)
	if err == nil {
		zapCfg.Level = zap.NewAtomicLevelAt(level)
	}

	log, err := zapCfg.Build()
	if err != nil {
		panic(fmt.Sprintf("failed to initialize logger: %v", err))
	}
	return log
}
