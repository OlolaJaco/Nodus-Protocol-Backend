package payments

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nodus-protocol/backend/internal/middleware"
	"github.com/nodus-protocol/backend/internal/utils"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Handler struct {
	svc *Service
	log *zap.Logger
}

func NewHandler(svc *Service, log *zap.Logger) *Handler {
	return &Handler{svc: svc, log: log}
}

// InitiatePayment godoc
// @Summary      Initiate a new payment via Core Engine
// @Tags         payments
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body body initiateRequest true "Payment details"
// @Success      201  {object} utils.Response
// @Failure      400  {object} utils.Response
// @Router       /payments [post]
func (h *Handler) InitiatePayment(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}

	var req initiateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "INVALID_BODY", "malformed request body", nil)
		return
	}

	tx, err := h.svc.Initiate(c.Request.Context(), userID, req)
	if err != nil {
		h.log.Error("initiate payment failed", zap.Error(err))
		utils.InternalServerError(c, "payment initiation failed: "+err.Error())
		return
	}

	utils.Created(c, "payment initiated", tx)
}

// ListPayments godoc
// @Summary      List authenticated user's transactions
// @Tags         payments
// @Security     BearerAuth
// @Produce      json
// @Param        page      query int false "Page number (default 1)"
// @Param        page_size query int false "Items per page (default 20, max 100)"
// @Success      200 {object} utils.Response
// @Router       /payments [get]
func (h *Handler) ListPayments(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	txs, total, err := h.svc.List(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		h.log.Error("list payments failed", zap.Error(err))
		utils.InternalServerError(c, "failed to fetch transactions")
		return
	}

	utils.OK(c, "transactions retrieved", gin.H{
		"transactions": txs,
		"total":        total,
		"page":         page,
		"page_size":    pageSize,
	})
}

// GetPayment godoc
// @Summary      Get a single transaction by ID
// @Tags         payments
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "Transaction ID (UUID)"
// @Success      200 {object} utils.Response
// @Failure      404 {object} utils.Response
// @Router       /payments/{id} [get]
func (h *Handler) GetPayment(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}

	txID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "INVALID_ID", "transaction id must be a valid UUID", nil)
		return
	}

	tx, err := h.svc.GetByID(c.Request.Context(), txID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.NotFound(c, "transaction")
			return
		}
		utils.InternalServerError(c, "failed to fetch transaction")
		return
	}

	utils.OK(c, "transaction retrieved", tx)
}

// SimulatePayment godoc
// @Summary      Simulate a payment (dry-run, no funds moved)
// @Tags         payments
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body body simulateRequest true "Simulation parameters"
// @Success      200 {object} utils.Response
// @Router       /payments/simulate [post]
func (h *Handler) SimulatePayment(c *gin.Context) {
	if _, ok := currentUserID(c); !ok {
		return
	}

	var req simulateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "INVALID_BODY", "malformed request body", nil)
		return
	}

	result, err := h.svc.Simulate(c.Request.Context(), req)
	if err != nil {
		utils.InternalServerError(c, "simulation failed: "+err.Error())
		return
	}

	utils.OK(c, "simulation complete", result)
}

// BatchPayments godoc
// @Summary      Submit a batch of payments (up to 100)
// @Tags         payments
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body body []batchItem true "Batch items"
// @Success      207 {object} utils.Response
// @Router       /payments/batch [post]
func (h *Handler) BatchPayments(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}

	var items []batchItem
	if err := c.ShouldBindJSON(&items); err != nil {
		utils.BadRequest(c, "INVALID_BODY", "malformed request body", nil)
		return
	}

	if len(items) == 0 {
		utils.BadRequest(c, "EMPTY_BATCH", "batch must contain at least 1 item", nil)
		return
	}

	result, err := h.svc.Batch(c.Request.Context(), userID, items)
	if err != nil {
		utils.InternalServerError(c, "batch failed: "+err.Error())
		return
	}

	utils.OK(c, "batch processed", result)
}

// GetReceipt godoc
// @Summary      Get a payment receipt (confirmed transactions only)
// @Tags         payments
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "Transaction ID (UUID)"
// @Success      200 {object} utils.Response
// @Failure      404 {object} utils.Response
// @Router       /payments/{id}/receipt [get]
func (h *Handler) GetReceipt(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		return
	}

	txID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "INVALID_ID", "transaction id must be a valid UUID", nil)
		return
	}

	receipt, err := h.svc.GetReceipt(c.Request.Context(), txID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.NotFound(c, "transaction")
			return
		}
		utils.InternalServerError(c, "failed to fetch receipt: "+err.Error())
		return
	}

	utils.OK(c, "receipt retrieved", receipt)
}

// GetFees godoc
// @Summary      Get current network fee estimates from Core Engine
// @Tags         payments
// @Security     BearerAuth
// @Produce      json
// @Success      200 {object} utils.Response
// @Router       /payments/fees [get]
func (h *Handler) GetFees(c *gin.Context) {
	fees, err := h.svc.GetFees(c.Request.Context())
	if err != nil {
		utils.InternalServerError(c, "failed to fetch fees: "+err.Error())
		return
	}
	utils.OK(c, "fees retrieved", fees)
}

// GetRates godoc
// @Summary      Get current USD exchange rates for supported tokens
// @Tags         payments
// @Produce      json
// @Param        tokens query string false "Comma-separated token symbols (e.g. XLM,USDC)"
// @Success      200 {object} utils.Response
// @Router       /payments/rates [get]
func (h *Handler) GetRates(c *gin.Context) {
	tokens := c.Query("tokens")
	rates, err := h.svc.GetRates(c.Request.Context(), tokens)
	if err != nil {
		utils.InternalServerError(c, "failed to fetch rates: "+err.Error())
		return
	}
	utils.OK(c, "rates retrieved", rates)
}

// EngineHealth godoc
// @Summary      Proxy Core Engine health check
// @Tags         payments
// @Produce      json
// @Success      200 {object} utils.Response
// @Router       /payments/engine/health [get]
func (h *Handler) EngineHealth(c *gin.Context) {
	health, err := h.svc.EngineHealth(c.Request.Context())
	if err != nil {
		utils.InternalServerError(c, "core engine unreachable: "+err.Error())
		return
	}
	utils.OK(c, "core engine status", health)
}

// WebhookHandler receives payment status updates pushed by the Core Engine.
func (h *Handler) WebhookHandler(c *gin.Context) {
	var payload struct {
		Event   string `json:"event"`
		Payment struct {
			ID     string  `json:"id"`
			Status string  `json:"status"`
			TxHash *string `json:"tx_hash"`
			Error  *string `json:"error"`
		} `json:"payment"`
	}

	if err := c.ShouldBindJSON(&payload); err != nil {
		utils.BadRequest(c, "INVALID_BODY", "malformed webhook payload", nil)
		return
	}

	txHash := ""
	if payload.Payment.TxHash != nil {
		txHash = *payload.Payment.TxHash
	}
	errMsg := ""
	if payload.Payment.Error != nil {
		errMsg = *payload.Payment.Error
	}

	if err := h.svc.repo.UpdateStatus(
		payload.Payment.ID,
		payload.Payment.Status,
		txHash,
		errMsg,
	); err != nil {
		h.log.Warn("webhook status sync failed",
			zap.String("engine_id", payload.Payment.ID),
			zap.Error(err),
		)
	}

	c.Status(200)
}

func currentUserID(c *gin.Context) (uuid.UUID, bool) {
	raw, exists := c.Get(middleware.ContextKeyUserID)
	if !exists {
		utils.Unauthorized(c, "authentication required")
		return uuid.Nil, false
	}
	id, err := uuid.Parse(raw.(string))
	if err != nil {
		utils.Unauthorized(c, "invalid user context")
		return uuid.Nil, false
	}
	return id, true
}
