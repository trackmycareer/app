package passwordreset

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository handles persistence for password reset tokens.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a new password reset Repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// InvalidateTokensForUser marks all unused tokens for the given user as used.
func (r *Repository) InvalidateTokensForUser(ctx context.Context, userID uuid.UUID) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE password_reset_tokens SET used_at = NOW() WHERE user_id = $1 AND used_at IS NULL`,
		userID,
	)
	if err != nil {
		return fmt.Errorf("invalidating reset tokens for user: %w", err)
	}
	return nil
}

// CreateToken inserts a new password reset token.
func (r *Repository) CreateToken(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO password_reset_tokens (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)`,
		userID, tokenHash, expiresAt,
	)
	if err != nil {
		return fmt.Errorf("creating reset token: %w", err)
	}
	return nil
}

// GetValidToken retrieves a password reset token by its hash.
// Only returns tokens that have not expired and have not been used.
func (r *Repository) GetValidToken(ctx context.Context, tokenHash string) (Token, error) {
	var t Token
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, token_hash, expires_at, used_at, created_at
		FROM password_reset_tokens
		WHERE token_hash = $1 AND expires_at > NOW() AND used_at IS NULL`,
		tokenHash,
	).Scan(&t.ID, &t.UserID, &t.TokenHash, &t.ExpiresAt, &t.UsedAt, &t.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Token{}, fmt.Errorf("token not found")
		}
		return Token{}, fmt.Errorf("querying reset token: %w", err)
	}
	return t, nil
}

// MarkTokenUsed sets the used_at timestamp on the given token.
func (r *Repository) MarkTokenUsed(ctx context.Context, tokenID uuid.UUID) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE password_reset_tokens SET used_at = NOW() WHERE id = $1`,
		tokenID,
	)
	if err != nil {
		return fmt.Errorf("marking reset token as used: %w", err)
	}
	return nil
}
