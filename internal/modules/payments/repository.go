package payments

import (
	"github.com/google/uuid"
	"github.com/nodus-protocol/backend/internal/models"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(tx *models.Transaction) error {
	return r.db.Create(tx).Error
}

func (r *Repository) FindByID(id, userID uuid.UUID) (*models.Transaction, error) {
	var tx models.Transaction
	err := r.db.
		Where("id = ? AND user_id = ? AND deleted_at IS NULL", id, userID).
		First(&tx).Error
	return &tx, err
}

func (r *Repository) FindByEngineID(engineID string, userID uuid.UUID) (*models.Transaction, error) {
	var tx models.Transaction
	err := r.db.
		Where("engine_id = ? AND user_id = ? AND deleted_at IS NULL", engineID, userID).
		First(&tx).Error
	return &tx, err
}

func (r *Repository) ListByUser(userID uuid.UUID, limit, offset int) ([]models.Transaction, int64, error) {
	var txs []models.Transaction
	var total int64

	base := r.db.Model(&models.Transaction{}).
		Where("user_id = ? AND deleted_at IS NULL", userID)

	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := base.
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&txs).Error

	return txs, total, err
}

func (r *Repository) UpdateStatus(engineID, status, txHash, errMsg string) error {
	updates := map[string]any{"status": status}
	if txHash != "" {
		updates["tx_hash"] = txHash
	}
	if errMsg != "" {
		updates["error"] = errMsg
	}
	return r.db.Model(&models.Transaction{}).
		Where("engine_id = ?", engineID).
		Updates(updates).Error
}
