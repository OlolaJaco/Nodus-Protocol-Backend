package database

import (
	"context"
	"fmt"
	"time"

	"github.com/nodus-protocol/backend/internal/config"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// NewRedis creates and returns a connected Redis client.
func NewRedis(cfg *config.Config, log *zap.Logger) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         cfg.Redis.Addr(),
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     20,
		MinIdleConns: 5,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	log.Info("Redis connected", zap.String("addr", cfg.Redis.Addr()))
	return client, nil
}
