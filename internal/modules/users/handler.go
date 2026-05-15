package users

import (
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/nodus-protocol/backend/internal/middleware"
	"github.com/nodus-protocol/backend/internal/utils"
	"go.uber.org/zap"
)

// Handler holds HTTP handlers for the users domain.
type Handler struct {
	svc      *Service
	log      *zap.Logger
	validate *validator.Validate
}

func NewHandler(svc *Service, log *zap.Logger) *Handler {
	return &Handler{svc: svc, log: log, validate: validator.New()}
}

// ---- DTOs ----

type updateProfileRequest struct {
	FirstName string `json:"first_name" validate:"omitempty,max=100"`
	LastName  string `json:"last_name"  validate:"omitempty,max=100"`
	AvatarURL string `json:"avatar_url" validate:"omitempty,url,max=512"`
}

type changePasswordRequest struct {
	OldPassword string `json:"old_password" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8,max=72"`
}

// ---- Handlers ----

// GetMe godoc
// @Summary      Get current user profile
// @Tags         users
// @Security     BearerAuth
// @Produce      json
// @Success      200 {object} utils.Response
// @Failure      401 {object} utils.Response
// @Router       /users/me [get]
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

// UpdateMe godoc
// @Summary      Update current user profile
// @Tags         users
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body body updateProfileRequest true "Update payload"
// @Success      200 {object} utils.Response
// @Router       /users/me [put]
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

// ChangePassword godoc
// @Summary      Change current user password
// @Tags         users
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body body changePasswordRequest true "Password change payload"
// @Success      200 {object} utils.Response
// @Router       /users/me/password [put]
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

// DeleteMe godoc
// @Summary      Delete current user account
// @Tags         users
// @Security     BearerAuth
// @Produce      json
// @Success      200 {object} utils.Response
// @Router       /users/me [delete]
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

// ---- helpers ----

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

// suppress unused import warning
var _ = strings.TrimSpace
