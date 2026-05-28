package tag

import "github.com/google/uuid"

type Tag struct {
	ID     uuid.UUID `json:"id"`
	UserID uuid.UUID `json:"-"`
	Name   string    `json:"name"`
	Colour string    `json:"colour"`
}
