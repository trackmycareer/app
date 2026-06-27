package notification

import (
	"time"

	"github.com/google/uuid"

	"github.com/trackmycareer/app/pkg/types"
)

// Notification type constants. The cert_expiry_* variants encode the lead-time
// band so the dedup key can distinguish a 90-day reminder from a 30- or 7-day
// one for the same certification.
const (
	TypeExpiry90 = "cert_expiry_90"
	TypeExpiry30 = "cert_expiry_30"
	TypeExpiry7  = "cert_expiry_7"
	TypeExpired  = "cert_expired"
)

// Notification is a single in-app notification record.
type Notification struct {
	ID                     uuid.UUID   `json:"id"`
	UserID                 uuid.UUID   `json:"-"`
	Type                   string      `json:"type"`
	Title                  string      `json:"title"`
	Body                   string      `json:"body"`
	RelatedCertificationID *uuid.UUID  `json:"related_certification_id,omitempty"`
	ReminderForDate        *types.Date `json:"reminder_for_date,omitempty"`
	ReadAt                 *time.Time  `json:"read_at,omitempty"`
	EmailSentAt            *time.Time  `json:"-"`
	CreatedAt              time.Time   `json:"created_at"`
}

// Preferences holds a user's renewal reminder settings.
type Preferences struct {
	Enabled      bool `json:"enabled"`
	Remind90     bool `json:"remind_90"`
	Remind30     bool `json:"remind_30"`
	Remind7      bool `json:"remind_7"`
	ChannelEmail bool `json:"channel_email"`
	ChannelInApp bool `json:"channel_in_app"`
}

// DefaultPreferences are returned when a user has no stored row. Reminders are
// on by default for users with a verified email.
func DefaultPreferences() Preferences {
	return Preferences{
		Enabled:      true,
		Remind90:     true,
		Remind30:     true,
		Remind7:      true,
		ChannelEmail: true,
		ChannelInApp: true,
	}
}

// ListParams controls notification listing.
type ListParams struct {
	UnreadOnly bool
	Limit      int
	Offset     int
}

// DueReminder is one row from the scheduler join: a certification within the
// reminder window for a user who should be reminded. It carries everything the
// reminder email needs so the scheduler depends only on the notification repo
// and the mailer. ApplicableType is computed from the day count and the per-band
// preferences (see applicableType); it is not read from the database.
type DueReminder struct {
	CertID         uuid.UUID
	UserID         uuid.UUID
	Email          string
	Name           string
	CertName       string
	Provider       string
	ExpiryDate     types.Date
	DaysRemaining  int
	CredentialURL  *string
	ChannelEmail   bool
	ChannelInApp   bool
	Remind90       bool
	Remind30       bool
	Remind7        bool
	ApplicableType string
}
