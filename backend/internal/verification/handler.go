package verification

import (
	"log/slog"
	"net/mail"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/trackmycareer/app/pkg/response"
)

// Handler provides HTTP endpoints for email verification and email change.
type Handler struct {
	svc *Service
}

// NewHandler creates a new verification Handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// SendVerification handles POST /verification/send.
// Requires authentication. Sends a verification email to the current user.
func (h *Handler) SendVerification(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	if err := h.svc.SendVerification(c.Request.Context(), userID); err != nil {
		if strings.Contains(err.Error(), "email already verified") {
			response.BadRequest(c, "email already verified")
			return
		}
		slog.Error("failed to send verification email",
			"error", err.Error(),
			"user_id", userID.String(),
		)
		response.InternalError(c, err)
		return
	}

	response.OK(c, gin.H{"message": "verification email sent"})
}

type verifyEmailRequest struct {
	Token string `json:"token" binding:"required"`
}

// VerifyEmail handles POST /verification/verify.
// Public endpoint (no auth required). Verifies the user's email address.
func (h *Handler) VerifyEmail(c *gin.Context) {
	var req verifyEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "token is required")
		return
	}

	req.Token = strings.TrimSpace(req.Token)
	if req.Token == "" {
		response.BadRequest(c, "token is required")
		return
	}

	userID, err := h.svc.VerifyEmail(c.Request.Context(), req.Token)
	if err != nil {
		if strings.Contains(err.Error(), "not found or expired") {
			response.BadRequest(c, "invalid or expired verification token")
			return
		}
		slog.Error("failed to verify email",
			"error", err.Error(),
		)
		response.InternalError(c, err)
		return
	}

	slog.Info("email verified", "user_id", userID.String())
	response.OK(c, gin.H{"message": "email verified"})
}

type requestEmailChangeRequest struct {
	NewEmail string `json:"new_email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// RequestEmailChange handles POST /verification/change-email.
// Requires authentication. Initiates an email address change.
func (h *Handler) RequestEmailChange(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	var req requestEmailChangeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "new_email and password are required")
		return
	}

	req.NewEmail = strings.TrimSpace(req.NewEmail)
	if req.NewEmail == "" {
		response.BadRequest(c, "new_email is required")
		return
	}

	// Validate email format.
	if _, err := mail.ParseAddress(req.NewEmail); err != nil {
		response.BadRequest(c, "invalid email format")
		return
	}

	if err := h.svc.RequestEmailChange(c.Request.Context(), userID, req.NewEmail, req.Password); err != nil {
		switch {
		case strings.Contains(err.Error(), "incorrect password"):
			response.BadRequest(c, "incorrect password")
		case strings.Contains(err.Error(), "already in use"):
			response.BadRequest(c, "email address already in use")
		default:
			slog.Error("failed to request email change",
				"error", err.Error(),
				"user_id", userID.String(),
			)
			response.InternalError(c, err)
		}
		return
	}

	response.OK(c, gin.H{"message": "confirmation email sent to new address"})
}

type confirmEmailChangeRequest struct {
	Token string `json:"token" binding:"required"`
}

// ConfirmEmailChange handles POST /verification/confirm-change.
// Public endpoint (no auth required). Confirms an email address change.
func (h *Handler) ConfirmEmailChange(c *gin.Context) {
	var req confirmEmailChangeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "token is required")
		return
	}

	req.Token = strings.TrimSpace(req.Token)
	if req.Token == "" {
		response.BadRequest(c, "token is required")
		return
	}

	userID, err := h.svc.ConfirmEmailChange(c.Request.Context(), req.Token)
	if err != nil {
		switch {
		case strings.Contains(err.Error(), "not found or expired"):
			response.BadRequest(c, "invalid or expired confirmation token")
		case strings.Contains(err.Error(), "already in use"):
			response.BadRequest(c, "email address already in use")
		default:
			slog.Error("failed to confirm email change",
				"error", err.Error(),
			)
			response.InternalError(c, err)
		}
		return
	}

	slog.Info("email address changed", "user_id", userID.String())
	response.OK(c, gin.H{"message": "email address updated"})
}
