package user

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID                uuid.UUID       `json:"id"`
	Email             string          `json:"email"`
	PasswordHash      string          `json:"-"`
	Name              string          `json:"name"`
	AvatarURL         *string         `json:"avatar_url,omitempty"`
	Provider          string          `json:"provider"`
	ProviderID        *string         `json:"-"`
	IsAdmin           bool            `json:"is_admin"`
	Username          *string         `json:"username,omitempty"`
	Bio               *string         `json:"bio,omitempty"`
	ProfileVisibility json.RawMessage `json:"profile_visibility"`
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
}
