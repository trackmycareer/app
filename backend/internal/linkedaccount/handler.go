package linkedaccount

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/oauth2"

	"github.com/trackmycareer/app/pkg/response"
)

type oauthLinkProvider struct {
	config      *oauth2.Config
	userInfoURL string
	headers     map[string]string
	parseUser   func(body []byte) (oauthLinkUser, error)
}

type oauthLinkUser struct {
	ID         string
	ProfileURL string
}

type Handler struct {
	service     *Service
	providers   map[string]*oauthLinkProvider
	frontendURL string
	jwtSecret   []byte
}

func NewHandler(
	service *Service,
	frontendURL string,
	linkedinClientID, linkedinClientSecret string,
	githubClientID, githubClientSecret string,
	jwtSecret string,
) *Handler {
	h := &Handler{
		service:     service,
		providers:   make(map[string]*oauthLinkProvider),
		frontendURL: frontendURL,
		jwtSecret:   []byte(jwtSecret),
	}

	if linkedinClientID != "" && linkedinClientSecret != "" {
		h.providers[ProviderLinkedIn] = &oauthLinkProvider{
			config: &oauth2.Config{
				ClientID:     linkedinClientID,
				ClientSecret: linkedinClientSecret,
				RedirectURL:  frontendURL + "/auth/link/linkedin/callback",
				Scopes:       []string{"openid", "profile", "r_profile_basicinfo"},
				Endpoint: oauth2.Endpoint{
					AuthURL:  "https://www.linkedin.com/oauth/v2/authorization",
					TokenURL: "https://www.linkedin.com/oauth/v2/accessToken",
				},
			},
			userInfoURL: "https://api.linkedin.com/rest/identityMe",
			headers: map[string]string{
				"LinkedIn-Version": "202604",
			},
			parseUser: func(body []byte) (oauthLinkUser, error) {
				var data struct {
					ID        string `json:"id"`
					BasicInfo struct {
						ProfileURL string `json:"profileUrl"`
					} `json:"basicInfo"`
				}
				if err := json.Unmarshal(body, &data); err != nil {
					return oauthLinkUser{}, err
				}
				profileURL := data.BasicInfo.ProfileURL
				if profileURL != "" {
					resolved, err := resolveRedirect(profileURL)
					if err == nil && resolved != "" {
						profileURL = resolved
					}
				}
				if profileURL == "" {
					profileURL = "https://www.linkedin.com/in/" + data.ID
				}
				return oauthLinkUser{ID: data.ID, ProfileURL: profileURL}, nil
			},
		}
	}

	if githubClientID != "" && githubClientSecret != "" {
		h.providers[ProviderGitHub] = &oauthLinkProvider{
			config: &oauth2.Config{
				ClientID:     githubClientID,
				ClientSecret: githubClientSecret,
				RedirectURL:  frontendURL + "/auth/link/github/callback",
				Scopes:       []string{"read:user"},
				Endpoint: oauth2.Endpoint{
					AuthURL:  "https://github.com/login/oauth/authorize",
					TokenURL: "https://github.com/login/oauth/access_token",
				},
			},
			userInfoURL: "https://api.github.com/user",
			parseUser: func(body []byte) (oauthLinkUser, error) {
				var data struct {
					ID    int    `json:"id"`
					Login string `json:"login"`
				}
				if err := json.Unmarshal(body, &data); err != nil {
					return oauthLinkUser{}, err
				}
				return oauthLinkUser{
					ID:         fmt.Sprintf("%d", data.ID),
					ProfileURL: "https://github.com/" + data.Login,
				}, nil
			},
		}
	}

	return h
}

func (h *Handler) List(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	accounts, err := h.service.ListByUser(c.Request.Context(), userID)
	if err != nil {
		response.InternalError(c, err)
		return
	}
	response.OK(c, accounts)
}

func (h *Handler) signState(userID string) string {
	mac := hmac.New(sha256.New, h.jwtSecret)
	mac.Write([]byte(userID))
	sig := hex.EncodeToString(mac.Sum(nil))
	return userID + ":" + sig
}

func (h *Handler) verifyState(state string) (uuid.UUID, error) {
	parts := strings.SplitN(state, ":", 2)
	if len(parts) != 2 {
		return uuid.Nil, fmt.Errorf("invalid state format")
	}
	userIDStr, sig := parts[0], parts[1]

	mac := hmac.New(sha256.New, h.jwtSecret)
	mac.Write([]byte(userIDStr))
	expected := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(sig), []byte(expected)) {
		return uuid.Nil, fmt.Errorf("invalid state signature")
	}

	return uuid.Parse(userIDStr)
}

func (h *Handler) InitiateLink(c *gin.Context) {
	provider := c.Param("provider")
	p, ok := h.providers[provider]
	if !ok {
		response.BadRequest(c, "unsupported provider: "+provider)
		return
	}

	userID := c.MustGet("user_id").(uuid.UUID)
	state := h.signState(userID.String())

	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("oauth_link_state", state, 600, "/api/v1/linked-accounts", "", true, true)
	response.OK(c, gin.H{"auth_url": p.config.AuthCodeURL(state)})
}

func (h *Handler) HandleCallback(c *gin.Context) {
	provider := c.Param("provider")
	p, ok := h.providers[provider]
	if !ok {
		response.BadRequest(c, "unsupported provider: "+provider)
		return
	}

	var req struct {
		Code  string `json:"code" binding:"required"`
		State string `json:"state" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "code and state are required")
		return
	}

	storedState, err := c.Cookie("oauth_link_state")
	if err != nil || storedState != req.State {
		response.Unauthorised(c, "invalid OAuth state")
		return
	}
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("oauth_link_state", "", -1, "/api/v1/linked-accounts", "", true, true)

	userID, err := h.verifyState(storedState)
	if err != nil {
		slog.Warn("invalid OAuth link state", "error", err)
		response.Unauthorised(c, "invalid OAuth state")
		return
	}

	ctx := c.Request.Context()
	oauthUser, err := exchangeAndFetchUser(ctx, p, req.Code)
	if err != nil {
		slog.Error("OAuth link callback failed", "provider", provider, "error", err)
		response.Error(c, http.StatusBadGateway, "failed to verify account with provider")
		return
	}

	la, err := h.service.LinkOAuth(ctx, userID, provider, oauthUser.ID, oauthUser.ProfileURL)
	if err != nil {
		slog.Error("failed to link account", "provider", provider, "error", err)
		response.Error(c, http.StatusConflict, "account linking failed")
		return
	}

	response.OK(c, la)
}

func exchangeAndFetchUser(ctx context.Context, p *oauthLinkProvider, code string) (oauthLinkUser, error) {
	token, err := p.config.Exchange(ctx, code)
	if err != nil {
		return oauthLinkUser{}, fmt.Errorf("exchanging code: %w", err)
	}

	client := p.config.Client(ctx, token)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.userInfoURL, nil)
	if err != nil {
		return oauthLinkUser{}, fmt.Errorf("creating request: %w", err)
	}
	for k, v := range p.headers {
		req.Header.Set(k, v)
	}
	resp, err := client.Do(req)
	if err != nil {
		return oauthLinkUser{}, fmt.Errorf("fetching user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return oauthLinkUser{}, fmt.Errorf("user info returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return oauthLinkUser{}, fmt.Errorf("reading user info body: %w", err)
	}

	return p.parseUser(body)
}

func resolveRedirect(rawURL string) (string, error) {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	resp, err := client.Get(rawURL)
	if err != nil {
		return "", err
	}
	resp.Body.Close()
	if resp.StatusCode == http.StatusMovedPermanently || resp.StatusCode == http.StatusFound || resp.StatusCode == http.StatusTemporaryRedirect {
		return resp.Header.Get("Location"), nil
	}
	return "", fmt.Errorf("no redirect, status %d", resp.StatusCode)
}

type addWebsiteRequest struct {
	URL string `json:"url" binding:"required"`
}

func (h *Handler) AddWebsite(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	var req addWebsiteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "url is required")
		return
	}

	la, err := h.service.AddWebsite(c.Request.Context(), userID, req.URL)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Created(c, la)
}

func (h *Handler) VerifyWebsite(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)

	la, err := h.service.VerifyWebsite(c.Request.Context(), userID)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.OK(c, la)
}

func (h *Handler) Unlink(c *gin.Context) {
	userID := c.MustGet("user_id").(uuid.UUID)
	provider := c.Param("provider")

	if err := h.service.Unlink(c.Request.Context(), userID, provider); err != nil {
		response.NotFound(c, "linked account not found")
		return
	}

	response.OK(c, gin.H{"message": "account unlinked"})
}
