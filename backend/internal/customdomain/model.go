package customdomain

import (
	"time"

	"github.com/google/uuid"
)

// CustomDomain represents a user's custom domain configuration for their
// public profile.
type CustomDomain struct {
	ID                   uuid.UUID  `json:"id"`
	UserID               uuid.UUID  `json:"-"`
	Domain               string     `json:"domain"`
	Status               string     `json:"status"`
	CloudflareHostnameID *string    `json:"-"`
	SSLStatus            string     `json:"ssl_status"`
	AccentColour         string     `json:"accent_colour"`
	CreatedAt            time.Time  `json:"created_at"`
	VerifiedAt           *time.Time `json:"verified_at,omitempty"`

	// CNAMETarget is populated at read time from config, not stored in the database.
	CNAMETarget string `json:"cname_target,omitempty"`
}
