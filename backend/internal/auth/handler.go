package auth

import (
	"context"
	"crypto/subtle"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

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
}

func NewHandler(service *Service, oauthManager *OAuthManager, settingsRepo *settings.Repository, frontendURL string, vs VerificationSender) *Handler {
	return &Handler{
		service:            service,
		oauthManager:       oauthManager,
		settingsRepo:       settingsRepo,
		frontendURL:        frontendURL,
		verificationSender: vs,
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

	tokens, err := h.service.Login(c.Request.Context(), req)
	if err != nil {
		response.Unauthorised(c, "invalid email or password")
		return
	}

	h.setRefreshCookie(c, tokens.RefreshToken)
	response.OK(c, gin.H{"access_token": tokens.AccessToken})
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

	tokens, err := h.oauthManager.HandleCallback(c.Request.Context(), provider, req.Code)
	if err != nil {
		response.Error(c, http.StatusUnauthorized, "authentication failed")
		return
	}

	h.setRefreshCookie(c, tokens.RefreshToken)
	response.OK(c, gin.H{"access_token": tokens.AccessToken})
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
