package auth

import (
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/nodus-protocol/backend/internal/middleware"
	"github.com/nodus-protocol/backend/internal/utils"
	"go.uber.org/zap"
)

// Handler holds the HTTP handlers for auth endpoints.
type Handler struct {
	svc      *Service
	log      *zap.Logger
	validate *validator.Validate
}

// NewHandler creates a new auth Handler.
func NewHandler(svc *Service, log *zap.Logger) *Handler {
	return &Handler{svc: svc, log: log, validate: validator.New()}
}

// ---- Request DTOs ----

type registerRequest struct {
	Email     string `json:"email"      validate:"required,email,max=255"`
	Password  string `json:"password"   validate:"required,min=10,max=72"`
	FirstName string `json:"first_name" validate:"required,max=100"`
	LastName  string `json:"last_name"  validate:"required,max=100"`
}

type loginRequest struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type verifyEmailRequest struct {
	Token string `form:"token" validate:"required"`
}

type resendVerificationRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type forgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type resetPasswordRequest struct {
	Token       string `json:"token"        validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=10,max=72"`
}

// ---- Handlers ----

// Register godoc
// @Summary      Register a new user
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body body registerRequest true "Registration payload"
// @Success      201  {object} utils.Response
// @Failure      400  {object} utils.Response
// @Failure      409  {object} utils.Response
// @Router       /auth/register [post]
func (h *Handler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "INVALID_BODY", "malformed request body", nil)
		return
	}
	if errs := h.validateRequest(&req); errs != nil {
		utils.UnprocessableEntity(c, "VALIDATION_ERROR", "validation failed", errs)
		return
	}

	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	user, err := h.svc.Register(req.Email, req.Password, req.FirstName, req.LastName)
	if err != nil {
		var validErr *ValidationError
		if errors.As(err, &validErr) {
			utils.BadRequest(c, validErr.Code, validErr.Message, gin.H{
				"field": validErr.Field,
			})
			return
		}
		if errors.Is(err, ErrEmailAlreadyTaken) {
			utils.Conflict(c, "email is already registered")
			return
		}
		h.log.Error("register failed", zap.Error(err))
		utils.InternalServerError(c, "registration failed")
		return
	}

	utils.Created(c, "registration successful, please verify your email", gin.H{
		"user": sanitizeUser(user),
	})
}

// Login godoc
// @Summary      Login with email and password
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body body loginRequest true "Login payload"
// @Success      200  {object} utils.Response
// @Failure      401  {object} utils.Response
// @Router       /auth/login [post]
func (h *Handler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "INVALID_BODY", "malformed request body", nil)
		return
	}
	if errs := h.validateRequest(&req); errs != nil {
		utils.UnprocessableEntity(c, "VALIDATION_ERROR", "validation failed", errs)
		return
	}

	pair, user, err := h.svc.Login(strings.ToLower(strings.TrimSpace(req.Email)), req.Password)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) || errors.Is(err, ErrAccountDisabled) {
			utils.Unauthorized(c, "invalid email or password")
			return
		}
		h.log.Error("login failed", zap.Error(err))
		utils.InternalServerError(c, "login failed")
		return
	}

	utils.OK(c, "login successful", gin.H{
		"user":   sanitizeUser(user),
		"tokens": pair,
	})
}

// RefreshToken godoc
// @Summary      Refresh access token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body body refreshRequest true "Refresh token"
// @Success      200  {object} utils.Response
// @Failure      401  {object} utils.Response
// @Router       /auth/refresh [post]
func (h *Handler) RefreshToken(c *gin.Context) {
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "INVALID_BODY", "malformed request body", nil)
		return
	}

	pair, err := h.svc.RefreshTokens(req.RefreshToken)
	if err != nil {
		utils.Unauthorized(c, "invalid or expired refresh token")
		return
	}

	utils.OK(c, "tokens refreshed", gin.H{"tokens": pair})
}

// Logout godoc
// @Summary      Logout and invalidate access token
// @Tags         auth
// @Security     BearerAuth
// @Produce      json
// @Success      200 {object} utils.Response
// @Router       /auth/logout [post]
func (h *Handler) Logout(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) == 2 {
		_ = h.svc.Logout(parts[1])
	}
	utils.OK(c, "logged out successfully", nil)
}

// VerifyEmail godoc
// @Summary      Verify email address
// @Tags         auth
// @Param        token query string true "Verification token"
// @Success      200   {object} utils.Response
// @Failure      400   {object} utils.Response
// @Router       /auth/verify-email [get]
func (h *Handler) VerifyEmail(c *gin.Context) {
	var req verifyEmailRequest
	if err := c.ShouldBindQuery(&req); err != nil || req.Token == "" {
		utils.BadRequest(c, "MISSING_TOKEN", "verification token is required", nil)
		return
	}

	if err := h.svc.VerifyEmail(req.Token); err != nil {
		utils.BadRequest(c, "INVALID_TOKEN", "invalid or expired verification token", nil)
		return
	}

	utils.OK(c, "email verified successfully", nil)
}

// ResendVerification godoc
// @Summary      Resend email verification
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body body resendVerificationRequest true "Email"
// @Success      200 {object} utils.Response
// @Router       /auth/resend-verification [post]
func (h *Handler) ResendVerification(c *gin.Context) {
	var req resendVerificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "INVALID_BODY", "malformed request body", nil)
		return
	}

	// Always respond 200 to prevent email enumeration
	_ = h.svc.SendVerificationEmail(strings.ToLower(strings.TrimSpace(req.Email)))
	utils.OK(c, "if your email is registered, a verification link has been sent", nil)
}

// ForgotPassword godoc
// @Summary      Request password reset email
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body body forgotPasswordRequest true "Email"
// @Success      200 {object} utils.Response
// @Router       /auth/forgot-password [post]
func (h *Handler) ForgotPassword(c *gin.Context) {
	var req forgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "INVALID_BODY", "malformed request body", nil)
		return
	}

	// Always 200 — never leak whether email exists
	_ = h.svc.ForgotPassword(strings.ToLower(strings.TrimSpace(req.Email)))
	utils.OK(c, "if your email is registered, a reset link has been sent", nil)
}

// ResetPassword godoc
// @Summary      Reset password using token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body body resetPasswordRequest true "Reset payload"
// @Success      200 {object} utils.Response
// @Failure      400 {object} utils.Response
// @Router       /auth/reset-password [post]
func (h *Handler) ResetPassword(c *gin.Context) {
	var req resetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "INVALID_BODY", "malformed request body", nil)
		return
	}
	if errs := h.validateRequest(&req); errs != nil {
		utils.UnprocessableEntity(c, "VALIDATION_ERROR", "validation failed", errs)
		return
	}

	if err := h.svc.ResetPassword(req.Token, req.NewPassword); err != nil {
		var validErr *ValidationError
		if errors.As(err, &validErr) {
			utils.BadRequest(c, validErr.Code, validErr.Message, gin.H{
				"field": validErr.Field,
			})
			return
		}
		utils.BadRequest(c, "INVALID_TOKEN", "invalid or expired reset token", nil)
		return
	}

	utils.OK(c, "password reset successful", nil)
}

// ---- helpers ----

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

func sanitizeUser(user interface{}) interface{} {
	// The User model already omits PasswordHash via json:"-", so returning as-is is safe.
	return user
}

// GetCurrentUser is a placeholder called from users module to get the authenticated user.
// @Summary      Get authenticated user context key
func GetCurrentUserID(c *gin.Context) (string, bool) {
	id, ok := c.Get(middleware.ContextKeyUserID)
	if !ok {
		return "", false
	}
	return id.(string), true
}
