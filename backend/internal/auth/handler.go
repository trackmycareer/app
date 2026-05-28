package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/bhcloudlabs/trackmy-career/internal/settings"
	"github.com/bhcloudlabs/trackmy-career/pkg/response"
)

type Handler struct {
	service      *Service
	oauthManager *OAuthManager
	settingsRepo *settings.Repository
	frontendURL  string
}

func NewHandler(service *Service, oauthManager *OAuthManager, settingsRepo *settings.Repository, frontendURL string) *Handler {
	return &Handler{
		service:      service,
		oauthManager: oauthManager,
		settingsRepo: settingsRepo,
		frontendURL:  frontendURL,
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
		response.BadRequest(c, "invalid request: "+err.Error())
		return
	}

	tokens, err := h.service.Register(c.Request.Context(), req)
	if err != nil {
		response.Error(c, http.StatusConflict, "could not create account: "+err.Error())
		return
	}

	h.setRefreshCookie(c, tokens.RefreshToken)
	response.Created(c, gin.H{"access_token": tokens.AccessToken})
}

func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request: "+err.Error())
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
	c.Redirect(http.StatusTemporaryRedirect, url)
}

func (h *Handler) OAuthCallback(c *gin.Context) {
	provider := c.Param("provider")
	code := c.Query("code")
	if code == "" {
		c.Redirect(http.StatusTemporaryRedirect, h.frontendURL+"/login?error=oauth_failed")
		return
	}

	storedState, err := c.Cookie("oauth_state")
	if err != nil || storedState == "" || storedState != c.Query("state") {
		c.Redirect(http.StatusTemporaryRedirect, h.frontendURL+"/login?error=oauth_failed")
		return
	}

	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("oauth_state", "", -1, "/api/v1/auth", "", true, true)

	tokens, err := h.oauthManager.HandleCallback(c.Request.Context(), provider, code)
	if err != nil {
		c.Redirect(http.StatusTemporaryRedirect, h.frontendURL+"/login?error=oauth_failed")
		return
	}

	h.setRefreshCookie(c, tokens.RefreshToken)
	c.Redirect(http.StatusTemporaryRedirect, h.frontendURL+"/auth/callback")
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
