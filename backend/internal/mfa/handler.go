// Package mfa provides HTTP handlers for multi-factor authentication management.
//
// # Wiring in cmd/server/main.go
//
// To wire these routes, add the following to main.go after the existing service/handler setup:
//
//	// MFA sub-service repositories
//	totpRepo := totp.NewRepository(pool)
//	passkeyRepo := passkey.NewRepository(pool)
//	backupRepo := backup.NewRepository(pool)
//
//	// MFA sub-services
//	totpService := totp.NewService(totpRepo, mfaEncryptor, "trackmy.career")
//	passkeyService, err := passkey.NewService(passkeyRepo, rpID, "trackmy.career", cfg.FrontendURL)
//	if err != nil {
//		log.Fatalf("initialising passkey service: %v", err)
//	}
//	backupService := backup.NewService(backupRepo, mfaKey)
//
//	// MFA orchestration service and handler
//	mfaService := mfa.NewService(totpService, passkeyService, backupService, userRepo)
//	mfaHandler := mfa.NewHandler(mfaService, userRepo)
//
//	// MFA routes (under verified group)
//	mfaGroup := verified.Group("/user/me/mfa")
//	{
//		mfaGroup.GET("/status", mfaHandler.GetMFAStatus)
//		mfaGroup.POST("/totp/setup", mfaHandler.SetupTOTP)
//		mfaGroup.POST("/totp/verify", mfaHandler.VerifyTOTP)
//		mfaGroup.DELETE("/totp", mfaHandler.DeleteTOTP)
//		mfaGroup.POST("/passkeys/register/begin", mfaHandler.BeginPasskeyRegistration)
//		mfaGroup.POST("/passkeys/register/complete", mfaHandler.CompletePasskeyRegistration)
//		mfaGroup.GET("/passkeys", mfaHandler.ListPasskeys)
//		mfaGroup.PUT("/passkeys/:id", mfaHandler.RenamePasskey)
//		mfaGroup.DELETE("/passkeys/:id", mfaHandler.DeletePasskey)
//		mfaGroup.GET("/backup-codes/count", mfaHandler.GetBackupCodeCount)
//		mfaGroup.POST("/backup-codes/regenerate", mfaHandler.RegenerateBackupCodes)
//		mfaGroup.POST("/disable", mfaHandler.DisableMFA)
//	}
//
// Imports needed in main.go:
//
//	"github.com/trackmycareer/app/internal/mfa"
//	"github.com/trackmycareer/app/internal/mfa/totp"
//	"github.com/trackmycareer/app/internal/mfa/passkey"
//	"github.com/trackmycareer/app/internal/mfa/backup"
package mfa

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/trackmycareer/app/internal/password"
	"github.com/trackmycareer/app/internal/user"
	"github.com/trackmycareer/app/pkg/response"
)

// Handler provides HTTP endpoints for MFA management.
type Handler struct {
	svc      *Service
	userRepo *user.Repository
}

// NewHandler creates a new MFA Handler.
func NewHandler(svc *Service, userRepo *user.Repository) *Handler {
	return &Handler{
		svc:      svc,
		userRepo: userRepo,
	}
}

// verifyPassword checks the user's password for destructive MFA operations.
// Returns the user on success, or writes an error response and returns nil on failure.
func (h *Handler) verifyPassword(c *gin.Context, userID uuid.UUID, pw string) *user.User {
	u, err := h.userRepo.GetByID(c.Request.Context(), userID)
	if err != nil {
		response.NotFound(c, "user not found")
		return nil
	}

	if u.PasswordHash == "" {
		response.BadRequest(c, "account uses OAuth authentication; password confirmation is not available")
		return nil
	}

	match, err := password.Verify(pw, u.PasswordHash)
	if err != nil {
		response.InternalError(c, err)
		return nil
	}
	if !match {
		response.Unauthorised(c, "incorrect password")
		return nil
	}

	return &u
}

// getUserForMFA fetches the user and returns it. On error it writes a response and returns nil.
func (h *Handler) getUserForMFA(c *gin.Context, userID uuid.UUID) *user.User {
	u, err := h.userRepo.GetByID(c.Request.Context(), userID)
	if err != nil {
		response.NotFound(c, "user not found")
		return nil
	}
	return &u
}

// isFirstMFAMethod checks whether the user currently has no MFA methods configured.
func (h *Handler) isFirstMFAMethod(ctx context.Context, userID uuid.UUID) (bool, error) {
	has, err := h.svc.HasAnyMethod(ctx, userID)
	if err != nil {
		return false, err
	}
	return !has, nil
}

// GetMFAStatus handles GET /status.
func (h *Handler) GetMFAStatus(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	status, err := h.svc.GetStatus(c.Request.Context(), userID)
	if err != nil {
		slog.Error("failed to get MFA status", "error", err.Error(), "user_id", userID.String())
		response.InternalError(c, err)
		return
	}

	response.OK(c, status)
}

type setupTOTPResponse struct {
	URI    string `json:"uri"`
	Secret string `json:"secret"`
}

// SetupTOTP handles POST /totp/setup.
func (h *Handler) SetupTOTP(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	u := h.getUserForMFA(c, userID)
	if u == nil {
		return
	}

	uri, secret, err := h.svc.TOTP().Setup(c.Request.Context(), userID, u.Email)
	if err != nil {
		if strings.Contains(err.Error(), "already configured") {
			response.Error(c, http.StatusConflict, "TOTP is already configured")
			return
		}
		slog.Error("failed to set up TOTP", "error", err.Error(), "user_id", userID.String())
		response.InternalError(c, err)
		return
	}

	response.OK(c, setupTOTPResponse{URI: uri, Secret: secret})
}

type verifyTOTPRequest struct {
	Code string `json:"code" binding:"required"`
}

type verifyTOTPResponse struct {
	BackupCodes []string `json:"backup_codes,omitempty"`
}

// VerifyTOTP handles POST /totp/verify.
func (h *Handler) VerifyTOTP(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	var req verifyTOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "code is required")
		return
	}

	// Check if this is the first MFA method before verification.
	firstMethod, err := h.isFirstMFAMethod(c.Request.Context(), userID)
	if err != nil {
		slog.Error("failed to check MFA methods", "error", err.Error(), "user_id", userID.String())
		response.InternalError(c, err)
		return
	}

	if err := h.svc.TOTP().Verify(c.Request.Context(), userID, req.Code); err != nil {
		if strings.Contains(err.Error(), "invalid verification code") {
			response.Unauthorised(c, "invalid verification code")
			return
		}
		slog.Error("failed to verify TOTP", "error", err.Error(), "user_id", userID.String())
		response.InternalError(c, err)
		return
	}

	// Update the MFA flag on the user.
	if err := h.svc.UpdateMFAFlag(c.Request.Context(), userID); err != nil {
		slog.Error("failed to update MFA flag", "error", err.Error(), "user_id", userID.String())
		response.InternalError(c, err)
		return
	}

	resp := verifyTOTPResponse{}

	// Generate backup codes if this is the first MFA method.
	if firstMethod {
		codes, err := h.svc.Backup().Generate(c.Request.Context(), userID)
		if err != nil {
			slog.Error("failed to generate backup codes", "error", err.Error(), "user_id", userID.String())
			response.InternalError(c, err)
			return
		}
		resp.BackupCodes = codes
	}

	response.OK(c, resp)
}

type deleteTOTPRequest struct {
	Password string `json:"password" binding:"required"`
}

// DeleteTOTP handles DELETE /totp.
func (h *Handler) DeleteTOTP(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	var req deleteTOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "password is required")
		return
	}

	if h.verifyPassword(c, userID, req.Password) == nil {
		return
	}

	if err := h.svc.TOTP().Delete(c.Request.Context(), userID); err != nil {
		slog.Error("failed to delete TOTP", "error", err.Error(), "user_id", userID.String())
		response.InternalError(c, err)
		return
	}

	if err := h.svc.UpdateMFAFlag(c.Request.Context(), userID); err != nil {
		slog.Error("failed to update MFA flag", "error", err.Error(), "user_id", userID.String())
		response.InternalError(c, err)
		return
	}

	response.OK(c, gin.H{"message": "TOTP removed"})
}

// BeginPasskeyRegistration handles POST /passkeys/register/begin.
func (h *Handler) BeginPasskeyRegistration(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	u := h.getUserForMFA(c, userID)
	if u == nil {
		return
	}

	existing, err := h.svc.Passkey().ListByUserID(c.Request.Context(), userID)
	if err != nil {
		slog.Error("failed to list passkeys", "error", err.Error(), "user_id", userID.String())
		response.InternalError(c, err)
		return
	}

	options, err := h.svc.Passkey().BeginRegistration(c.Request.Context(), userID, u.Email, u.Name, existing)
	if err != nil {
		slog.Error("failed to begin passkey registration", "error", err.Error(), "user_id", userID.String())
		response.InternalError(c, err)
		return
	}

	response.OK(c, options)
}

type passkeyResponseItem struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	CreatedAt string    `json:"created_at"`
}

type completePasskeyRegistrationResponse struct {
	Passkey     passkeyResponseItem `json:"passkey"`
	BackupCodes []string            `json:"backup_codes,omitempty"`
}

// CompletePasskeyRegistration handles POST /passkeys/register/complete.
func (h *Handler) CompletePasskeyRegistration(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	var req struct {
		Name       string          `json:"name"`
		Credential json.RawMessage `json:"credential"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "name and credential are required")
		return
	}

	passkeyName := strings.TrimSpace(req.Name)
	if passkeyName == "" {
		passkeyName = "My Passkey"
	}
	if len(passkeyName) > 255 {
		response.BadRequest(c, "name must be at most 255 characters")
		return
	}

	u := h.getUserForMFA(c, userID)
	if u == nil {
		return
	}

	// Check if this is the first MFA method before registration.
	firstMethod, err := h.isFirstMFAMethod(c.Request.Context(), userID)
	if err != nil {
		slog.Error("failed to check MFA methods", "error", err.Error(), "user_id", userID.String())
		response.InternalError(c, err)
		return
	}

	existing, err := h.svc.Passkey().ListByUserID(c.Request.Context(), userID)
	if err != nil {
		slog.Error("failed to list passkeys", "error", err.Error(), "user_id", userID.String())
		response.InternalError(c, err)
		return
	}

	pk, err := h.svc.Passkey().CompleteRegistration(c.Request.Context(), userID, u.Email, u.Name, passkeyName, existing, req.Credential)
	if err != nil {
		slog.Error("failed to complete passkey registration", "error", err.Error(), "user_id", userID.String())
		response.InternalError(c, err)
		return
	}

	// Update the MFA flag on the user.
	if err := h.svc.UpdateMFAFlag(c.Request.Context(), userID); err != nil {
		slog.Error("failed to update MFA flag", "error", err.Error(), "user_id", userID.String())
		response.InternalError(c, err)
		return
	}

	resp := completePasskeyRegistrationResponse{
		Passkey: passkeyResponseItem{
			ID:        pk.ID,
			Name:      pk.Name,
			CreatedAt: pk.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		},
	}

	// Generate backup codes if this is the first MFA method.
	if firstMethod {
		codes, err := h.svc.Backup().Generate(c.Request.Context(), userID)
		if err != nil {
			slog.Error("failed to generate backup codes", "error", err.Error(), "user_id", userID.String())
			response.InternalError(c, err)
			return
		}
		resp.BackupCodes = codes
	}

	response.OK(c, resp)
}

type listPasskeysResponseItem struct {
	ID         uuid.UUID `json:"id"`
	Name       string    `json:"name"`
	CreatedAt  string    `json:"created_at"`
	LastUsedAt *string   `json:"last_used_at,omitempty"`
}

// ListPasskeys handles GET /passkeys.
func (h *Handler) ListPasskeys(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	passkeys, err := h.svc.Passkey().ListByUserID(c.Request.Context(), userID)
	if err != nil {
		slog.Error("failed to list passkeys", "error", err.Error(), "user_id", userID.String())
		response.InternalError(c, err)
		return
	}

	items := make([]listPasskeysResponseItem, len(passkeys))
	for i, p := range passkeys {
		item := listPasskeysResponseItem{
			ID:        p.ID,
			Name:      p.Name,
			CreatedAt: p.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
		if p.LastUsedAt != nil {
			formatted := p.LastUsedAt.Format("2006-01-02T15:04:05Z07:00")
			item.LastUsedAt = &formatted
		}
		items[i] = item
	}

	response.OK(c, items)
}

type renamePasskeyRequest struct {
	Name string `json:"name" binding:"required,max=255"`
}

// RenamePasskey handles PUT /passkeys/:id.
func (h *Handler) RenamePasskey(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	passkeyID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid passkey ID")
		return
	}

	var req renamePasskeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "name is required")
		return
	}

	if err := h.svc.Passkey().Rename(c.Request.Context(), passkeyID, userID, req.Name); err != nil {
		if strings.Contains(err.Error(), "not found") {
			response.NotFound(c, "passkey not found")
			return
		}
		slog.Error("failed to rename passkey", "error", err.Error(), "user_id", userID.String())
		response.InternalError(c, err)
		return
	}

	response.OK(c, gin.H{"message": "passkey renamed"})
}

type deletePasskeyRequest struct {
	Password string `json:"password" binding:"required"`
}

// DeletePasskey handles DELETE /passkeys/:id.
func (h *Handler) DeletePasskey(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	passkeyID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "invalid passkey ID")
		return
	}

	var req deletePasskeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "password is required")
		return
	}

	u := h.verifyPassword(c, userID, req.Password)
	if u == nil {
		return
	}

	// If admin, check this is not the last MFA method.
	if u.IsAdmin {
		totpConfigured, err := h.svc.TOTP().IsConfigured(c.Request.Context(), userID)
		if err != nil {
			slog.Error("failed to check TOTP status", "error", err.Error(), "user_id", userID.String())
			response.InternalError(c, err)
			return
		}

		passkeyCount, err := h.svc.Passkey().Count(c.Request.Context(), userID)
		if err != nil {
			slog.Error("failed to count passkeys", "error", err.Error(), "user_id", userID.String())
			response.InternalError(c, err)
			return
		}

		// If TOTP is not configured and this is the only passkey, block the delete.
		if !totpConfigured && passkeyCount <= 1 {
			response.Forbidden(c, "cannot remove last MFA method for administrator accounts")
			return
		}
	}

	if err := h.svc.Passkey().Delete(c.Request.Context(), passkeyID, userID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			response.NotFound(c, "passkey not found")
			return
		}
		slog.Error("failed to delete passkey", "error", err.Error(), "user_id", userID.String())
		response.InternalError(c, err)
		return
	}

	if err := h.svc.UpdateMFAFlag(c.Request.Context(), userID); err != nil {
		slog.Error("failed to update MFA flag", "error", err.Error(), "user_id", userID.String())
		response.InternalError(c, err)
		return
	}

	response.OK(c, gin.H{"message": "passkey removed"})
}

type backupCodeCountResponse struct {
	Remaining int `json:"remaining"`
}

// GetBackupCodeCount handles GET /backup-codes/count.
func (h *Handler) GetBackupCodeCount(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	count, err := h.svc.Backup().CountRemaining(c.Request.Context(), userID)
	if err != nil {
		slog.Error("failed to count backup codes", "error", err.Error(), "user_id", userID.String())
		response.InternalError(c, err)
		return
	}

	response.OK(c, backupCodeCountResponse{Remaining: count})
}

type regenerateBackupCodesRequest struct {
	Password string `json:"password" binding:"required"`
}

type regenerateBackupCodesResponse struct {
	BackupCodes []string `json:"backup_codes"`
}

// RegenerateBackupCodes handles POST /backup-codes/regenerate.
func (h *Handler) RegenerateBackupCodes(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	var req regenerateBackupCodesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "password is required")
		return
	}

	if h.verifyPassword(c, userID, req.Password) == nil {
		return
	}

	codes, err := h.svc.Backup().Regenerate(c.Request.Context(), userID)
	if err != nil {
		slog.Error("failed to regenerate backup codes", "error", err.Error(), "user_id", userID.String())
		response.InternalError(c, err)
		return
	}

	response.OK(c, regenerateBackupCodesResponse{BackupCodes: codes})
}

type disableMFARequest struct {
	Password string `json:"password" binding:"required"`
}

// DisableMFA handles POST /disable.
func (h *Handler) DisableMFA(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	var req disableMFARequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "password is required")
		return
	}

	u := h.verifyPassword(c, userID, req.Password)
	if u == nil {
		return
	}

	if u.IsAdmin {
		response.Forbidden(c, "administrators must have MFA enabled")
		return
	}

	// Delete all passkeys.
	passkeys, err := h.svc.Passkey().ListByUserID(c.Request.Context(), userID)
	if err != nil {
		slog.Error("failed to list passkeys for disable", "error", err.Error(), "user_id", userID.String())
		response.InternalError(c, err)
		return
	}
	for _, p := range passkeys {
		if delErr := h.svc.Passkey().Delete(c.Request.Context(), p.ID, userID); delErr != nil {
			slog.Error("failed to delete passkey during MFA disable",
				"error", delErr.Error(),
				"passkey_id", p.ID.String(),
			)
		}
	}

	// Use orchestration service for TOTP deletion and flag update.
	// The isAdmin=false check is already done above.
	if err := h.svc.DisableAll(c.Request.Context(), userID, false); err != nil {
		slog.Error("failed to disable MFA", "error", err.Error(), "user_id", userID.String())
		response.InternalError(c, err)
		return
	}

	response.OK(c, gin.H{"message": "MFA disabled"})
}
