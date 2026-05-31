package linkedaccount

import (
	"time"

	"github.com/google/uuid"
)

const (
	ProviderLinkedIn = "linkedin"
	ProviderGitHub   = "github"
	ProviderWebsite  = "website"
)

type LinkedAccount struct {
	ID          uuid.UUID  `json:"id"`
	UserID      uuid.UUID  `json:"user_id"`
	Provider    string     `json:"provider"`
	ProviderID  *string    `json:"-"`
	ProfileURL  string     `json:"profile_url"`
	Verified    bool       `json:"verified"`
	VerifyToken *string    `json:"verify_token,omitempty"`
	VerifiedAt  *time.Time `json:"verified_at,omitempty"`
	LastChecked *time.Time `json:"-"`
	CreatedAt   time.Time  `json:"created_at"`
}

type PublicLinkedAccount struct {
	Provider   string     `json:"provider"`
	ProfileURL string     `json:"profile_url"`
	Verified   bool       `json:"verified"`
	VerifiedAt *time.Time `json:"verified_at,omitempty"`
}
