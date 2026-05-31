package verification

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/trackmycareer/app/internal/mailer"
	"github.com/trackmycareer/app/internal/password"
	"github.com/trackmycareer/app/internal/user"
)

// Service provides business logic for email verification and email change flows.
type Service struct {
	repo     *Repository
	userRepo *user.Repository
	mailer   mailer.Mailer
}

// NewService creates a new verification Service.
func NewService(repo *Repository, userRepo *user.Repository, mailer mailer.Mailer) *Service {
	return &Service{
		repo:     repo,
		userRepo: userRepo,
		mailer:   mailer,
	}
}

// SendVerification generates a verification token and sends a verification email
// to the user's current email address.
func (s *Service) SendVerification(ctx context.Context, userID uuid.UUID) error {
	u, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("fetching user: %w", err)
	}

	if u.EmailVerified {
		return fmt.Errorf("email already verified")
	}

	if err := s.repo.DeleteVerificationTokensByUser(ctx, userID); err != nil {
		return fmt.Errorf("clearing existing tokens: %w", err)
	}

	token, err := generateToken()
	if err != nil {
		return fmt.Errorf("generating token: %w", err)
	}

	expiresAt := time.Now().Add(24 * time.Hour)
	if err := s.repo.CreateVerificationToken(ctx, userID, token, expiresAt); err != nil {
		return err
	}

	if err := s.mailer.SendVerificationEmail(u.Email, u.Name, token); err != nil {
		return fmt.Errorf("sending verification email: %w", err)
	}

	return nil
}

// VerifyEmail validates the token, marks the user's email as verified, and
// cleans up all verification tokens for that user. Returns the user ID so the
// caller can re-issue a JWT with updated claims.
func (s *Service) VerifyEmail(ctx context.Context, token string) (uuid.UUID, error) {
	vt, err := s.repo.GetVerificationByToken(ctx, token)
	if err != nil {
		return uuid.Nil, err
	}

	if err := s.userRepo.SetEmailVerified(ctx, vt.UserID); err != nil {
		return uuid.Nil, fmt.Errorf("setting email verified: %w", err)
	}

	if err := s.repo.DeleteVerificationTokensByUser(ctx, vt.UserID); err != nil {
		return uuid.Nil, fmt.Errorf("cleaning up verification tokens: %w", err)
	}

	return vt.UserID, nil
}

// RequestEmailChange initiates an email address change. The user must provide
// their current password for confirmation. A confirmation token is sent to the
// new email address.
func (s *Service) RequestEmailChange(ctx context.Context, userID uuid.UUID, newEmail, currentPassword string) error {
	u, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("fetching user: %w", err)
	}

	match, err := password.Verify(currentPassword, u.PasswordHash)
	if err != nil {
		return fmt.Errorf("verifying password: %w", err)
	}
	if !match {
		return fmt.Errorf("incorrect password")
	}

	// Check the new email is not already in use.
	_, err = s.userRepo.GetByEmail(ctx, newEmail)
	if err == nil {
		return fmt.Errorf("email address already in use")
	}
	if !strings.Contains(err.Error(), "user not found") {
		return fmt.Errorf("checking email availability: %w", err)
	}

	if err := s.repo.DeleteEmailChangeRequestsByUser(ctx, userID); err != nil {
		return fmt.Errorf("clearing existing email change requests: %w", err)
	}

	token, err := generateToken()
	if err != nil {
		return fmt.Errorf("generating token: %w", err)
	}

	expiresAt := time.Now().Add(1 * time.Hour)
	if err := s.repo.CreateEmailChangeRequest(ctx, userID, newEmail, token, expiresAt); err != nil {
		return err
	}

	if err := s.mailer.SendEmailChangeConfirmation(newEmail, u.Name, token); err != nil {
		return fmt.Errorf("sending email change confirmation: %w", err)
	}

	return nil
}

// ConfirmEmailChange validates the token, updates the user's email address,
// marks it as verified, and cleans up all email change requests. Returns the
// user ID so the caller can re-issue a JWT with updated claims.
func (s *Service) ConfirmEmailChange(ctx context.Context, token string) (uuid.UUID, error) {
	ecr, err := s.repo.GetEmailChangeByToken(ctx, token)
	if err != nil {
		return uuid.Nil, err
	}

	// Verify the new email is still available (could have been taken since the request).
	_, err = s.userRepo.GetByEmail(ctx, ecr.NewEmail)
	if err == nil {
		return uuid.Nil, fmt.Errorf("email address already in use")
	}
	if !strings.Contains(err.Error(), "user not found") {
		return uuid.Nil, fmt.Errorf("checking email availability: %w", err)
	}

	if err := s.userRepo.UpdateEmail(ctx, ecr.UserID, ecr.NewEmail); err != nil {
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique") {
			return uuid.Nil, fmt.Errorf("email address already in use")
		}
		return uuid.Nil, fmt.Errorf("updating email: %w", err)
	}

	if err := s.userRepo.SetEmailVerified(ctx, ecr.UserID); err != nil {
		return uuid.Nil, fmt.Errorf("setting email verified: %w", err)
	}

	if err := s.repo.DeleteEmailChangeRequestsByUser(ctx, ecr.UserID); err != nil {
		return uuid.Nil, fmt.Errorf("cleaning up email change requests: %w", err)
	}

	return ecr.UserID, nil
}

// generateToken produces a cryptographically secure 64-character hex token.
func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("reading random bytes: %w", err)
	}
	return hex.EncodeToString(b), nil
}
