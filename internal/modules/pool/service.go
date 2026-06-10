package pool

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/nodus-protocol/backend/internal/models"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const (
	cacheKeyReserves = "pool:reserves"
	cacheKeyStats    = "pool:stats"
	cacheTTL         = 30 * time.Second
)

type Service struct {
	client *Client
	db     *gorm.DB
	rdb    *redis.Client
	log    *zap.Logger
}

func NewService(client *Client, db *gorm.DB, rdb *redis.Client, log *zap.Logger) *Service {
	return &Service{client: client, db: db, rdb: rdb, log: log}
}

func (s *Service) GetReserves(ctx context.Context) (*reserves, error) {
	if cached := s.getCache(ctx, cacheKeyReserves); cached != nil {
		var out reserves
		if err := json.Unmarshal(cached, &out); err == nil {
			return &out, nil
		}
	}

	r, err := s.client.GetReserves(ctx)
	if err != nil {
		return nil, err
	}

	s.setCache(ctx, cacheKeyReserves, r)
	go s.snapshotReserves(r)
	return r, nil
}

func (s *Service) GetQuote(ctx context.Context, amountIn, tokenIn string) (*priceQuote, error) {
	return s.client.GetQuote(ctx, amountIn, tokenIn)
}

func (s *Service) GetLPBalance(ctx context.Context, address string) (*lpBalance, error) {
	return s.client.GetLPBalance(ctx, address)
}

func (s *Service) GetStats(ctx context.Context) (*poolStats, error) {
	if cached := s.getCache(ctx, cacheKeyStats); cached != nil {
		var out poolStats
		if err := json.Unmarshal(cached, &out); err == nil {
			return &out, nil
		}
	}

	st, err := s.client.GetStats(ctx)
	if err != nil {
		return nil, err
	}

	s.setCache(ctx, cacheKeyStats, st)
	return st, nil
}

func (s *Service) BuildSwapParams(ctx context.Context, req swapParamsRequest) (*unsignedTx, error) {
	return s.client.BuildSwapParams(ctx, req)
}

func (s *Service) BuildAddLiquidity(ctx context.Context, req addLiquidityRequest) (*unsignedTx, error) {
	return s.client.BuildAddLiquidity(ctx, req)
}

func (s *Service) BuildRemoveLiquidity(ctx context.Context, req removeLiquidityRequest) (*unsignedTx, error) {
	return s.client.BuildRemoveLiquidity(ctx, req)
}

func (s *Service) RecentSnapshots(ctx context.Context, limit int) ([]models.PoolSnapshot, error) {
	var snaps []models.PoolSnapshot
	err := s.db.
		Order("created_at DESC").
		Limit(limit).
		Find(&snaps).Error
	return snaps, err
}

func (s *Service) getCache(ctx context.Context, key string) []byte {
	val, err := s.rdb.Get(ctx, key).Bytes()
	if err != nil {
		return nil
	}
	return val
}

func (s *Service) setCache(ctx context.Context, key string, v any) {
	b, err := json.Marshal(v)
	if err != nil {
		return
	}
	if err := s.rdb.Set(ctx, key, b, cacheTTL).Err(); err != nil {
		s.log.Warn("pool cache set failed", zap.String("key", key), zap.Error(err))
	}
}

func (s *Service) snapshotReserves(r *reserves) {
	snap := &models.PoolSnapshot{
		ContractID:    fmt.Sprintf("pool:%s/%s", r.Token0, r.Token1),
		Reserve0:      r.Reserve0,
		Reserve1:      r.Reserve1,
		Token0:        r.Token0,
		Token1:        r.Token1,
		LpTotalSupply: r.LpTotalSupply,
		TimestampLast: int64(r.TimestampLast),
	}
	if err := s.db.Create(snap).Error; err != nil {
		s.log.Warn("pool snapshot failed", zap.Error(err))
	}
}
