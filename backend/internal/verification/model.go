package verification

import (
	"time"

	"github.com/google/uuid"
)

// VerificationToken represents a pending email verification request.
type VerificationToken struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Token     string    `json:"-"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

// EmailChangeRequest represents a pending email address change awaiting confirmation.
type EmailChangeRequest struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	NewEmail  string    `json:"new_email"`
	Token     string    `json:"-"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}
