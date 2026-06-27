package application

import (
	"time"

	"github.com/google/uuid"

	"github.com/trackmycareer/app/pkg/types"
)

// Application is a job application a user is tracking through their pipeline.
// It is a prospect, distinct from a held role (see the job package), and is
// always private: it never appears on the public profile or in data export.
type Application struct {
	ID          uuid.UUID   `json:"id"`
	UserID      uuid.UUID   `json:"-"`
	Company     string      `json:"company"`
	Title       string      `json:"title"`
	Status      string      `json:"status"`
	Location    *string     `json:"location,omitempty"`
	WorkMode    *string     `json:"work_mode,omitempty"`
	JobURL      *string     `json:"job_url,omitempty"`
	Source      *string     `json:"source,omitempty"`
	Salary      *string     `json:"salary,omitempty"`
	AppliedDate *types.Date `json:"applied_date,omitempty"`
	Notes       *string     `json:"notes,omitempty"`
	SortOrder   int         `json:"sort_order"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}
