package job

import (
	"time"

	"github.com/google/uuid"

	"github.com/bhcloudlabs/trackmy-career/pkg/types"
)

type Job struct {
	ID               uuid.UUID   `json:"id"`
	UserID           uuid.UUID   `json:"-"`
	Company          string      `json:"company"`
	Title            string      `json:"title"`
	StartDate        types.Date  `json:"start_date"`
	EndDate          *types.Date `json:"end_date,omitempty"`
	EmploymentType   string      `json:"employment_type"`
	TransitionType   *string     `json:"transition_type,omitempty"`
	Location         *string     `json:"location,omitempty"`
	Remote           bool        `json:"remote"`
	Responsibilities *string     `json:"responsibilities,omitempty"`
	Notes            *string     `json:"notes,omitempty"`
	SortOrder        int         `json:"sort_order"`
	CreatedAt        time.Time   `json:"created_at"`
	UpdatedAt        time.Time   `json:"updated_at"`
}
