package users

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/nodus-protocol/backend/internal/models"
	"github.com/nodus-protocol/backend/internal/utils"
	"go.uber.org/zap"
)

// LPFetcher is implemented by pool.Service to avoid a circular import.
type LPFetcher interface {
	GetLPPosition(ctx context.Context, address string) (map[string]any, error)
}

type Service struct {
	repo      *Repository
	lpFetcher LPFetcher
	log       *zap.Logger
}

func NewService(repo *Repository, log *zap.Logger) *Service {
	return &Service{repo: repo, log: log}
}

func (s *Service) WithLPFetcher(f LPFetcher) *Service {
	s.lpFetcher = f
	return s
}

// ── User profile ──────────────────────────────────────────────────────────────

func (s *Service) GetProfile(userID uuid.UUID) (*models.User, error) {
	user, err := s.repo.FindByID(userID)
	if err != nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

func (s *Service) UpdateProfile(userID uuid.UUID, firstName, lastName, avatarURL string) (*models.User, error) {
	user, err := s.repo.FindByID(userID)
	if err != nil {
		return nil, ErrUserNotFound
	}
	if firstName != "" {
		user.FirstName = firstName
	}
	if lastName != "" {
		user.LastName = lastName
	}
	if avatarURL != "" {
		user.AvatarURL = avatarURL
	}
	if err := s.repo.Update(user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}
	s.log.Info("user profile updated", zap.String("user_id", userID.String()))
	return user, nil
}

func (s *Service) ChangePassword(userID uuid.UUID, oldPassword, newPassword string) error {
	user, err := s.repo.FindByID(userID)
	if err != nil {
		return ErrUserNotFound
	}
	if !utils.CheckPassword(oldPassword, user.PasswordHash) {
		return ErrInvalidPassword
	}
	newHash, err := utils.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("password hashing failed: %w", err)
	}
	user.PasswordHash = newHash
	return s.repo.Update(user)
}

func (s *Service) DeleteAccount(userID uuid.UUID) error {
	return s.repo.SoftDelete(userID)
}

// ── Wallet linking ────────────────────────────────────────────────────────────

func (s *Service) LinkWallet(userID uuid.UUID, address string) (*models.User, error) {
	address = strings.TrimSpace(address)
	if len(address) != 56 || !strings.HasPrefix(address, "G") {
		return nil, ErrInvalidStellarAddress
	}
	if err := s.repo.UpdateStellarAddress(userID, address); err != nil {
		return nil, fmt.Errorf("failed to link wallet: %w", err)
	}
	return s.repo.FindByID(userID)
}

func (s *Service) UnlinkWallet(userID uuid.UUID) error {
	return s.repo.UpdateStellarAddress(userID, "")
}

// ── LP position ───────────────────────────────────────────────────────────────

func (s *Service) GetLPPosition(ctx context.Context, userID uuid.UUID) (map[string]any, error) {
	user, err := s.repo.FindByID(userID)
	if err != nil {
		return nil, ErrUserNotFound
	}
	if user.StellarAddress == "" {
		return nil, ErrNoWalletLinked
	}
	if s.lpFetcher == nil {
		return nil, fmt.Errorf("LP fetcher not configured")
	}
	return s.lpFetcher.GetLPPosition(ctx, user.StellarAddress)
}

// ── Transactions ──────────────────────────────────────────────────────────────

func (s *Service) ListTransactions(userID uuid.UUID, page, limit int, status, token string) (map[string]any, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if page <= 0 {
		page = 1
	}
	txs, total, err := s.repo.ListTransactionsByUser(userID, limit, (page-1)*limit, status, token)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"transactions": txs,
		"total":        total,
		"page":         page,
		"limit":        limit,
		"pages":        (total + int64(limit) - 1) / int64(limit),
	}, nil
}

func (s *Service) GetTransaction(userID uuid.UUID, txIDStr string) (*models.Transaction, error) {
	txID, err := uuid.Parse(txIDStr)
	if err != nil {
		return nil, ErrTransactionNotFound
	}
	return s.repo.FindTransactionByID(txID, userID)
}

// ── GDPR export ───────────────────────────────────────────────────────────────

func (s *Service) ExportData(userID uuid.UUID) (map[string]any, error) {
	user, err := s.repo.FindByID(userID)
	if err != nil {
		return nil, ErrUserNotFound
	}
	txs, _, _ := s.repo.ListTransactionsByUser(userID, 1000, 0, "", "")
	return map[string]any{
		"profile":      user,
		"transactions": txs,
	}, nil
}

// ── Admin ─────────────────────────────────────────────────────────────────────

func (s *Service) AdminListUsers(page, limit int, search string) (map[string]any, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if page <= 0 {
		page = 1
	}
	users, total, err := s.repo.FindAll(limit, (page-1)*limit, search)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"users": users,
		"total": total,
		"page":  page,
		"limit": limit,
		"pages": (total + int64(limit) - 1) / int64(limit),
	}, nil
}

func (s *Service) AdminGetUser(userIDStr string) (*models.User, error) {
	id, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, ErrUserNotFound
	}
	return s.repo.FindByID(id)
}

func (s *Service) AdminUpdateRole(userIDStr, role string) (*models.User, error) {
	id, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, ErrUserNotFound
	}
	var r models.UserRole
	switch role {
	case "admin":
		r = models.RoleAdmin
	case "user":
		r = models.RoleUser
	default:
		return nil, ErrInvalidRole
	}
	if err := s.repo.UpdateRole(id, r); err != nil {
		return nil, err
	}
	return s.repo.FindByID(id)
}

func (s *Service) AdminHardDelete(userIDStr string) error {
	id, err := uuid.Parse(userIDStr)
	if err != nil {
		return ErrUserNotFound
	}
	return s.repo.HardDelete(id)
}

func (s *Service) AdminStats() (map[string]any, error) {
	return s.repo.ProtocolStats()
}

func (s *Service) AdminListTransactions(page, limit int, status string) (map[string]any, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if page <= 0 {
		page = 1
	}
	txs, total, err := s.repo.ListAllTransactions(limit, (page-1)*limit, status)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"transactions": txs,
		"total":        total,
		"page":         page,
		"limit":        limit,
		"pages":        (total + int64(limit) - 1) / int64(limit),
	}, nil
}

// ── Leaderboard ───────────────────────────────────────────────────────────────

func (s *Service) LeaderboardTraders(limit int) ([]map[string]any, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	return s.repo.TopTraders(limit)
}

var (
	ErrInvalidPassword       = errors.New("current password is incorrect")
	ErrNoWalletLinked        = errors.New("no stellar wallet linked to this account")
	ErrInvalidStellarAddress = errors.New("invalid stellar address: must be 56 characters starting with G")
	ErrInvalidRole           = errors.New("invalid role: must be 'user' or 'admin'")
)
