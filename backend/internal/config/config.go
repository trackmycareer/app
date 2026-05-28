package config

import (
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

	OAuthRedirectBase string `env:"OAUTH_REDIRECT_BASE" envDefault:"http://localhost:8080"`
	FrontendURL       string `env:"FRONTEND_URL"        envDefault:"http://localhost:5173"`

	// Admin seed
	AdminEmail    string `env:"ADMIN_EMAIL"`
	AdminPassword string `env:"ADMIN_PASSWORD"`
}

func Load() (Config, error) {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}
