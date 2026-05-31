package passwordreset

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/trackmycareer/app/pkg/response"
)

// PasskeyChallengeFunc generates WebAuthn assertion options for a user.
// It is typically bound to the passkey service via a closure in main.go.
type PasskeyChallengeFunc func(ctx context.Context, userID uuid.UUID) (any, error)

// Handler provides HTTP endpoints for password reset flows.
type Handler struct {
	svc                  *Service
	passkeyChallengeFunc PasskeyChallengeFunc
}

// NewHandler creates a new password reset Handler.
func NewHandler(svc *Service, passkeyChallengeFunc PasskeyChallengeFunc) *Handler {
	return &Handler{
		svc:                  svc,
		passkeyChallengeFunc: passkeyChallengeFunc,
	}
}

type forgotPasswordRequest struct {
	Email string `json:"email" binding:"required"`
}

// ForgotPassword handles POST /auth/forgot-password.
// Always returns 200 regardless of whether the email exists, to prevent enumeration.
func (h *Handler) ForgotPassword(c *gin.Context) {
	var req forgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "email is required")
		return
	}

	req.Email = strings.TrimSpace(req.Email)
	if req.Email == "" {
		response.BadRequest(c, "email is required")
		return
	}

	// Fire-and-forget: send the reset email asynchronously so the response
	// time does not reveal whether the account exists.
	go func() {
		if err := h.svc.RequestReset(c.Copy().Request.Context(), req.Email); err != nil {
			slog.Error("failed to process password reset request",
				"error", err.Error(),
				"email", req.Email,
			)
		}
	}()

	response.OK(c, gin.H{"message": "If an account with that email exists, a password reset link has been sent"})
}

type resetPasswordRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
	MFAMethod   string `json:"mfa_method,omitempty"`
	MFACode     string `json:"mfa_code,omitempty"`
}

// ResetPassword handles POST /auth/reset-password.
// Validates the reset token and updates the user's password.
func (h *Handler) ResetPassword(c *gin.Context) {
	var req resetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "token and new_password are required")
		return
	}

	req.Token = strings.TrimSpace(req.Token)
	if req.Token == "" {
		response.BadRequest(c, "token is required")
		return
	}

	// If MFA credentials are provided, use the MFA-aware reset path.
	if req.MFAMethod != "" && req.MFACode != "" {
		if err := h.svc.ResetPasswordWithMFA(c.Request.Context(), req.Token, req.NewPassword, req.MFAMethod, req.MFACode); err != nil {
			if strings.Contains(err.Error(), "token not found") || strings.Contains(err.Error(), "expired") {
				response.BadRequest(c, "Invalid or expired reset token")
				return
			}
			if strings.Contains(err.Error(), "invalid verification code") {
				response.Unauthorised(c, "invalid verification code")
				return
			}
			if strings.Contains(err.Error(), "password must be at least") {
				response.BadRequest(c, err.Error())
				return
			}
			slog.Error("failed to reset password with MFA",
				"error", err.Error(),
			)
			response.InternalError(c, err)
			return
		}

		response.OK(c, gin.H{"message": "Password reset successfully"})
		return
	}

	mfaRequired, methods, err := h.svc.ResetPassword(c.Request.Context(), req.Token, req.NewPassword)
	if err != nil {
		if strings.Contains(err.Error(), "token not found") || strings.Contains(err.Error(), "expired") {
			response.BadRequest(c, "Invalid or expired reset token")
			return
		}
		if strings.Contains(err.Error(), "password must be at least") {
			response.BadRequest(c, err.Error())
			return
		}
		slog.Error("failed to reset password",
			"error", err.Error(),
		)
		response.InternalError(c, err)
		return
	}

	if mfaRequired {
		c.JSON(http.StatusOK, gin.H{
			"data": gin.H{
				"mfa_required": true,
				"methods":      methods,
			},
		})
		return
	}

	response.OK(c, gin.H{"message": "Password reset successfully"})
}

type passkeyChallengeRequest struct {
	Token string `json:"token" binding:"required"`
}

// PasskeyChallenge handles POST /auth/reset-password/passkey-challenge.
// Validates the reset token and returns WebAuthn assertion options for
// passkey-based MFA verification during password reset.
func (h *Handler) PasskeyChallenge(c *gin.Context) {
	if h.passkeyChallengeFunc == nil {
		response.Error(c, http.StatusNotImplemented, "MFA not available")
		return
	}

	var req passkeyChallengeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "token is required")
		return
	}

	req.Token = strings.TrimSpace(req.Token)
	if req.Token == "" {
		response.BadRequest(c, "token is required")
		return
	}

	userID, err := h.svc.ValidateTokenAndGetUserID(c.Request.Context(), req.Token)
	if err != nil {
		response.BadRequest(c, "Invalid or expired reset token")
		return
	}

	assertion, err := h.passkeyChallengeFunc(c.Request.Context(), userID)
	if err != nil {
		slog.Error("failed to generate passkey challenge for password reset",
			"error", err.Error(),
			"user_id", userID.String(),
		)
		response.InternalError(c, err)
		return
	}

	response.OK(c, assertion)
}
