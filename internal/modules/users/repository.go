package users

import (
	"errors"

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

func (r *Repository) FindByID(id uuid.UUID) (*models.User, error) {
	var user models.User
	err := r.db.Where("id = ? AND is_active = true", id).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}
	return &user, err
}

func (r *Repository) FindAll(limit, offset int, search string) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	base := r.db.Model(&models.User{})
	if search != "" {
		like := "%" + search + "%"
		base = base.Where("email ILIKE ? OR first_name ILIKE ? OR last_name ILIKE ?", like, like, like)
	}

	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := base.Order("created_at DESC").Limit(limit).Offset(offset).Find(&users).Error
	return users, total, err
}

func (r *Repository) Update(user *models.User) error {
	return r.db.Save(user).Error
}

func (r *Repository) UpdateStellarAddress(id uuid.UUID, address string) error {
	return r.db.Model(&models.User{}).Where("id = ?", id).Update("stellar_address", address).Error
}

func (r *Repository) UpdateRole(id uuid.UUID, role models.UserRole) error {
	return r.db.Model(&models.User{}).Where("id = ?", id).Update("role", role).Error
}

func (r *Repository) SoftDelete(id uuid.UUID) error {
	return r.db.Model(&models.User{}).Where("id = ?", id).Update("is_active", false).Error
}

func (r *Repository) HardDelete(id uuid.UUID) error {
	return r.db.Unscoped().Delete(&models.User{}, "id = ?", id).Error
}

func (r *Repository) ListTransactionsByUser(
	userID uuid.UUID, limit, offset int, status, token string,
) ([]models.Transaction, int64, error) {
	var txs []models.Transaction
	var total int64

	base := r.db.Model(&models.Transaction{}).Where("user_id = ? AND deleted_at IS NULL", userID)
	if status != "" {
		base = base.Where("status = ?", status)
	}
	if token != "" {
		base = base.Where("token = ?", token)
	}

	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := base.Order("created_at DESC").Limit(limit).Offset(offset).Find(&txs).Error
	return txs, total, err
}

func (r *Repository) FindTransactionByID(txID, userID uuid.UUID) (*models.Transaction, error) {
	var tx models.Transaction
	err := r.db.
		Where("id = ? AND user_id = ? AND deleted_at IS NULL", txID, userID).
		First(&tx).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrTransactionNotFound
	}
	return &tx, err
}

func (r *Repository) ProtocolStats() (map[string]any, error) {
	var totalUsers, activeUsers, verifiedUsers, totalTxs int64
	var volumeResult struct{ Volume *float64 }

	r.db.Model(&models.User{}).Count(&totalUsers)
	r.db.Model(&models.User{}).Where("is_active = true").Count(&activeUsers)
	r.db.Model(&models.User{}).Where("is_email_verified = true").Count(&verifiedUsers)
	r.db.Model(&models.Transaction{}).Where("deleted_at IS NULL").Count(&totalTxs)
	r.db.Model(&models.Transaction{}).
		Where("deleted_at IS NULL AND status = 'confirmed'").
		Select("SUM(amount) as volume").
		Scan(&volumeResult)

	vol := float64(0)
	if volumeResult.Volume != nil {
		vol = *volumeResult.Volume
	}

	return map[string]any{
		"total_users":              totalUsers,
		"active_users":             activeUsers,
		"verified_users":           verifiedUsers,
		"total_transactions":       totalTxs,
		"confirmed_volume_stroops": vol,
	}, nil
}

func (r *Repository) ListAllTransactions(limit, offset int, status string) ([]models.Transaction, int64, error) {
	var txs []models.Transaction
	var total int64

	base := r.db.Model(&models.Transaction{}).Where("deleted_at IS NULL")
	if status != "" {
		base = base.Where("status = ?", status)
	}

	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := base.Order("created_at DESC").Limit(limit).Offset(offset).Find(&txs).Error
	return txs, total, err
}

func (r *Repository) TopTraders(limit int) ([]map[string]any, error) {
	var results []struct {
		DisplayName string
		Volume      float64
		TxCount     int64
	}
	// Only include users who have explicitly opted in (show_in_leaderboard = true).
	// The full Stellar address is never returned; opted-in users show their alias
	// or an abbreviated form (first4...last4) if no alias is set.
	err := r.db.Raw(`
		SELECT
			CASE
				WHEN u.leaderboard_alias IS NOT NULL AND u.leaderboard_alias <> ''
					THEN u.leaderboard_alias
				ELSE LEFT(t.sender, 4) || '...' || RIGHT(t.sender, 4)
			END AS display_name,
			SUM(t.amount) AS volume,
			COUNT(*) AS tx_count
		FROM transactions t
		INNER JOIN users u
			ON u.stellar_address = t.sender
			AND u.deleted_at IS NULL
			AND u.show_in_leaderboard = true
		WHERE t.deleted_at IS NULL
		  AND t.status = 'confirmed'
		GROUP BY t.sender, u.leaderboard_alias
		ORDER BY volume DESC
		LIMIT ?
	`, limit).Scan(&results).Error
	if err != nil {
		return nil, err
	}

	out := make([]map[string]any, len(results))
	for i, row := range results {
		out[i] = map[string]any{
			"display_name":      row.DisplayName,
			"volume_stroops":    row.Volume,
			"transaction_count": row.TxCount,
		}
	}
	return out, nil
}

func (r *Repository) UpdatePreferences(userID uuid.UUID, showInLeaderboard bool, alias string) error {
	return r.db.Model(&models.User{}).
		Where("id = ?", userID).
		Updates(map[string]any{
			"show_in_leaderboard": showInLeaderboard,
			"leaderboard_alias":   alias,
		}).Error
}

var (
	ErrUserNotFound        = errors.New("user not found")
	ErrTransactionNotFound = errors.New("transaction not found")
)
