package certification

import (
	"time"

	"github.com/google/uuid"

	"github.com/trackmycareer/app/pkg/types"
)

type Certification struct {
	ID            uuid.UUID   `json:"id"`
	UserID        uuid.UUID   `json:"-"`
	Name          string      `json:"name"`
	Provider      string      `json:"provider"`
	Status        string      `json:"status"`
	EarnedDate    *types.Date `json:"earned_date,omitempty"`
	ExpiryDate    *types.Date `json:"expiry_date,omitempty"`
	Cost          *float64    `json:"cost,omitempty"`
	Currency      string      `json:"currency"`
	CredentialURL *string     `json:"credential_url,omitempty"`
	StudyNotes    *string     `json:"study_notes,omitempty"`
	StudyProgress int         `json:"study_progress"`
	CreatedAt     time.Time   `json:"created_at"`
	UpdatedAt     time.Time   `json:"updated_at"`
}

type ListParams struct {
	Search   string
	Status   string
	Statuses []string
	Limit    int
	Offset   int
}
