package auth

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/trackmycareer/app/internal/mfa"
	"github.com/trackmycareer/app/internal/settings"
	"github.com/trackmycareer/app/pkg/response"
)

// VerificationSender sends a verification email to a user.
// Implemented by the verification service; defined here to avoid circular imports.
type VerificationSender interface {
	SendVerification(ctx context.Context, userID uuid.UUID) error
}

type Handler struct {
	service            *Service
	oauthManager       *OAuthManager
	settingsRepo       *settings.Repository
	frontendURL        string
	verificationSender VerificationSender
	mfaSvc             *mfa.Service
}

func NewHandler(service *Service, oauthManager *OAuthManager, settingsRepo *settings.Repository, frontendURL string, vs VerificationSender, mfaSvc *mfa.Service) *Handler {
	return &Handler{
		service:            service,
		oauthManager:       oauthManager,
		settingsRepo:       settingsRepo,
		frontendURL:        frontendURL,
		verificationSender: vs,
		mfaSvc:             mfaSvc,
	}
}

func (h *Handler) Register(c *gin.Context) {
	enabled, err := h.settingsRepo.IsRegistrationEnabled(c.Request.Context())
	if err != nil {
		response.InternalError(c, err)
		return
	}
	if !enabled {
		response.Forbidden(c, "registration is currently disabled")
		return
	}

	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, response.FormatBindingError(err))
		return
	}

	result, err := h.service.Register(c.Request.Context(), req)
	if err != nil {
		response.BadRequest(c, "could not create account")
		return
	}

	if h.verificationSender != nil {
		go func() {
			if sendErr := h.verificationSender.SendVerification(context.Background(), result.UserID); sendErr != nil {
				slog.Error("sending verification email", "error", sendErr.Error(), "user_id", result.UserID)
			}
		}()
	}

	h.setRefreshCookie(c, result.Tokens.RefreshToken)
	response.Created(c, gin.H{"access_token": result.Tokens.AccessToken})
}

func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, response.FormatBindingError(err))
		return
	}

	result, err := h.service.Login(c.Request.Context(), req)
	if err != nil {
		response.Unauthorised(c, "invalid email or password")
		return
	}

	if result.MFASession != "" {
		response.OK(c, gin.H{
			"mfa_required": true,
			"mfa_session":  result.MFASession,
			"methods":      result.Methods,
		})
		return
	}

	h.setRefreshCookie(c, result.TokenPair.RefreshToken)
	response.OK(c, gin.H{"access_token": result.TokenPair.AccessToken})
}

func (h *Handler) Refresh(c *gin.Context) {
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		response.Unauthorised(c, "no refresh token")
		return
	}

	tokens, err := h.service.Refresh(c.Request.Context(), refreshToken)
	if err != nil {
		response.Unauthorised(c, "invalid refresh token")
		return
	}

	h.setRefreshCookie(c, tokens.RefreshToken)
	response.OK(c, gin.H{"access_token": tokens.AccessToken})
}

func (h *Handler) OAuthInitiate(c *gin.Context) {
	provider := c.Param("provider")
	if !h.oauthManager.HasProvider(provider) {
		response.BadRequest(c, "unsupported OAuth provider")
		return
	}

	url, state, err := h.oauthManager.GetAuthURL(provider)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("oauth_state", state, 600, "/api/v1/auth", "", true, true)
	response.OK(c, gin.H{"auth_url": url})
}

func (h *Handler) OAuthCallback(c *gin.Context) {
	provider := c.Param("provider")

	var req struct {
		Code  string `json:"code" binding:"required"`
		State string `json:"state" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "code and state are required")
		return
	}

	storedState, err := c.Cookie("oauth_state")
	if err != nil || storedState == "" || subtle.ConstantTimeCompare([]byte(storedState), []byte(req.State)) != 1 {
		response.Unauthorised(c, "invalid OAuth state")
		return
	}

	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("oauth_state", "", -1, "/api/v1/auth", "", true, true)

	result, err := h.oauthManager.HandleCallback(c.Request.Context(), provider, req.Code)
	if err != nil {
		slog.Error("OAuth callback failed", "provider", provider, "error", err.Error())
		response.Error(c, http.StatusUnauthorized, "authentication failed")
		return
	}

	if result.MFASession != "" {
		response.OK(c, gin.H{
			"mfa_required": true,
			"mfa_session":  result.MFASession,
			"methods":      result.Methods,
		})
		return
	}

	h.setRefreshCookie(c, result.TokenPair.RefreshToken)
	response.OK(c, gin.H{"access_token": result.TokenPair.AccessToken})
}

type verifyMFARequest struct {
	MFASession string          `json:"mfa_session" binding:"required"`
	Method     string          `json:"method" binding:"required"`
	Code       string          `json:"code"`
	Assertion  json.RawMessage `json:"assertion"`
}

// VerifyMFA handles POST /auth/mfa/verify.
// Validates the MFA session token and verifies the provided code, then issues
// a full token pair on success.
func (h *Handler) VerifyMFA(c *gin.Context) {
	if h.mfaSvc == nil {
		response.Error(c, http.StatusNotImplemented, "MFA not available")
		return
	}

	var req verifyMFARequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, response.FormatBindingError(err))
		return
	}

	tokens, userID, err := h.service.CompleteMFALogin(c.Request.Context(), req.MFASession)
	if err != nil {
		response.Unauthorised(c, "MFA session expired, please log in again")
		return
	}

	method := mfa.Method(req.Method)
	switch method {
	case mfa.MethodTOTP, mfa.MethodBackup:
		if req.Code == "" {
			response.BadRequest(c, "code is required for this MFA method")
			return
		}
		if err := h.mfaSvc.VerifyAnyMethod(c.Request.Context(), userID, method, req.Code); err != nil {
			response.Unauthorised(c, "invalid verification code")
			return
		}
	case mfa.MethodPasskey:
		if len(req.Assertion) == 0 {
			response.BadRequest(c, "assertion is required for passkey verification")
			return
		}
		passkeys, listErr := h.mfaSvc.Passkey().ListByUserID(c.Request.Context(), userID)
		if listErr != nil {
			slog.Error("failed to list passkeys for MFA verify", "error", listErr.Error())
			response.InternalError(c, listErr)
			return
		}
		if err := h.mfaSvc.Passkey().CompleteAuthentication(c.Request.Context(), userID, passkeys, req.Assertion); err != nil {
			response.Unauthorised(c, "passkey verification failed")
			return
		}
	default:
		response.BadRequest(c, "unsupported MFA method")
		return
	}

	h.setRefreshCookie(c, tokens.RefreshToken)
	response.OK(c, gin.H{"access_token": tokens.AccessToken})
}

type mfaPasskeyChallengeRequest struct {
	MFASession string `json:"mfa_session" binding:"required"`
}

// MFAPasskeyChallenge handles POST /auth/mfa/passkey/challenge.
// Validates the MFA session token and returns WebAuthn assertion options for
// passkey-based MFA verification during login.
func (h *Handler) MFAPasskeyChallenge(c *gin.Context) {
	if h.mfaSvc == nil {
		response.Error(c, http.StatusNotImplemented, "MFA not available")
		return
	}

	var req mfaPasskeyChallengeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, response.FormatBindingError(err))
		return
	}

	claims, err := h.service.jwtManager.ValidateToken(req.MFASession)
	if err != nil || claims.TokenType != "mfa_pending" {
		response.Unauthorised(c, "MFA session expired, please log in again")
		return
	}

	passkeys, err := h.mfaSvc.Passkey().ListByUserID(c.Request.Context(), claims.UserID)
	if err != nil {
		slog.Error("failed to list passkeys for MFA challenge", "error", err.Error(), "user_id", claims.UserID.String())
		response.InternalError(c, err)
		return
	}

	assertion, err := h.mfaSvc.Passkey().BeginAuthentication(c.Request.Context(), claims.UserID, passkeys)
	if err != nil {
		slog.Error("failed to begin passkey authentication", "error", err.Error(), "user_id", claims.UserID.String())
		response.InternalError(c, err)
		return
	}

	response.OK(c, assertion)
}

func (h *Handler) Logout(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	if err := h.service.Logout(c.Request.Context(), userID); err != nil {
		response.InternalError(c, err)
		return
	}
	// Clear the refresh token cookie
	h.setRefreshCookie(c, "")
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie("refresh_token", "", -1, "/api/v1/auth", "", true, true)
	response.OK(c, gin.H{"message": "logged out"})
}

func (h *Handler) Providers(c *gin.Context) {
	enabled, _ := h.settingsRepo.IsRegistrationEnabled(c.Request.Context())

	response.OK(c, gin.H{
		"providers":            h.oauthManager.EnabledProviders(),
		"registration_enabled": enabled,
	})
}

func (h *Handler) setRefreshCookie(c *gin.Context, token string) {
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie("refresh_token", token, int(RefreshTokenDuration.Seconds()), "/api/v1/auth", "", true, true)
}
