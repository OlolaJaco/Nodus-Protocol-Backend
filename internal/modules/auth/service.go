package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nodus-protocol/backend/internal/config"
	"github.com/nodus-protocol/backend/internal/models"
	"github.com/nodus-protocol/backend/internal/utils"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// TokenPair holds both JWT tokens returned on login/refresh.
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"` // access token TTL in seconds
}

// Service contains all auth business logic.
type Service struct {
	repo       *Repository
	jwt        *utils.JWTManager
	rdb        *redis.Client
	mailer     *utils.Mailer
	cfg        *config.Config
	log        *zap.Logger
}

// NewService creates a new auth Service.
func NewService(repo *Repository, jwt *utils.JWTManager, rdb *redis.Client, mailer *utils.Mailer, cfg *config.Config, log *zap.Logger) *Service {
	return &Service{repo: repo, jwt: jwt, rdb: rdb, mailer: mailer, cfg: cfg, log: log}
}

// Register creates a new user, hashes password, and sends verification email.
func (s *Service) Register(email, password, firstName, lastName string) (*models.User, error) {
	// Check if email already exists
	existing, err := s.repo.FindUserByEmail(email)
	if err == nil && existing != nil {
		return nil, ErrEmailAlreadyTaken
	}

	hash, err := utils.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("password hashing failed: %w", err)
	}

	user := &models.User{
		Email:        email,
		PasswordHash: hash,
		FirstName:    firstName,
		LastName:     lastName,
		Role:         models.RoleUser,
	}

	if err := s.repo.CreateUser(user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Send verification email (non-blocking: log error but don't fail registration)
	go func() {
		if verifyErr := s.sendVerificationEmail(user); verifyErr != nil {
			s.log.Error("verification email failed", zap.String("user_id", user.ID.String()), zap.Error(verifyErr))
		}
	}()

	s.log.Info("user registered", zap.String("user_id", user.ID.String()), zap.String("email", email))
	return user, nil
}

// Login validates credentials and returns a JWT token pair.
func (s *Service) Login(email, password string) (*TokenPair, *models.User, error) {
	user, err := s.repo.FindUserByEmail(email)
	if err != nil {
		return nil, nil, ErrInvalidCredentials
	}

	if !utils.CheckPassword(password, user.PasswordHash) {
		return nil, nil, ErrInvalidCredentials
	}

	if !user.IsActive {
		return nil, nil, ErrAccountDisabled
	}

	pair, err := s.issueTokenPair(user)
	if err != nil {
		return nil, nil, err
	}

	_ = s.repo.UpdateUserLastLogin(user.ID)
	s.log.Info("user logged in", zap.String("user_id", user.ID.String()))
	return pair, user, nil
}

// RefreshTokens validates a refresh JWT and issues a new token pair (rotation).
func (s *Service) RefreshTokens(refreshTokenStr string) (*TokenPair, error) {
	claims, err := s.jwt.ValidateRefreshToken(refreshTokenStr)
	if err != nil {
		return nil, ErrInvalidRefreshToken
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return nil, ErrInvalidRefreshToken
	}

	user, err := s.repo.FindUserByID(userID)
	if err != nil {
		return nil, ErrInvalidRefreshToken
	}

	// Invalidate the old refresh token's hash in Redis to prevent reuse
	jti, _ := s.jwt.ExtractJTI(refreshTokenStr)
	if jti != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		s.rdb.Set(ctx, "blacklist:"+jti, "1", s.cfg.JWT.RefreshTokenTTL)
	}

	pair, err := s.issueTokenPair(user)
	if err != nil {
		return nil, err
	}

	s.log.Info("tokens refreshed", zap.String("user_id", user.ID.String()))
	return pair, nil
}

// Logout blacklists the access token JTI in Redis so it can't be reused.
func (s *Service) Logout(accessTokenStr string) error {
	jti, err := s.jwt.ExtractJTI(accessTokenStr)
	if err != nil {
		return nil // not a valid token anyway
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Blacklist until the access token would naturally expire
	return s.rdb.Set(ctx, "blacklist:"+jti, "1", s.cfg.JWT.AccessTokenTTL).Err()
}

// SendVerificationEmail generates a new verification token and (re)sends the email.
func (s *Service) SendVerificationEmail(email string) error {
	user, err := s.repo.FindUserByEmail(email)
	if err != nil {
		return nil // don't leak whether email exists
	}
	if user.IsEmailVerified {
		return ErrEmailAlreadyVerified
	}
	return s.sendVerificationEmail(user)
}

// VerifyEmail validates the verification token and marks the user's email verified.
func (s *Service) VerifyEmail(rawToken string) error {
	hash, err := utils.HashPassword(rawToken)
	if err != nil {
		return ErrInvalidToken
	}

	token, err := s.repo.FindValidToken(hash, models.TokenTypeEmailVerify)
	if err != nil {
		// Fallback: try raw token comparison (token stored as bcrypt hash)
		return ErrInvalidToken
	}

	if err := s.repo.MarkTokenUsed(token.ID); err != nil {
		return fmt.Errorf("failed to mark token used: %w", err)
	}

	return s.repo.MarkEmailVerified(token.UserID)
}

// ForgotPassword generates a reset token and emails it to the user.
func (s *Service) ForgotPassword(email string) error {
	user, err := s.repo.FindUserByEmail(email)
	if err != nil {
		return nil // don't leak whether email exists
	}
	return s.sendPasswordResetEmail(user)
}

// ResetPassword validates the reset token and updates the user's password.
func (s *Service) ResetPassword(rawToken, newPassword string) error {
	tokenHash, err := utils.HashPassword(rawToken)
	if err != nil {
		return ErrInvalidToken
	}

	token, err := s.repo.FindValidToken(tokenHash, models.TokenTypePasswordReset)
	if err != nil {
		return ErrInvalidToken
	}

	newHash, err := utils.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("password hashing failed: %w", err)
	}

	if err := s.repo.UpdatePassword(token.UserID, newHash); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	_ = s.repo.MarkTokenUsed(token.ID)

	// Invalidate all refresh tokens for this user after password reset
	_ = s.repo.DeleteRefreshTokensForUser(token.UserID)

	s.log.Info("password reset successful", zap.String("user_id", token.UserID.String()))
	return nil
}

// ---- private helpers ----

func (s *Service) issueTokenPair(user *models.User) (*TokenPair, error) {
	accessToken, err := s.jwt.GenerateAccessToken(user.ID, user.Email, string(user.Role))
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := s.jwt.GenerateRefreshToken(user.ID, user.Email, string(user.Role))
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.cfg.JWT.AccessTokenTTL.Seconds()),
	}, nil
}

func (s *Service) sendVerificationEmail(user *models.User) error {
	rawToken, err := utils.GenerateSecureToken(32)
	if err != nil {
		return err
	}

	hash, err := utils.HashPassword(rawToken)
	if err != nil {
		return err
	}

	dbToken := &models.Token{
		UserID:    user.ID,
		Hash:      hash,
		Type:      models.TokenTypeEmailVerify,
		ExpiresAt: time.Now().Add(s.cfg.JWT.EmailTokenTTL),
	}
	if err := s.repo.CreateToken(dbToken); err != nil {
		return err
	}

	verifyURL := fmt.Sprintf("%s/verify-email?token=%s", s.cfg.App.FrontendURL, rawToken)
	return s.mailer.SendVerificationEmail(user.Email, user.FullName(), verifyURL)
}

func (s *Service) sendPasswordResetEmail(user *models.User) error {
	rawToken, err := utils.GenerateSecureToken(32)
	if err != nil {
		return err
	}

	hash, err := utils.HashPassword(rawToken)
	if err != nil {
		return err
	}

	dbToken := &models.Token{
		UserID:    user.ID,
		Hash:      hash,
		Type:      models.TokenTypePasswordReset,
		ExpiresAt: time.Now().Add(s.cfg.JWT.PasswordResetTTL),
	}
	if err := s.repo.CreateToken(dbToken); err != nil {
		return err
	}

	resetURL := fmt.Sprintf("%s/reset-password?token=%s", s.cfg.App.FrontendURL, rawToken)
	return s.mailer.SendPasswordResetEmail(user.Email, user.FullName(), resetURL)
}

// Sentinel service errors
var (
	ErrEmailAlreadyTaken    = errors.New("email is already registered")
	ErrInvalidCredentials   = errors.New("invalid email or password")
	ErrAccountDisabled      = errors.New("account is disabled")
	ErrInvalidRefreshToken  = errors.New("invalid or expired refresh token")
	ErrInvalidToken         = errors.New("invalid or expired token")
	ErrEmailAlreadyVerified = errors.New("email is already verified")
)
