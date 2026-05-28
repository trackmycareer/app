package win

import (
	"time"

	"github.com/bhcloudlabs/trackmy-career/internal/tag"
	"github.com/bhcloudlabs/trackmy-career/pkg/types"
	"github.com/google/uuid"
)

type Win struct {
	ID          uuid.UUID  `json:"id"`
	UserID      uuid.UUID  `json:"-"`
	Title       string     `json:"title"`
	Description *string    `json:"description,omitempty"`
	OccurredOn  types.Date `json:"occurred_on"`
	Category    string     `json:"category"`
	Tags        []tag.Tag  `json:"tags"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type ListParams struct {
	Search   string
	Category string
	TagID    string
	From     string
	To       string
	Limit    int
	Offset   int
}
