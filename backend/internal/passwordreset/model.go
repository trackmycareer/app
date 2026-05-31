package passwordreset

import (
	"time"

	"github.com/google/uuid"
)

type Token struct {
	ID        uuid.UUID  `json:"id"`
	UserID    uuid.UUID  `json:"user_id"`
	TokenHash string     `json:"-"`
	ExpiresAt time.Time  `json:"expires_at"`
	UsedAt    *time.Time `json:"-"`
	CreatedAt time.Time  `json:"created_at"`
}
