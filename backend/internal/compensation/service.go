package compensation

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/trackmycareer/app/internal/mfa/crypto"
)

var validPayBases = []string{
	"annual", "monthly", "weekly", "daily", "hourly", "one_off",
}

const maxNoteLen = 2000

// repo is the persistence surface the service depends on. It is an interface so
// the service can be unit-tested with an in-memory fake while production uses the
// pgx-backed Repository.
type repo interface {
	List(ctx context.Context, userID uuid.UUID) ([]Compensation, error)
	GetByID(ctx context.Context, userID, id uuid.UUID) (Compensation, error)
	Create(ctx context.Context, c *Compensation) error
	Update(ctx context.Context, c *Compensation) error
	Delete(ctx context.Context, userID, id uuid.UUID) error
}

// Service validates compensation entries and owns the encryption boundary: the
// repository never sees plaintext amounts and the client never sees ciphertext.
type Service struct {
	repo      repo
	encryptor *crypto.Encryptor
}

func NewService(r repo, encryptor *crypto.Encryptor) *Service {
	return &Service{repo: r, encryptor: encryptor}
}

func (s *Service) List(ctx context.Context, userID uuid.UUID) ([]Compensation, error) {
	entries, err := s.repo.List(ctx, userID)
	if err != nil {
		return nil, err
	}
	for i := range entries {
		if err := s.decrypt(&entries[i]); err != nil {
			return nil, err
		}
	}
	return entries, nil
}

func (s *Service) GetByID(ctx context.Context, userID, id uuid.UUID) (Compensation, error) {
	c, err := s.repo.GetByID(ctx, userID, id)
	if err != nil {
		return Compensation{}, err
	}
	if err := s.decrypt(&c); err != nil {
		return Compensation{}, err
	}
	return c, nil
}

func (s *Service) Create(ctx context.Context, c *Compensation) error {
	if err := normaliseAndValidate(c); err != nil {
		return err
	}
	if err := s.encrypt(c); err != nil {
		return err
	}
	c.ID = uuid.New()
	return s.repo.Create(ctx, c)
}

func (s *Service) Update(ctx context.Context, c *Compensation) error {
	if err := normaliseAndValidate(c); err != nil {
		return err
	}
	if err := s.encrypt(c); err != nil {
		return err
	}
	return s.repo.Update(ctx, c)
}

func (s *Service) Delete(ctx context.Context, userID, id uuid.UUID) error {
	return s.repo.Delete(ctx, userID, id)
}

// normaliseAndValidate applies defaults and rejects bad input. It mutates the
// entry in place (uppercasing the currency, defaulting the pay basis, trimming
// the note).
func normaliseAndValidate(c *Compensation) error {
	c.Currency = strings.ToUpper(strings.TrimSpace(c.Currency))
	if c.Currency == "" {
		c.Currency = "GBP"
	}
	if !isValidCurrency(c.Currency) {
		return fmt.Errorf("currency must be a 3-letter ISO code")
	}

	c.PayBasis = strings.ToLower(strings.TrimSpace(c.PayBasis))
	if c.PayBasis == "" {
		c.PayBasis = "annual"
	}
	if !isValidPayBasis(c.PayBasis) {
		return fmt.Errorf("invalid pay_basis: %s", c.PayBasis)
	}

	if c.Amounts.Base < 0 || c.Amounts.Bonus < 0 || c.Amounts.Equity < 0 || c.Amounts.Other < 0 {
		return fmt.Errorf("amounts must be non-negative")
	}

	if c.Amounts.Note != nil {
		note := strings.TrimSpace(*c.Amounts.Note)
		if len(note) > maxNoteLen {
			return fmt.Errorf("note must be at most %d characters", maxNoteLen)
		}
		if note == "" {
			c.Amounts.Note = nil
		} else {
			c.Amounts.Note = &note
		}
	}

	return nil
}

func (s *Service) encrypt(c *Compensation) error {
	blob, err := json.Marshal(c.Amounts)
	if err != nil {
		return fmt.Errorf("marshalling amounts: %w", err)
	}
	ciphertext, nonce, err := s.encryptor.Encrypt(blob)
	if err != nil {
		return fmt.Errorf("encrypting compensation: %w", err)
	}
	c.EncryptedData = ciphertext
	c.Nonce = nonce
	return nil
}

func (s *Service) decrypt(c *Compensation) error {
	plaintext, err := s.encryptor.Decrypt(c.EncryptedData, c.Nonce)
	if err != nil {
		return fmt.Errorf("decrypting compensation: %w", err)
	}
	if err := json.Unmarshal(plaintext, &c.Amounts); err != nil {
		return fmt.Errorf("unmarshalling amounts: %w", err)
	}
	// The raw bytes are not serialised to the client and are not needed once
	// decrypted; drop them so a decrypted entry carries no ciphertext.
	c.EncryptedData = nil
	c.Nonce = nil
	return nil
}

func isValidCurrency(code string) bool {
	if len(code) != 3 {
		return false
	}
	for _, r := range code {
		if r < 'A' || r > 'Z' {
			return false
		}
	}
	return true
}

func isValidPayBasis(t string) bool {
	for _, v := range validPayBases {
		if v == t {
			return true
		}
	}
	return false
}
