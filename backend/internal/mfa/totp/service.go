package totp

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/pquerna/otp/totp"

	"github.com/trackmycareer/app/internal/mfa/crypto"
)

// Service handles TOTP setup, verification, and validation.
type Service struct {
	repo      *Repository
	encryptor *crypto.Encryptor
	issuer    string
}

// NewService creates a new TOTP Service.
func NewService(repo *Repository, encryptor *crypto.Encryptor, issuer string) *Service {
	return &Service{
		repo:      repo,
		encryptor: encryptor,
		issuer:    issuer,
	}
}

// Setup generates a new TOTP secret for the user. If an unverified secret already
// exists it is replaced. Returns the provisioning URI and the plaintext secret.
func (s *Service) Setup(ctx context.Context, userID uuid.UUID, email string) (uri string, secret string, err error) {
	existing, getErr := s.repo.GetByUserID(ctx, userID)
	if getErr == nil {
		if existing.Verified {
			return "", "", fmt.Errorf("TOTP is already configured")
		}
		// Remove stale unverified secret before creating a new one.
		if delErr := s.repo.DeleteByUserID(ctx, userID); delErr != nil {
			return "", "", fmt.Errorf("removing unverified TOTP secret: %w", delErr)
		}
	}

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      s.issuer,
		AccountName: email,
	})
	if err != nil {
		return "", "", fmt.Errorf("generating TOTP key: %w", err)
	}

	ciphertext, nonce, err := s.encryptor.Encrypt([]byte(key.Secret()))
	if err != nil {
		return "", "", fmt.Errorf("encrypting TOTP secret: %w", err)
	}

	if err := s.repo.Create(ctx, userID, ciphertext, nonce); err != nil {
		return "", "", err
	}

	return key.URL(), key.Secret(), nil
}

// Verify validates a TOTP code during the setup flow and marks the secret as verified.
func (s *Service) Verify(ctx context.Context, userID uuid.UUID, code string) error {
	ts, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("fetching TOTP secret: %w", err)
	}

	plaintext, err := s.encryptor.Decrypt(ts.EncryptedSecret, ts.Nonce)
	if err != nil {
		return fmt.Errorf("decrypting TOTP secret: %w", err)
	}

	if !totp.Validate(code, string(plaintext)) {
		return fmt.Errorf("invalid verification code")
	}

	if err := s.repo.MarkVerified(ctx, ts.ID); err != nil {
		return err
	}

	return nil
}

// ValidateCode checks a TOTP code at login time without changing any state.
func (s *Service) ValidateCode(ctx context.Context, userID uuid.UUID, code string) (bool, error) {
	ts, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return false, nil
		}
		return false, fmt.Errorf("fetching TOTP secret: %w", err)
	}

	if !ts.Verified {
		return false, nil
	}

	plaintext, err := s.encryptor.Decrypt(ts.EncryptedSecret, ts.Nonce)
	if err != nil {
		return false, fmt.Errorf("decrypting TOTP secret: %w", err)
	}

	return totp.Validate(code, string(plaintext)), nil
}

// IsConfigured returns true if the user has a verified TOTP secret.
func (s *Service) IsConfigured(ctx context.Context, userID uuid.UUID) (bool, error) {
	ts, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return false, nil
		}
		return false, fmt.Errorf("checking TOTP configuration: %w", err)
	}
	return ts.Verified, nil
}

// Delete removes the TOTP secret for the given user.
func (s *Service) Delete(ctx context.Context, userID uuid.UUID) error {
	return s.repo.DeleteByUserID(ctx, userID)
}
