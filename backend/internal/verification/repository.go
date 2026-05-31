package verification

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository handles persistence for email verification tokens and email change requests.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a new verification Repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// CreateVerificationToken inserts a new email verification token.
func (r *Repository) CreateVerificationToken(ctx context.Context, userID uuid.UUID, token string, expiresAt time.Time) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO email_verification_tokens (user_id, token, expires_at)
		VALUES ($1, $2, $3)`,
		userID, token, expiresAt,
	)
	if err != nil {
		return fmt.Errorf("creating verification token: %w", err)
	}
	return nil
}

// GetVerificationByToken retrieves a verification token by its token string.
// Returns an error if the token is not found or has expired.
func (r *Repository) GetVerificationByToken(ctx context.Context, token string) (VerificationToken, error) {
	var vt VerificationToken
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, token, expires_at, created_at
		FROM email_verification_tokens
		WHERE token = $1 AND expires_at > NOW()`,
		token,
	).Scan(&vt.ID, &vt.UserID, &vt.Token, &vt.ExpiresAt, &vt.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return VerificationToken{}, fmt.Errorf("verification token not found or expired")
		}
		return VerificationToken{}, fmt.Errorf("querying verification token: %w", err)
	}
	return vt, nil
}

// DeleteVerificationTokensByUser removes all verification tokens for a given user.
func (r *Repository) DeleteVerificationTokensByUser(ctx context.Context, userID uuid.UUID) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM email_verification_tokens WHERE user_id = $1`,
		userID,
	)
	if err != nil {
		return fmt.Errorf("deleting verification tokens for user: %w", err)
	}
	return nil
}

// CreateEmailChangeRequest inserts a new email change request.
func (r *Repository) CreateEmailChangeRequest(ctx context.Context, userID uuid.UUID, newEmail, token string, expiresAt time.Time) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO email_change_requests (user_id, new_email, token, expires_at)
		VALUES ($1, $2, $3, $4)`,
		userID, newEmail, token, expiresAt,
	)
	if err != nil {
		return fmt.Errorf("creating email change request: %w", err)
	}
	return nil
}

// GetEmailChangeByToken retrieves an email change request by its token string.
// Returns an error if the token is not found or has expired.
func (r *Repository) GetEmailChangeByToken(ctx context.Context, token string) (EmailChangeRequest, error) {
	var ecr EmailChangeRequest
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, new_email, token, expires_at, created_at
		FROM email_change_requests
		WHERE token = $1 AND expires_at > NOW()`,
		token,
	).Scan(&ecr.ID, &ecr.UserID, &ecr.NewEmail, &ecr.Token, &ecr.ExpiresAt, &ecr.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return EmailChangeRequest{}, fmt.Errorf("email change request not found or expired")
		}
		return EmailChangeRequest{}, fmt.Errorf("querying email change request: %w", err)
	}
	return ecr, nil
}

// DeleteEmailChangeRequestsByUser removes all email change requests for a given user.
func (r *Repository) DeleteEmailChangeRequestsByUser(ctx context.Context, userID uuid.UUID) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM email_change_requests WHERE user_id = $1`,
		userID,
	)
	if err != nil {
		return fmt.Errorf("deleting email change requests for user: %w", err)
	}
	return nil
}

// PurgeExpired removes all expired tokens from both tables.
func (r *Repository) PurgeExpired(ctx context.Context) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM email_verification_tokens WHERE expires_at <= NOW()`,
	)
	if err != nil {
		return fmt.Errorf("purging expired verification tokens: %w", err)
	}

	_, err = r.pool.Exec(ctx,
		`DELETE FROM email_change_requests WHERE expires_at <= NOW()`,
	)
	if err != nil {
		return fmt.Errorf("purging expired email change requests: %w", err)
	}

	return nil
}
