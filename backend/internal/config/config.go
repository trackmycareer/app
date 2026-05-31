package config

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	Env             string `env:"ENV"               envDefault:"development"`
	Port            string `env:"PORT"              envDefault:"8080"`
	DatabaseURL     string `env:"DATABASE_URL,required"`
	JWTSecret       string `env:"JWT_SECRET,required"`
	AllowedOrigins  string `env:"ALLOWED_ORIGINS"   envDefault:"http://localhost:5173"`

	// OAuth - Google
	OAuthGoogleClientID     string `env:"OAUTH_GOOGLE_CLIENT_ID"`
	OAuthGoogleClientSecret string `env:"OAUTH_GOOGLE_CLIENT_SECRET"`

	// OAuth - GitHub
	OAuthGitHubClientID     string `env:"OAUTH_GITHUB_CLIENT_ID"`
	OAuthGitHubClientSecret string `env:"OAUTH_GITHUB_CLIENT_SECRET"`

	// OAuth - LinkedIn
	OAuthLinkedInClientID     string `env:"OAUTH_LINKEDIN_CLIENT_ID"`
	OAuthLinkedInClientSecret string `env:"OAUTH_LINKEDIN_CLIENT_SECRET"`

	OAuthRedirectBase string `env:"OAUTH_REDIRECT_BASE" envDefault:"http://localhost:8080"`
	FrontendURL       string `env:"FRONTEND_URL"        envDefault:"http://localhost:5173"`

	// Admin seed
	AdminEmail    string `env:"ADMIN_EMAIL"`
	AdminPassword string `env:"ADMIN_PASSWORD"`

	// SMTP / Email
	SMTPHost     string `env:"SMTP_HOST"`
	SMTPPort     string `env:"SMTP_PORT"     envDefault:"587"`
	SMTPUsername string `env:"SMTP_USERNAME"`
	SMTPPassword string `env:"SMTP_PASSWORD"`
	SMTPFrom     string `env:"SMTP_FROM"     envDefault:"noreply@trackmy.career"`

	// Companies House
	CompaniesHouseAPIKey  string `env:"COMPANIES_HOUSE_API_KEY"`
	CompaniesHouseBaseURL string `env:"COMPANIES_HOUSE_BASE_URL"`

	// Photon geocoder
	PhotonBaseURL string `env:"PHOTON_BASE_URL"`

	// S3 object storage
	S3Endpoint        string `env:"S3_ENDPOINT"         envDefault:"localhost:3900"`
	S3AccessKeyID     string `env:"S3_ACCESS_KEY_ID"    envDefault:"trackmy-dev-key"`
	S3SecretAccessKey string `env:"S3_SECRET_ACCESS_KEY" envDefault:"trackmy-dev-secret"`
	S3Bucket          string `env:"S3_BUCKET"            envDefault:"avatars"`
	S3UseSSL          bool   `env:"S3_USE_SSL"           envDefault:"false"`
	S3PublicURL       string `env:"S3_PUBLIC_URL"        envDefault:"http://localhost:3900"`
	S3Region          string `env:"S3_REGION"            envDefault:"garage"`

	// Trusted proxies (comma-separated CIDRs or IPs; empty = trust no proxies)
	TrustedProxies string `env:"TRUSTED_PROXIES"`

	// MFA encryption key (base64-encoded 32-byte key for AES-256-GCM)
	MFAEncryptionKey string `env:"MFA_ENCRYPTION_KEY"`

	// Polar supporter integration (optional — leave blank to disable)
	PolarAccessToken           string `env:"POLAR_ACCESS_TOKEN"`
	PolarWebhookSecret         string `env:"POLAR_WEBHOOK_SECRET"`
	PolarProductIDOneTime      string `env:"POLAR_PRODUCT_ID_ONE_TIME"`
	PolarProductIDSubscription string `env:"POLAR_PRODUCT_ID_SUBSCRIPTION"`
	PolarSandbox               bool   `env:"POLAR_SANDBOX" envDefault:"false"`
}

func (c Config) MFAKeyBytes() ([]byte, error) {
	if c.MFAEncryptionKey == "" {
		return nil, fmt.Errorf("MFA_ENCRYPTION_KEY is not set")
	}
	key, err := base64.StdEncoding.DecodeString(c.MFAEncryptionKey)
	if err != nil {
		return nil, fmt.Errorf("decoding MFA_ENCRYPTION_KEY: %w", err)
	}
	if len(key) != 32 {
		return nil, fmt.Errorf("MFA_ENCRYPTION_KEY must decode to exactly 32 bytes, got %d", len(key))
	}
	return key, nil
}

func (c Config) PolarEnabled() bool {
	return c.PolarAccessToken != ""
}

// TrustedProxyList returns the parsed list of trusted proxy addresses.
// An empty TRUSTED_PROXIES value yields nil, which tells Gin to trust no proxies.
func (c Config) TrustedProxyList() []string {
	if c.TrustedProxies == "" {
		return nil
	}
	parts := strings.Split(c.TrustedProxies, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func Load() (Config, error) {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}
