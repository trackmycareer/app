package mfa

import (
	"time"

	"github.com/google/uuid"
)

type Method string

const (
	MethodTOTP    Method = "totp"
	MethodPasskey Method = "passkey"
	MethodBackup  Method = "backup"
)

type Status struct {
	Enabled              bool     `json:"enabled"`
	Methods              []Method `json:"methods"`
	PasskeyCount         int      `json:"passkey_count"`
	BackupCodesRemaining int      `json:"backup_codes_remaining"`
}

type TOTPSecret struct {
	ID              uuid.UUID `json:"id"`
	UserID          uuid.UUID `json:"user_id"`
	EncryptedSecret []byte    `json:"-"`
	Nonce           []byte    `json:"-"`
	Verified        bool      `json:"verified"`
	CreatedAt       time.Time `json:"created_at"`
}

type Passkey struct {
	ID              uuid.UUID  `json:"id"`
	UserID          uuid.UUID  `json:"-"`
	CredentialID    []byte     `json:"-"`
	PublicKey       []byte     `json:"-"`
	AttestationType *string    `json:"-"`
	Transport       []string   `json:"-"`
	SignCount       uint32     `json:"-"`
	Name            string     `json:"name"`
	AAGUID          []byte     `json:"-"`
	CreatedAt       time.Time  `json:"created_at"`
	LastUsedAt      *time.Time `json:"last_used_at,omitempty"`
}

type BackupCode struct {
	ID        uuid.UUID  `json:"id"`
	UserID    uuid.UUID  `json:"-"`
	CodeHash  string     `json:"-"`
	UsedAt    *time.Time `json:"-"`
	CreatedAt time.Time  `json:"created_at"`
}

type WebAuthnChallenge struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	Challenge   []byte    `json:"-"`
	SessionData []byte    `json:"-"`
	Operation   string    `json:"operation"`
	ExpiresAt   time.Time `json:"expires_at"`
	CreatedAt   time.Time `json:"created_at"`
}
