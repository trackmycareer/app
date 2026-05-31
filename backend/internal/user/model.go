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
	Location          *string         `json:"location,omitempty"`
	Headline          *string         `json:"headline,omitempty"`
	OpenToWork        string          `json:"open_to_work"`
	ProfileVisibility json.RawMessage `json:"profile_visibility"`
	EmailVerified     bool            `json:"email_verified"`
	EmailVerifiedAt   *time.Time      `json:"email_verified_at,omitempty"`
	NewsletterOptIn    bool            `json:"newsletter_opt_in"`
	NewsletterOptInAt  *time.Time      `json:"newsletter_opt_in_at,omitempty"`
	IsOneTimeSupporter bool            `json:"is_one_time_supporter"`
	IsSubscriber       bool            `json:"is_subscriber"`
	SupporterSince     *time.Time      `json:"supporter_since,omitempty"`
	PolarCustomerID    *string         `json:"-"`
	TokenVersion       int             `json:"-"`
	CreatedAt          time.Time       `json:"created_at"`
	UpdatedAt          time.Time       `json:"updated_at"`
}

func (u User) DerefPolarCustomerID() string {
	if u.PolarCustomerID != nil {
		return *u.PolarCustomerID
	}
	return ""
}
