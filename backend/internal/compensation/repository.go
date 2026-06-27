package compensation

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository is a pure persistence layer for compensation rows. It stores and
// returns the encrypted blob unchanged; encryption and decryption are the
// service's responsibility.
type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

const selectColumns = `id, user_id, job_id, effective_date, currency, pay_basis,
	encrypted_data, nonce, created_at, updated_at`

func scanRow(row pgx.Row, c *Compensation) error {
	return row.Scan(
		&c.ID, &c.UserID, &c.JobID, &c.EffectiveDate, &c.Currency, &c.PayBasis,
		&c.EncryptedData, &c.Nonce, &c.CreatedAt, &c.UpdatedAt,
	)
}

func (r *Repository) List(ctx context.Context, userID uuid.UUID) ([]Compensation, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+selectColumns+`
		FROM compensation
		WHERE user_id = $1
		ORDER BY effective_date DESC, created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("listing compensation: %w", err)
	}
	defer rows.Close()

	var entries []Compensation
	for rows.Next() {
		var c Compensation
		if err := scanRow(rows, &c); err != nil {
			return nil, fmt.Errorf("scanning compensation: %w", err)
		}
		entries = append(entries, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating compensation: %w", err)
	}

	return entries, nil
}

func (r *Repository) GetByID(ctx context.Context, userID, id uuid.UUID) (Compensation, error) {
	var c Compensation
	err := scanRow(r.pool.QueryRow(ctx,
		`SELECT `+selectColumns+`
		FROM compensation
		WHERE id = $1 AND user_id = $2`,
		id, userID,
	), &c)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Compensation{}, fmt.Errorf("compensation not found")
		}
		return Compensation{}, fmt.Errorf("querying compensation: %w", err)
	}

	return c, nil
}

// Create inserts a compensation row only if the target job belongs to the same
// user. The EXISTS guard means the foreign key alone cannot be used to attach
// compensation to another user's job; a failed guard returns "job not found".
func (r *Repository) Create(ctx context.Context, c *Compensation) error {
	err := r.pool.QueryRow(ctx,
		`INSERT INTO compensation (id, user_id, job_id, effective_date, currency, pay_basis, encrypted_data, nonce)
		SELECT $1, $2, $3, $4, $5, $6, $7, $8
		WHERE EXISTS (SELECT 1 FROM jobs WHERE id = $3 AND user_id = $2)
		RETURNING created_at, updated_at`,
		c.ID, c.UserID, c.JobID, c.EffectiveDate, c.Currency, c.PayBasis, c.EncryptedData, c.Nonce,
	).Scan(&c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("job not found")
		}
		return fmt.Errorf("inserting compensation: %w", err)
	}

	return nil
}

// Update writes a compensation row scoped to its owner, again guarding that the
// (possibly changed) job belongs to the user. Callers verify the row exists
// first, so a zero-row result means the job is not owned.
func (r *Repository) Update(ctx context.Context, c *Compensation) error {
	result, err := r.pool.Exec(ctx,
		`UPDATE compensation
		SET job_id = $3, effective_date = $4, currency = $5, pay_basis = $6,
			encrypted_data = $7, nonce = $8, updated_at = NOW()
		WHERE id = $1 AND user_id = $2
			AND EXISTS (SELECT 1 FROM jobs WHERE id = $3 AND user_id = $2)`,
		c.ID, c.UserID, c.JobID, c.EffectiveDate, c.Currency, c.PayBasis, c.EncryptedData, c.Nonce,
	)
	if err != nil {
		return fmt.Errorf("updating compensation: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("job not found")
	}

	return nil
}

func (r *Repository) Delete(ctx context.Context, userID, id uuid.UUID) error {
	result, err := r.pool.Exec(ctx,
		`DELETE FROM compensation WHERE id = $1 AND user_id = $2`,
		id, userID,
	)
	if err != nil {
		return fmt.Errorf("deleting compensation: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("compensation not found")
	}
	return nil
}
