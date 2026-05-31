package backup

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/trackmycareer/app/internal/mfa"
)

// Repository handles persistence for MFA backup codes.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a new backup code Repository.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// DeleteAllForUser removes all backup codes for the given user.
func (r *Repository) DeleteAllForUser(ctx context.Context, userID uuid.UUID) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM mfa_backup_codes WHERE user_id = $1`,
		userID,
	)
	if err != nil {
		return fmt.Errorf("deleting backup codes: %w", err)
	}
	return nil
}

// CreateBatch inserts a batch of backup code hashes for the given user.
func (r *Repository) CreateBatch(ctx context.Context, userID uuid.UUID, codeHashes []string) error {
	query := `INSERT INTO mfa_backup_codes (user_id, code_hash) VALUES ($1, $2)`
	for _, hash := range codeHashes {
		if _, err := r.pool.Exec(ctx, query, userID, hash); err != nil {
			return fmt.Errorf("inserting backup code: %w", err)
		}
	}
	return nil
}

// CountUnused returns the number of unused backup codes for the given user.
func (r *Repository) CountUnused(ctx context.Context, userID uuid.UUID) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM mfa_backup_codes WHERE user_id = $1 AND used_at IS NULL`,
		userID,
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("counting unused backup codes: %w", err)
	}
	return count, nil
}

// GetUnusedByUser retrieves all unused backup codes for the given user.
func (r *Repository) GetUnusedByUser(ctx context.Context, userID uuid.UUID) ([]mfa.BackupCode, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, code_hash, used_at, created_at
		FROM mfa_backup_codes
		WHERE user_id = $1 AND used_at IS NULL
		ORDER BY created_at ASC`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("querying unused backup codes: %w", err)
	}
	defer rows.Close()

	var codes []mfa.BackupCode
	for rows.Next() {
		var c mfa.BackupCode
		if err := rows.Scan(&c.ID, &c.UserID, &c.CodeHash, &c.UsedAt, &c.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning backup code: %w", err)
		}
		codes = append(codes, c)
	}
	return codes, rows.Err()
}

// MarkUsed sets the used_at timestamp on a backup code.
func (r *Repository) MarkUsed(ctx context.Context, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE mfa_backup_codes SET used_at = NOW() WHERE id = $1`,
		id,
	)
	if err != nil {
		return fmt.Errorf("marking backup code as used: %w", err)
	}
	return nil
}
