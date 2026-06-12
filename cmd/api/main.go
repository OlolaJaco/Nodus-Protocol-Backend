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
	"github.com/nodus-protocol/backend/internal/modules/payments"
	"github.com/nodus-protocol/backend/internal/modules/pool"
	"github.com/nodus-protocol/backend/internal/modules/users"
	"github.com/nodus-protocol/backend/internal/utils"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	log := newLogger(cfg)
	defer log.Sync() //nolint:errcheck

	db, err := database.NewPostgres(cfg, log)
	if err != nil {
		log.Fatal("postgres connection failed", zap.Error(err))
	}

	if !cfg.IsProd() || os.Getenv("DB_AUTOMIGRATE") == "true" {
		if err := database.AutoMigrate(db); err != nil {
			log.Fatal("auto-migration failed", zap.Error(err))
		}
		log.Info("database migrations applied")
	}

	rdb, err := database.NewRedis(cfg, log)
	if err != nil {
		log.Fatal("redis connection failed", zap.Error(err))
	}
	defer rdb.Close()

	jwtManager, err := utils.NewJWTManager(cfg.JWT)
	if err != nil {
		log.Fatal("failed to initialize JWT manager", zap.Error(err))
	}

	mailer := utils.NewMailer(cfg.Email, log)

	if cfg.IsProd() {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(
		middleware.Recovery(),
		middleware.SecurityHeaders(),
		middleware.CORS(cfg),
		middleware.RequestLogger(log),
		middleware.RateLimiter(300),
	)

	router.GET("/health", func(c *gin.Context) {
		sqlDB, _ := db.DB()
		dbErr := sqlDB.Ping()
		redisErr := rdb.Ping(context.Background()).Err()
		coreEngineOK := pingCoreEngine(cfg.CoreEngine.URL)

		status := http.StatusOK
		if dbErr != nil || redisErr != nil {
			status = http.StatusServiceUnavailable
		}
		c.JSON(status, gin.H{
			"status":      "ok",
			"database":    dbErr == nil,
			"redis":       redisErr == nil,
			"core_engine": coreEngineOK,
		})
	})

	v1 := router.Group("/api/v1")

	// Auth module
	authRepo := auth.NewRepository(db)

	// SEP-10 is optional — only enabled when STELLAR_SERVER_SECRET is configured.
	var sep10Manager *utils.Sep10Manager
	var sep10Handler *auth.Sep10Handler
	if cfg.Stellar.ServerSecretKey != "" {
		var sep10Err error
		sep10Manager, sep10Err = utils.NewSep10Manager(
			cfg.Stellar.ServerSecretKey,
			cfg.Stellar.WebAuthDomain,
			cfg.Stellar.Network,
			cfg.Stellar.ChallengeTTL,
		)
		if sep10Err != nil {
			log.Fatal("failed to initialize SEP-10 manager", zap.Error(sep10Err))
		}
		log.Info("SEP-10 stellar auth enabled", zap.String("server_account", sep10Manager.ServerAddress()))
	} else {
		log.Warn("STELLAR_SERVER_SECRET not set — SEP-10 stellar auth disabled")
	}

	authSvc := auth.NewService(authRepo, jwtManager, sep10Manager, rdb, mailer, cfg, log)
	authHandler := auth.NewHandler(authSvc, log)
	if sep10Manager != nil {
		sep10Handler = auth.NewSep10Handler(authSvc, sep10Manager, log)
	}
	auth.RegisterRoutes(v1, authHandler, sep10Handler, jwtManager, rdb)

	// Pool module — must be wired before Users so the LP fetcher can be injected
	poolClient := pool.NewClient(cfg.CoreEngine.URL)
	poolSvc := pool.NewService(poolClient, db, rdb, log)
	poolHandler := pool.NewHandler(poolSvc, log)
	pool.RegisterRoutes(v1, poolHandler, jwtManager, rdb)

	// Users module
	usersRepo := users.NewRepository(db)
	usersSvc := users.NewService(usersRepo, log).WithLPFetcher(poolSvc)
	usersHandler := users.NewHandler(usersSvc, log)
	users.RegisterRoutes(v1, usersHandler, jwtManager, rdb)

	// Payments module — wired to Core Engine
	paymentsClient := payments.NewClient(cfg.CoreEngine.URL)
	paymentsRepo := payments.NewRepository(db)
	paymentsSvc := payments.NewService(paymentsRepo, paymentsClient, log)
	paymentsHandler := payments.NewHandler(paymentsSvc, log)
	payments.RegisterRoutes(v1, paymentsHandler, jwtManager, rdb)
	log.Info("payments module wired", zap.String("core_engine", cfg.CoreEngine.URL))

	if cfg.Stellar.PoolConfigured() {
		log.Info("AMM pool module wired",
			zap.String("contract", cfg.Stellar.PoolContractID),
			zap.String("token_0", cfg.Stellar.PoolToken0),
			zap.String("token_1", cfg.Stellar.PoolToken1),
		)
	} else {
		log.Warn("AMM pool contract not configured — set POOL_CONTRACT_ID, SOROBAN_RPC_URL, POOL_TOKEN_0, POOL_TOKEN_1 in .env")
	}

	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	go func() {
		log.Info("server starting", zap.String("addr", srv.Addr), zap.String("env", cfg.App.Env))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("server failed", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down gracefully...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("forced shutdown", zap.Error(err))
	}
	log.Info("server stopped")
}

func pingCoreEngine(baseURL string) bool {
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(baseURL + "/healthz")
	if err != nil {
		return false
	}
	resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

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
