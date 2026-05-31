package passwordreset

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/trackmycareer/app/internal/mailer"
	"github.com/trackmycareer/app/internal/password"
	"github.com/trackmycareer/app/internal/user"
)

// MFACodeVerifier is a function that verifies an MFA code for a given user.
// It is typically bound to mfa.Service.VerifyAnyMethod via a closure in main.go.
// When nil, MFA verification during password reset is not available.
type MFACodeVerifier func(ctx context.Context, userID uuid.UUID, method, code string) error

// Service provides business logic for password reset flows.
type Service struct {
	repo            *Repository
	userRepo        *user.Repository
	mailer          mailer.Mailer
	frontendURL     string
	mfaCodeVerifier MFACodeVerifier
}

// NewService creates a new password reset Service.
func NewService(repo *Repository, userRepo *user.Repository, mailer mailer.Mailer, frontendURL string, mfaCodeVerifier MFACodeVerifier) *Service {
	return &Service{
		repo:            repo,
		userRepo:        userRepo,
		mailer:          mailer,
		frontendURL:     frontendURL,
		mfaCodeVerifier: mfaCodeVerifier,
	}
}

// ValidateTokenAndGetUserID validates a reset token and returns the user ID
// without consuming the token. Used for passkey challenge generation during
// password reset with MFA.
func (s *Service) ValidateTokenAndGetUserID(ctx context.Context, rawToken string) (uuid.UUID, error) {
	hash := sha256.Sum256([]byte(rawToken))
	tokenHash := hex.EncodeToString(hash[:])

	token, err := s.repo.GetValidToken(ctx, tokenHash)
	if err != nil {
		return uuid.Nil, err
	}

	return token.UserID, nil
}

// RequestReset initiates a password reset for the given email address.
// Returns nil even if no user is found to prevent email enumeration.
func (s *Service) RequestReset(ctx context.Context, email string) error {
	u, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if strings.Contains(err.Error(), "user not found") {
			return nil
		}
		return fmt.Errorf("looking up user: %w", err)
	}

	// OAuth users receive a different notification.
	if u.Provider != "email" {
		if err := s.mailer.SendOAuthResetNotification(u.Email, u.Name, u.Provider); err != nil {
			return fmt.Errorf("sending OAuth reset notification: %w", err)
		}
		return nil
	}

	// Invalidate any existing unused tokens for this user.
	if err := s.repo.InvalidateTokensForUser(ctx, u.ID); err != nil {
		return fmt.Errorf("invalidating existing tokens: %w", err)
	}

	// Generate a cryptographically secure random token.
	rawBytes := make([]byte, 32)
	if _, err := rand.Read(rawBytes); err != nil {
		return fmt.Errorf("generating random token: %w", err)
	}
	rawTokenHex := hex.EncodeToString(rawBytes)

	// Store the SHA-256 hash, not the raw token.
	hash := sha256.Sum256([]byte(rawTokenHex))
	tokenHash := hex.EncodeToString(hash[:])

	expiresAt := time.Now().Add(1 * time.Hour)
	if err := s.repo.CreateToken(ctx, u.ID, tokenHash, expiresAt); err != nil {
		return fmt.Errorf("storing reset token: %w", err)
	}

	if err := s.mailer.SendPasswordResetEmail(u.Email, u.Name, rawTokenHex); err != nil {
		return fmt.Errorf("sending password reset email: %w", err)
	}

	return nil
}

// ResetPassword validates the reset token and updates the user's password.
// If the user has MFA enabled, it returns mfaRequired=true with the available
// MFA methods, deferring the actual password change until MFA is verified.
func (s *Service) ResetPassword(ctx context.Context, rawToken string, newPassword string) (mfaRequired bool, methods []string, err error) {
	// Hash the raw token to look up the stored record.
	hash := sha256.Sum256([]byte(rawToken))
	tokenHash := hex.EncodeToString(hash[:])

	token, err := s.repo.GetValidToken(ctx, tokenHash)
	if err != nil {
		return false, nil, err
	}

	u, err := s.userRepo.GetByID(ctx, token.UserID)
	if err != nil {
		return false, nil, fmt.Errorf("fetching user: %w", err)
	}

	// If MFA is enabled, the caller must provide MFA verification first.
	if u.MFAEnabled {
		return true, []string{"totp", "passkey", "backup"}, nil
	}

	// Validate password length.
	if len(newPassword) < 8 {
		return false, nil, fmt.Errorf("password must be at least 8 characters")
	}

	hashedPassword, err := password.Hash(newPassword)
	if err != nil {
		return false, nil, fmt.Errorf("hashing password: %w", err)
	}

	if err := s.userRepo.UpdatePassword(ctx, token.UserID, hashedPassword); err != nil {
		return false, nil, fmt.Errorf("updating password: %w", err)
	}

	if err := s.repo.MarkTokenUsed(ctx, token.ID); err != nil {
		return false, nil, fmt.Errorf("marking token as used: %w", err)
	}

	// Invalidate all existing sessions by incrementing token version.
	if err := s.userRepo.IncrementTokenVersion(ctx, token.UserID); err != nil {
		return false, nil, fmt.Errorf("incrementing token version: %w", err)
	}

	return false, nil, nil
}

// ResetPasswordWithMFA validates the reset token, verifies the MFA code, and
// updates the user's password. This method is used when the user has MFA
// enabled and the initial ResetPassword call returned mfaRequired=true.
func (s *Service) ResetPasswordWithMFA(ctx context.Context, rawToken, newPassword, mfaMethod, mfaCode string) error {
	// Hash the raw token to look up the stored record.
	hash := sha256.Sum256([]byte(rawToken))
	tokenHash := hex.EncodeToString(hash[:])

	token, err := s.repo.GetValidToken(ctx, tokenHash)
	if err != nil {
		return err
	}

	u, err := s.userRepo.GetByID(ctx, token.UserID)
	if err != nil {
		return fmt.Errorf("fetching user: %w", err)
	}

	if !u.MFAEnabled {
		return fmt.Errorf("MFA is not enabled for this account")
	}

	if s.mfaCodeVerifier == nil {
		return fmt.Errorf("MFA verification is not available")
	}

	// Verify the MFA code.
	if err := s.mfaCodeVerifier(ctx, u.ID, mfaMethod, mfaCode); err != nil {
		return fmt.Errorf("invalid verification code")
	}

	// Validate password length.
	if len(newPassword) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}

	hashedPassword, err := password.Hash(newPassword)
	if err != nil {
		return fmt.Errorf("hashing password: %w", err)
	}

	if err := s.userRepo.UpdatePassword(ctx, token.UserID, hashedPassword); err != nil {
		return fmt.Errorf("updating password: %w", err)
	}

	if err := s.repo.MarkTokenUsed(ctx, token.ID); err != nil {
		return fmt.Errorf("marking token as used: %w", err)
	}

	// Invalidate all existing sessions by incrementing token version.
	if err := s.userRepo.IncrementTokenVersion(ctx, token.UserID); err != nil {
		return fmt.Errorf("incrementing token version: %w", err)
	}

	return nil
}
