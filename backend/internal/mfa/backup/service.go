package backup

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"

	"github.com/google/uuid"

	"github.com/trackmycareer/app/internal/mfa/crypto"
)

const (
	// codeCount is the number of backup codes generated per user.
	codeCount = 8
	// codeLength is the length of each backup code in characters.
	codeLength = 8
	// charset contains the allowed characters for backup codes.
	charset = "abcdefghijklmnopqrstuvwxyz0123456789"
)

// Service handles backup code generation and verification.
type Service struct {
	repo   *Repository
	mfaKey []byte
}

// NewService creates a new backup code Service.
func NewService(repo *Repository, mfaKey []byte) *Service {
	return &Service{
		repo:   repo,
		mfaKey: mfaKey,
	}
}

// Generate creates a new set of backup codes for the user, replacing any
// existing codes. Returns the plaintext codes (shown once to the user).
func (s *Service) Generate(ctx context.Context, userID uuid.UUID) ([]string, error) {
	// Remove any existing codes first.
	if err := s.repo.DeleteAllForUser(ctx, userID); err != nil {
		return nil, err
	}

	codes := make([]string, codeCount)
	hashes := make([]string, codeCount)

	for i := 0; i < codeCount; i++ {
		code, err := generateRandomCode()
		if err != nil {
			return nil, fmt.Errorf("generating backup code: %w", err)
		}
		codes[i] = code
		hashes[i] = crypto.HashBackupCode(s.mfaKey, code)
	}

	if err := s.repo.CreateBatch(ctx, userID, hashes); err != nil {
		return nil, err
	}

	return codes, nil
}

// VerifyAndConsume checks whether the provided code matches any unused backup
// code for the user. If a match is found the code is marked as used.
func (s *Service) VerifyAndConsume(ctx context.Context, userID uuid.UUID, code string) (bool, error) {
	unused, err := s.repo.GetUnusedByUser(ctx, userID)
	if err != nil {
		return false, err
	}

	for _, bc := range unused {
		if crypto.VerifyBackupCode(s.mfaKey, code, bc.CodeHash) {
			if markErr := s.repo.MarkUsed(ctx, bc.ID); markErr != nil {
				return false, fmt.Errorf("marking backup code as used: %w", markErr)
			}
			return true, nil
		}
	}

	return false, nil
}

// CountRemaining returns the number of unused backup codes for the user.
func (s *Service) CountRemaining(ctx context.Context, userID uuid.UUID) (int, error) {
	return s.repo.CountUnused(ctx, userID)
}

// Regenerate deletes all existing codes and generates a fresh set.
func (s *Service) Regenerate(ctx context.Context, userID uuid.UUID) ([]string, error) {
	return s.Generate(ctx, userID)
}

// generateRandomCode creates a cryptographically random alphanumeric code.
func generateRandomCode() (string, error) {
	max := big.NewInt(int64(len(charset)))
	result := make([]byte, codeLength)
	for i := range result {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		result[i] = charset[n.Int64()]
	}
	return string(result), nil
}
