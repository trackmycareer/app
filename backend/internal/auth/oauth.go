package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/google/uuid"
	"golang.org/x/oauth2"
	oauthgithub "golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"

	"github.com/trackmycareer/app/internal/settings"
	"github.com/trackmycareer/app/internal/user"
)

type OAuthProvider struct {
	config      *oauth2.Config
	userInfoURL string
	name        string
	parseUser   func(body []byte) (oauthUser, error)
}

type oauthUser struct {
	ID        string
	Email     string
	Name      string
	AvatarURL string
}

type OAuthManager struct {
	providers    map[string]*OAuthProvider
	userRepo     *user.Repository
	settingsRepo *settings.Repository
	jwtManager   *JWTManager
}

func NewOAuthManager(
	userRepo *user.Repository,
	settingsRepo *settings.Repository,
	jwtManager *JWTManager,
	googleClientID, googleClientSecret string,
	githubClientID, githubClientSecret string,
	frontendURL string,
) *OAuthManager {
	m := &OAuthManager{
		providers:    make(map[string]*OAuthProvider),
		userRepo:     userRepo,
		settingsRepo: settingsRepo,
		jwtManager:   jwtManager,
	}

	if googleClientID != "" && googleClientSecret != "" {
		m.providers["google"] = &OAuthProvider{
			name: "google",
			config: &oauth2.Config{
				ClientID:     googleClientID,
				ClientSecret: googleClientSecret,
				RedirectURL:  frontendURL + "/auth/google/callback",
				Scopes:       []string{"openid", "email", "profile"},
				Endpoint:     google.Endpoint,
			},
			userInfoURL: "https://www.googleapis.com/oauth2/v2/userinfo",
			parseUser: func(body []byte) (oauthUser, error) {
				var data struct {
					ID      string `json:"id"`
					Email   string `json:"email"`
					Name    string `json:"name"`
					Picture string `json:"picture"`
				}
				if err := json.Unmarshal(body, &data); err != nil {
					return oauthUser{}, err
				}
				return oauthUser{
					ID:        data.ID,
					Email:     data.Email,
					Name:      data.Name,
					AvatarURL: data.Picture,
				}, nil
			},
		}
	}

	if githubClientID != "" && githubClientSecret != "" {
		m.providers["github"] = &OAuthProvider{
			name: "github",
			config: &oauth2.Config{
				ClientID:     githubClientID,
				ClientSecret: githubClientSecret,
				RedirectURL:  frontendURL + "/auth/github/callback",
				Scopes:       []string{"user:email", "read:user"},
				Endpoint:     oauthgithub.Endpoint,
			},
			userInfoURL: "https://api.github.com/user",
			parseUser: func(body []byte) (oauthUser, error) {
				var data struct {
					ID        int    `json:"id"`
					Email     string `json:"email"`
					Name      string `json:"name"`
					Login     string `json:"login"`
					AvatarURL string `json:"avatar_url"`
				}
				if err := json.Unmarshal(body, &data); err != nil {
					return oauthUser{}, err
				}
				name := data.Name
				if name == "" {
					name = data.Login
				}
				return oauthUser{
					ID:        fmt.Sprintf("%d", data.ID),
					Email:     data.Email,
					Name:      name,
					AvatarURL: data.AvatarURL,
				}, nil
			},
		}
	}

	return m
}

func (m *OAuthManager) HasProvider(name string) bool {
	_, ok := m.providers[name]
	return ok
}

func (m *OAuthManager) EnabledProviders() []string {
	providers := make([]string, 0, len(m.providers))
	for name := range m.providers {
		providers = append(providers, name)
	}
	return providers
}

func (m *OAuthManager) GetAuthURL(providerName string) (string, string, error) {
	provider, ok := m.providers[providerName]
	if !ok {
		return "", "", fmt.Errorf("unknown OAuth provider: %s", providerName)
	}

	state, err := generateState()
	if err != nil {
		return "", "", fmt.Errorf("generating state: %w", err)
	}

	return provider.config.AuthCodeURL(state, oauth2.AccessTypeOffline), state, nil
}

// OAuthResult encapsulates the outcome of an OAuth callback. When the user has
// MFA enabled, TokenPair is nil and MFASession contains a short-lived JWT.
type OAuthResult struct {
	TokenPair  *TokenPair
	MFASession string
	Methods    []string
}

func (m *OAuthManager) HandleCallback(ctx context.Context, providerName, code string) (OAuthResult, error) {
	provider, ok := m.providers[providerName]
	if !ok {
		return OAuthResult{}, fmt.Errorf("unknown OAuth provider: %s", providerName)
	}

	token, err := provider.config.Exchange(ctx, code)
	if err != nil {
		return OAuthResult{}, fmt.Errorf("exchanging code: %w", err)
	}

	client := provider.config.Client(ctx, token)
	resp, err := client.Get(provider.userInfoURL)
	if err != nil {
		return OAuthResult{}, fmt.Errorf("fetching user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return OAuthResult{}, fmt.Errorf("user info returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return OAuthResult{}, fmt.Errorf("reading user info body: %w", err)
	}

	oUser, err := provider.parseUser(body)
	if err != nil {
		return OAuthResult{}, fmt.Errorf("parsing user info: %w", err)
	}

	// For GitHub, email may be private. Fetch from emails endpoint.
	if providerName == "github" && oUser.Email == "" {
		email, emailErr := fetchGitHubEmail(client)
		if emailErr == nil {
			oUser.Email = email
		}
	}

	if oUser.Email == "" {
		return OAuthResult{}, fmt.Errorf("could not retrieve email from %s", providerName)
	}

	// Upsert user
	u, err := m.upsertUser(ctx, providerName, oUser)
	if err != nil {
		return OAuthResult{}, fmt.Errorf("upserting user: %w", err)
	}

	// If MFA is enabled, return a pending session token instead of a full pair.
	if u.MFAEnabled {
		mfaSession, err := m.jwtManager.GenerateMFAPendingToken(u.ID, u.Email)
		if err != nil {
			return OAuthResult{}, fmt.Errorf("generating MFA session token: %w", err)
		}
		return OAuthResult{
			MFASession: mfaSession,
			Methods:    []string{"totp", "passkey", "backup"},
		}, nil
	}

	tokens, err := m.jwtManager.GenerateTokenPair(u.ID, u.Email, u.IsAdmin, u.EmailVerified, u.TokenVersion, u.MFAEnabled)
	if err != nil {
		return OAuthResult{}, err
	}
	return OAuthResult{TokenPair: &tokens}, nil
}

func (m *OAuthManager) upsertUser(ctx context.Context, providerName string, oUser oauthUser) (*user.User, error) {
	// Try to find by provider + provider_id
	existing, err := m.userRepo.GetByProvider(ctx, providerName, oUser.ID)
	if err == nil {
		return &existing, nil
	}

	// Try to find by email and link
	existing, err = m.userRepo.GetByEmail(ctx, oUser.Email)
	if err == nil {
		return &existing, nil
	}

	// Check if registration is enabled before creating a new user
	enabled, err := m.settingsRepo.IsRegistrationEnabled(ctx)
	if err != nil {
		return nil, fmt.Errorf("checking registration setting: %w", err)
	}
	if !enabled {
		return nil, fmt.Errorf("registration is currently disabled")
	}

	// Create new user (OAuth users are auto-verified via provider)
	avatarURL := oUser.AvatarURL
	providerID := oUser.ID
	u := &user.User{
		ID:            uuid.New(),
		Email:         oUser.Email,
		Name:          oUser.Name,
		AvatarURL:     &avatarURL,
		Provider:      providerName,
		ProviderID:    &providerID,
		EmailVerified: true,
	}

	if err := m.userRepo.Create(ctx, u); err != nil {
		return nil, fmt.Errorf("creating user: %w", err)
	}

	return u, nil
}

func fetchGitHubEmail(client *http.Client) (string, error) {
	resp, err := client.Get("https://api.github.com/user/emails")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var emails []struct {
		Email    string `json:"email"`
		Primary  bool   `json:"primary"`
		Verified bool   `json:"verified"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&emails); err != nil {
		return "", err
	}

	for _, e := range emails {
		if e.Primary && e.Verified {
			return e.Email, nil
		}
	}
	for _, e := range emails {
		if e.Verified {
			return e.Email, nil
		}
	}
	return "", fmt.Errorf("no verified email found")
}

func generateState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
