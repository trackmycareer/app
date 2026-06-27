package compensation

import (
	"time"

	"github.com/google/uuid"

	"github.com/trackmycareer/app/pkg/types"
)

// Amounts holds the monetary fields of a compensation entry. Every value is an
// int64 in minor units (for example pennies or cents) so totals stay exact when
// summed for the trend chart. This struct is marshalled to JSON and encrypted at
// rest; it is never persisted in plaintext. Note is kept here, not as a column,
// because a compensation note can itself be sensitive.
type Amounts struct {
	Base   int64   `json:"base"`
	Bonus  int64   `json:"bonus"`
	Equity int64   `json:"equity"`
	Other  int64   `json:"other"`
	Note   *string `json:"note,omitempty"`
}

// Compensation is a single compensation event (a raise, a bonus, a new role's
// package) linked to a job. EncryptedData and Nonce are the at-rest form of
// Amounts and are never serialised to the client; the service populates Amounts
// by decrypting them on read.
type Compensation struct {
	ID            uuid.UUID  `json:"id"`
	UserID        uuid.UUID  `json:"-"`
	JobID         uuid.UUID  `json:"job_id"`
	EffectiveDate types.Date `json:"effective_date"`
	Currency      string     `json:"currency"`
	PayBasis      string     `json:"pay_basis"`
	Amounts       Amounts    `json:"amounts"`
	EncryptedData []byte     `json:"-"`
	Nonce         []byte     `json:"-"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}
