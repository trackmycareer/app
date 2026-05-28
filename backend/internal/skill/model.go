package skill

import (
	"time"

	"github.com/google/uuid"
)

type Skill struct {
	ID          uuid.UUID       `json:"id"`
	UserID      uuid.UUID       `json:"-"`
	Name        string          `json:"name"`
	Category    *string         `json:"category,omitempty"`
	Proficiency int             `json:"proficiency"`
	Notes       *string         `json:"notes,omitempty"`
	Evidence    []SkillEvidence `json:"evidence"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

type SkillEvidence struct {
	ID           uuid.UUID `json:"id"`
	SkillID      uuid.UUID `json:"skill_id"`
	EvidenceType string    `json:"evidence_type"`
	EvidenceID   uuid.UUID `json:"evidence_id"`
	CreatedAt    time.Time `json:"created_at"`
}

type ListParams struct {
	Search   string
	Category string
	Limit    int
	Offset   int
}
