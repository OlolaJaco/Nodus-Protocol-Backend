package payments

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/nodus-protocol/backend/internal/models"
	"go.uber.org/zap"
)

type Service struct {
	repo   *Repository
	engine *Client
	log    *zap.Logger
}

func NewService(repo *Repository, engine *Client, log *zap.Logger) *Service {
	return &Service{repo: repo, engine: engine, log: log}
}

func (s *Service) Initiate(ctx context.Context, userID uuid.UUID, req initiateRequest) (*models.Transaction, error) {
	ep, err := s.engine.InitiatePayment(ctx, req)
	if err != nil {
		s.log.Error("core engine initiate failed", zap.Error(err))
		return nil, err
	}

	tx := &models.Transaction{
		UserID:     userID,
		EngineID:   ep.ID,
		Sender:     ep.Sender,
		Recipient:  ep.Recipient,
		Amount:     ep.Amount,
		Token:      ep.Token,
		Status:     ep.Status,
		FeeStroops: ep.FeeStroops,
		Urgency:    ep.Urgency,
	}
	if ep.TxHash != nil {
		tx.TxHash = *ep.TxHash
	}
	if ep.Error != nil {
		tx.Error = *ep.Error
	}

	if err := s.repo.Create(tx); err != nil {
		s.log.Error("failed to persist transaction", zap.Error(err), zap.String("engine_id", ep.ID))
		return nil, err
	}

	return tx, nil
}

func (s *Service) GetByID(ctx context.Context, txID, userID uuid.UUID) (*models.Transaction, error) {
	return s.repo.FindByID(txID, userID)
}

func (s *Service) List(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]models.Transaction, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	return s.repo.ListByUser(userID, pageSize, offset)
}

func (s *Service) Simulate(ctx context.Context, req simulateRequest) (*engineSimulation, error) {
	return s.engine.SimulatePayment(ctx, req)
}

func (s *Service) Batch(ctx context.Context, userID uuid.UUID, items []batchItem) (*engineBatchResult, error) {
	result, err := s.engine.BatchPayments(ctx, items)
	if err != nil {
		return nil, err
	}

	for _, r := range result.Results {
		if r.Payment == nil {
			continue
		}
		ep := r.Payment
		tx := &models.Transaction{
			UserID:     userID,
			EngineID:   ep.ID,
			Sender:     ep.Sender,
			Recipient:  ep.Recipient,
			Amount:     ep.Amount,
			Token:      ep.Token,
			Status:     ep.Status,
			FeeStroops: ep.FeeStroops,
			Urgency:    ep.Urgency,
		}
		if ep.TxHash != nil {
			tx.TxHash = *ep.TxHash
		}
		if err := s.repo.Create(tx); err != nil {
			s.log.Warn("failed to persist batch transaction",
				zap.Error(err), zap.String("engine_id", ep.ID))
		}
	}

	return result, nil
}

func (s *Service) GetReceipt(ctx context.Context, txID, userID uuid.UUID) (*engineReceipt, error) {
	tx, err := s.repo.FindByID(txID, userID)
	if err != nil {
		return nil, err
	}
	return s.engine.GetReceipt(ctx, tx.EngineID)
}

func (s *Service) GetFees(ctx context.Context) ([]engineFees, error) {
	return s.engine.GetFees(ctx)
}

func (s *Service) GetRates(ctx context.Context, tokens string) ([]engineRate, error) {
	return s.engine.GetRates(ctx, strings.TrimSpace(tokens))
}

func (s *Service) EngineHealth(ctx context.Context) (*engineHealth, error) {
	return s.engine.Health(ctx)
}

// SyncStatus refreshes a transaction's status from the Core Engine.
// Called by the webhook handler or on-demand by the client.
func (s *Service) SyncStatus(ctx context.Context, engineID string) error {
	ep, err := s.engine.GetPayment(ctx, engineID)
	if err != nil {
		return err
	}
	txHash := ""
	if ep.TxHash != nil {
		txHash = *ep.TxHash
	}
	errMsg := ""
	if ep.Error != nil {
		errMsg = *ep.Error
	}
	return s.repo.UpdateStatus(engineID, ep.Status, txHash, errMsg)
}
