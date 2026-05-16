package utils

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/nodus-protocol/backend/internal/config"
)

// Claims defines the JWT payload.
type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// JWTManager handles RS256 signing and verification.
type JWTManager struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	cfg        config.JWTConfig
}

// NewJWTManager loads RSA keys from disk and returns a JWTManager.
func NewJWTManager(cfg config.JWTConfig) (*JWTManager, error) {
	privBytes, err := os.ReadFile(cfg.PrivateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("cannot read private key %s: %w", cfg.PrivateKeyPath, err)
	}
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privBytes)
	if err != nil {
		return nil, fmt.Errorf("cannot parse private key: %w", err)
	}

	pubBytes, err := os.ReadFile(cfg.PublicKeyPath)
	if err != nil {
		return nil, fmt.Errorf("cannot read public key %s: %w", cfg.PublicKeyPath, err)
	}
	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(pubBytes)
	if err != nil {
		return nil, fmt.Errorf("cannot parse public key: %w", err)
	}

	return &JWTManager{privateKey: privateKey, publicKey: publicKey, cfg: cfg}, nil
}

// GenerateAccessToken creates a short-lived access JWT.
func (j *JWTManager) GenerateAccessToken(userID uuid.UUID, email, role string) (string, error) {
	return j.sign(userID.String(), email, role, j.cfg.AccessTokenTTL, "access")
}

// GenerateRefreshToken creates a long-lived refresh JWT (the raw token itself is not stored;
// we store an opaque identifier in the Token table instead).
func (j *JWTManager) GenerateRefreshToken(userID uuid.UUID, email, role string) (string, error) {
	return j.sign(userID.String(), email, role, j.cfg.RefreshTokenTTL, "refresh")
}

func (j *JWTManager) sign(userID, email, role string, ttl time.Duration, tokenType string) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "nodus-protocol",
			Subject:   userID,
			Audience:  jwt.ClaimStrings{tokenType},
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			ID:        uuid.New().String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(j.privateKey)
}

// ValidateAccessToken parses and validates an access JWT, returning its claims.
func (j *JWTManager) ValidateAccessToken(tokenStr string) (*Claims, error) {
	return j.validate(tokenStr, "access")
}

// ValidateRefreshToken parses and validates a refresh JWT.
func (j *JWTManager) ValidateRefreshToken(tokenStr string) (*Claims, error) {
	return j.validate(tokenStr, "refresh")
}

func (j *JWTManager) validate(tokenStr, expectedAudience string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return j.publicKey, nil
	}, jwt.WithAudience(expectedAudience))

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrTokenInvalid
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrTokenInvalid
	}

	return claims, nil
}

// ExtractJTI returns the JWT ID (jti) from a token string without full validation.
// Used to blacklist tokens on logout.
func (j *JWTManager) ExtractJTI(tokenStr string) (string, error) {
	token, _, err := jwt.NewParser().ParseUnverified(tokenStr, &Claims{})
	if err != nil {
		return "", ErrTokenInvalid
	}
	claims, ok := token.Claims.(*Claims)
	if !ok {
		return "", ErrTokenInvalid
	}
	return claims.ID, nil
}

// Sentinel errors
var (
	ErrTokenExpired = errors.New("token has expired")
	ErrTokenInvalid = errors.New("token is invalid")
)
