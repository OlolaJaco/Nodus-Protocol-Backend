package users

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/nodus-protocol/backend/internal/middleware"
	"github.com/nodus-protocol/backend/internal/utils"
	"go.uber.org/zap"
)

type Handler struct {
	svc      *Service
	log      *zap.Logger
	validate *validator.Validate
}

func NewHandler(svc *Service, log *zap.Logger) *Handler {
	return &Handler{svc: svc, log: log, validate: validator.New()}
}

// ── DTOs ──────────────────────────────────────────────────────────────────────

type updateProfileRequest struct {
	FirstName string `json:"first_name" validate:"omitempty,max=100"`
	LastName  string `json:"last_name"  validate:"omitempty,max=100"`
	AvatarURL string `json:"avatar_url" validate:"omitempty,url,max=512"`
}

type changePasswordRequest struct {
	OldPassword string `json:"old_password" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8,max=72"`
}

type linkWalletRequest struct {
	StellarAddress string `json:"stellar_address" validate:"required"`
}

type adminUpdateRoleRequest struct {
	Role string `json:"role" validate:"required,oneof=user admin"`
}

type updatePreferencesRequest struct {
	ShowInLeaderboard bool   `json:"show_in_leaderboard"`
	LeaderboardAlias  string `json:"leaderboard_alias" validate:"omitempty,max=32,alphanum"`
}

// ── Self endpoints ────────────────────────────────────────────────────────────

func (h *Handler) GetMe(c *gin.Context) {
	userID, ok := extractUserID(c)
	if !ok {
		utils.Unauthorized(c, "not authenticated")
		return
	}
	user, err := h.svc.GetProfile(userID)
	if err != nil {
		utils.NotFound(c, "user")
		return
	}
	utils.OK(c, "profile retrieved", user)
}

func (h *Handler) UpdateMe(c *gin.Context) {
	userID, ok := extractUserID(c)
	if !ok {
		utils.Unauthorized(c, "not authenticated")
		return
	}
	var req updateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "INVALID_BODY", "malformed request body", nil)
		return
	}
	user, err := h.svc.UpdateProfile(userID, req.FirstName, req.LastName, req.AvatarURL)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			utils.NotFound(c, "user")
			return
		}
		h.log.Error("update profile failed", zap.Error(err))
		utils.InternalServerError(c, "failed to update profile")
		return
	}
	utils.OK(c, "profile updated", user)
}

func (h *Handler) ChangePassword(c *gin.Context) {
	userID, ok := extractUserID(c)
	if !ok {
		utils.Unauthorized(c, "not authenticated")
		return
	}
	var req changePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "INVALID_BODY", "malformed request body", nil)
		return
	}
	if errs := h.validateRequest(&req); errs != nil {
		utils.UnprocessableEntity(c, "VALIDATION_ERROR", "validation failed", errs)
		return
	}
	if err := h.svc.ChangePassword(userID, req.OldPassword, req.NewPassword); err != nil {
		if errors.Is(err, ErrInvalidPassword) {
			utils.BadRequest(c, "WRONG_PASSWORD", "current password is incorrect", nil)
			return
		}
		utils.InternalServerError(c, "failed to change password")
		return
	}
	utils.OK(c, "password changed successfully", nil)
}

func (h *Handler) DeleteMe(c *gin.Context) {
	userID, ok := extractUserID(c)
	if !ok {
		utils.Unauthorized(c, "not authenticated")
		return
	}
	if err := h.svc.DeleteAccount(userID); err != nil {
		h.log.Error("delete account failed", zap.Error(err))
		utils.InternalServerError(c, "failed to delete account")
		return
	}
	utils.OK(c, "account deleted", nil)
}

// ── Wallet endpoints ──────────────────────────────────────────────────────────

func (h *Handler) LinkWallet(c *gin.Context) {
	userID, ok := extractUserID(c)
	if !ok {
		utils.Unauthorized(c, "not authenticated")
		return
	}
	var req linkWalletRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "INVALID_BODY", "malformed request body", nil)
		return
	}
	user, err := h.svc.LinkWallet(userID, req.StellarAddress)
	if err != nil {
		if errors.Is(err, ErrInvalidStellarAddress) {
			utils.BadRequest(c, "INVALID_ADDRESS", err.Error(), nil)
			return
		}
		h.log.Error("link wallet failed", zap.Error(err))
		utils.InternalServerError(c, "failed to link wallet")
		return
	}
	utils.OK(c, "wallet linked", user)
}

func (h *Handler) UnlinkWallet(c *gin.Context) {
	userID, ok := extractUserID(c)
	if !ok {
		utils.Unauthorized(c, "not authenticated")
		return
	}
	if err := h.svc.UnlinkWallet(userID); err != nil {
		utils.InternalServerError(c, "failed to unlink wallet")
		return
	}
	utils.OK(c, "wallet unlinked", nil)
}

// ── LP position ───────────────────────────────────────────────────────────────

func (h *Handler) GetLPPosition(c *gin.Context) {
	userID, ok := extractUserID(c)
	if !ok {
		utils.Unauthorized(c, "not authenticated")
		return
	}
	pos, err := h.svc.GetLPPosition(c.Request.Context(), userID)
	if err != nil {
		switch {
		case errors.Is(err, ErrUserNotFound):
			utils.NotFound(c, "user")
		case errors.Is(err, ErrNoWalletLinked):
			utils.BadRequest(c, "NO_WALLET", err.Error(), nil)
		default:
			h.log.Error("lp position failed", zap.Error(err))
			utils.InternalServerError(c, "failed to fetch LP position")
		}
		return
	}
	utils.OK(c, "LP position retrieved", pos)
}

// ── Transaction history ───────────────────────────────────────────────────────

func (h *Handler) ListTransactions(c *gin.Context) {
	userID, ok := extractUserID(c)
	if !ok {
		utils.Unauthorized(c, "not authenticated")
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	status := c.Query("status")
	token := c.Query("token")

	result, err := h.svc.ListTransactions(userID, page, limit, status, token)
	if err != nil {
		utils.InternalServerError(c, "failed to list transactions")
		return
	}
	utils.OK(c, "transactions retrieved", result)
}

func (h *Handler) GetTransaction(c *gin.Context) {
	userID, ok := extractUserID(c)
	if !ok {
		utils.Unauthorized(c, "not authenticated")
		return
	}
	tx, err := h.svc.GetTransaction(userID, c.Param("id"))
	if err != nil {
		if errors.Is(err, ErrTransactionNotFound) {
			utils.NotFound(c, "transaction")
			return
		}
		utils.InternalServerError(c, "failed to fetch transaction")
		return
	}
	utils.OK(c, "transaction retrieved", tx)
}

// ── GDPR export ───────────────────────────────────────────────────────────────

func (h *Handler) ExportData(c *gin.Context) {
	userID, ok := extractUserID(c)
	if !ok {
		utils.Unauthorized(c, "not authenticated")
		return
	}
	data, err := h.svc.ExportData(userID)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			utils.NotFound(c, "user")
			return
		}
		utils.InternalServerError(c, "export failed")
		return
	}
	utils.OK(c, "data export ready", data)
}

// ── Preferences ───────────────────────────────────────────────────────────────

func (h *Handler) UpdatePreferences(c *gin.Context) {
	userID, ok := extractUserID(c)
	if !ok {
		utils.Unauthorized(c, "not authenticated")
		return
	}
	var req updatePreferencesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "INVALID_BODY", "malformed request body", nil)
		return
	}
	if errs := h.validateRequest(&req); errs != nil {
		utils.UnprocessableEntity(c, "VALIDATION_ERROR", "validation failed", errs)
		return
	}
	user, err := h.svc.UpdatePreferences(userID, req.ShowInLeaderboard, req.LeaderboardAlias)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			utils.NotFound(c, "user")
			return
		}
		h.log.Error("update preferences failed", zap.Error(err))
		utils.InternalServerError(c, "failed to update preferences")
		return
	}
	utils.OK(c, "preferences updated", user)
}

// ── Admin endpoints ───────────────────────────────────────────────────────────

func (h *Handler) AdminListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	search := c.Query("search")

	result, err := h.svc.AdminListUsers(page, limit, search)
	if err != nil {
		utils.InternalServerError(c, "failed to list users")
		return
	}
	utils.OK(c, "users retrieved", result)
}

func (h *Handler) AdminGetUser(c *gin.Context) {
	user, err := h.svc.AdminGetUser(c.Param("id"))
	if err != nil {
		utils.NotFound(c, "user")
		return
	}
	utils.OK(c, "user retrieved", user)
}

func (h *Handler) AdminUpdateRole(c *gin.Context) {
	var req adminUpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "INVALID_BODY", "malformed request body", nil)
		return
	}
	user, err := h.svc.AdminUpdateRole(c.Param("id"), req.Role)
	if err != nil {
		if errors.Is(err, ErrInvalidRole) {
			utils.BadRequest(c, "INVALID_ROLE", err.Error(), nil)
			return
		}
		if errors.Is(err, ErrUserNotFound) {
			utils.NotFound(c, "user")
			return
		}
		utils.InternalServerError(c, "failed to update role")
		return
	}
	utils.OK(c, "role updated", user)
}

func (h *Handler) AdminDeleteUser(c *gin.Context) {
	if err := h.svc.AdminHardDelete(c.Param("id")); err != nil {
		if errors.Is(err, ErrUserNotFound) {
			utils.NotFound(c, "user")
			return
		}
		utils.InternalServerError(c, "failed to delete user")
		return
	}
	utils.OK(c, "user deleted", nil)
}

func (h *Handler) AdminStats(c *gin.Context) {
	stats, err := h.svc.AdminStats()
	if err != nil {
		utils.InternalServerError(c, "failed to fetch stats")
		return
	}
	utils.OK(c, "protocol stats retrieved", stats)
}

func (h *Handler) AdminListTransactions(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	status := c.Query("status")

	result, err := h.svc.AdminListTransactions(page, limit, status)
	if err != nil {
		utils.InternalServerError(c, "failed to list transactions")
		return
	}
	utils.OK(c, "transactions retrieved", result)
}

// ── Leaderboard ───────────────────────────────────────────────────────────────

func (h *Handler) LeaderboardTraders(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	traders, err := h.svc.LeaderboardTraders(limit)
	if err != nil {
		utils.InternalServerError(c, "failed to fetch leaderboard")
		return
	}
	utils.OK(c, "leaderboard retrieved", traders)
}

// ── helpers ───────────────────────────────────────────────────────────────────

func extractUserID(c *gin.Context) (uuid.UUID, bool) {
	raw, exists := c.Get(middleware.ContextKeyUserID)
	if !exists {
		return uuid.Nil, false
	}
	id, err := uuid.Parse(raw.(string))
	if err != nil {
		return uuid.Nil, false
	}
	return id, true
}

func (h *Handler) validateRequest(req interface{}) []string {
	if err := h.validate.Struct(req); err != nil {
		var validationErrors validator.ValidationErrors
		if errors.As(err, &validationErrors) {
			msgs := make([]string, 0, len(validationErrors))
			for _, e := range validationErrors {
				msgs = append(msgs, e.Field()+" "+e.Tag())
			}
			return msgs
		}
	}
	return nil
}
