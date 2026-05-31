package totp

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/trackmycareer/app/internal/mfa"
)

// Repository handles persistence for TOTP secrets.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a new TOTP Repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// Create inserts a new unverified TOTP secret for the given user.
func (r *Repository) Create(ctx context.Context, userID uuid.UUID, encryptedSecret, nonce []byte) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO mfa_totp_secrets (user_id, encrypted_secret, nonce)
		VALUES ($1, $2, $3)`,
		userID, encryptedSecret, nonce,
	)
	if err != nil {
		return fmt.Errorf("creating TOTP secret: %w", err)
	}
	return nil
}

// GetByUserID retrieves the TOTP secret for a given user.
func (r *Repository) GetByUserID(ctx context.Context, userID uuid.UUID) (mfa.TOTPSecret, error) {
	var s mfa.TOTPSecret
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, encrypted_secret, nonce, verified, created_at
		FROM mfa_totp_secrets
		WHERE user_id = $1`,
		userID,
	).Scan(&s.ID, &s.UserID, &s.EncryptedSecret, &s.Nonce, &s.Verified, &s.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return mfa.TOTPSecret{}, fmt.Errorf("totp secret not found")
		}
		return mfa.TOTPSecret{}, fmt.Errorf("querying TOTP secret: %w", err)
	}
	return s, nil
}

// MarkVerified sets the verified flag to true for the given TOTP secret.
func (r *Repository) MarkVerified(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE mfa_totp_secrets SET verified = TRUE WHERE id = $1`,
		id,
	)
	if err != nil {
		return fmt.Errorf("marking TOTP secret as verified: %w", err)
	}
	return nil
}

// DeleteByUserID removes the TOTP secret for the given user.
func (r *Repository) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM mfa_totp_secrets WHERE user_id = $1`,
		userID,
	)
	if err != nil {
		return fmt.Errorf("deleting TOTP secret: %w", err)
	}
	return nil
}
