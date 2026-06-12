package auth

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/nodus-protocol/backend/internal/utils"
	"go.uber.org/zap"
)

// Sep10Handler handles SEP-10 Stellar Web Authentication endpoints.
type Sep10Handler struct {
	svc    *Service
	sep10  *utils.Sep10Manager
	log    *zap.Logger
}

// NewSep10Handler creates a new Sep10Handler.
func NewSep10Handler(svc *Service, sep10 *utils.Sep10Manager, log *zap.Logger) *Sep10Handler {
	return &Sep10Handler{svc: svc, sep10: sep10, log: log}
}

type stellarTokenRequest struct {
	Transaction string `json:"transaction" validate:"required"`
}

// StellarChallenge godoc
// @Summary      Issue a SEP-10 challenge transaction
// @Description  Returns a Stellar transaction that the client must sign with their keypair to prove account ownership.
// @Tags         auth
// @Produce      json
// @Param        account query string true "Stellar account ID (G...)"
// @Success      200 {object} utils.Response
// @Failure      400 {object} utils.Response
// @Router       /auth/stellar/challenge [get]
func (h *Sep10Handler) StellarChallenge(c *gin.Context) {
	accountID := strings.TrimSpace(c.Query("account"))
	if accountID == "" {
		utils.BadRequest(c, "MISSING_ACCOUNT", "account query parameter is required", nil)
		return
	}

	xdr, err := h.sep10.BuildChallenge(accountID)
	if err != nil {
		h.log.Warn("sep10 challenge build failed", zap.String("account", accountID), zap.Error(err))
		utils.BadRequest(c, "INVALID_ACCOUNT", err.Error(), nil)
		return
	}

	utils.OK(c, "challenge issued", gin.H{
		"transaction":    xdr,
		"network":        "stellar",
		"server_account": h.sep10.ServerAddress(),
	})
}

// StellarToken godoc
// @Summary      Exchange a signed SEP-10 challenge for a JWT token pair
// @Description  Client signs the challenge transaction and submits it. Server verifies and returns access + refresh tokens.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body body stellarTokenRequest true "Signed challenge XDR"
// @Success      200 {object} utils.Response
// @Failure      400 {object} utils.Response
// @Failure      401 {object} utils.Response
// @Router       /auth/stellar/token [post]
func (h *Sep10Handler) StellarToken(c *gin.Context) {
	var req stellarTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "INVALID_BODY", "malformed request body", nil)
		return
	}
	if strings.TrimSpace(req.Transaction) == "" {
		utils.BadRequest(c, "MISSING_TRANSACTION", "transaction field is required", nil)
		return
	}

	pair, user, err := h.svc.StellarToken(req.Transaction)
	if err != nil {
		h.log.Warn("sep10 token exchange failed", zap.Error(err))
		utils.Unauthorized(c, "stellar authentication failed")
		return
	}

	utils.OK(c, "stellar authentication successful", gin.H{
		"user":   user,
		"tokens": pair,
	})
}
