package certification

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) List(ctx context.Context, userID uuid.UUID, params ListParams) ([]Certification, int, error) {
	args := []any{userID}
	where := `WHERE user_id = $1`
	argIdx := 2

	if params.Status != "" {
		where += fmt.Sprintf(` AND status = $%d`, argIdx)
		args = append(args, params.Status)
		argIdx++
	}

	if len(params.Statuses) > 0 {
		where += fmt.Sprintf(` AND status = ANY($%d)`, argIdx)
		args = append(args, params.Statuses)
		argIdx++
	}

	if params.Search != "" {
		where += fmt.Sprintf(` AND (name ILIKE $%d OR provider ILIKE $%d)`, argIdx, argIdx)
		args = append(args, "%"+params.Search+"%")
		argIdx++
	}

	// Count total
	var total int
	countQuery := `SELECT COUNT(*) FROM certifications ` + where
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("counting certifications: %w", err)
	}

	// Apply pagination
	limit := params.Limit
	if limit <= 0 {
		limit = 50
	}

	query := fmt.Sprintf(
		`SELECT id, user_id, name, provider, status, earned_date, expiry_date,
			cost, currency, credential_url, study_notes, study_progress,
			created_at, updated_at
		FROM certifications %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`,
		where, argIdx, argIdx+1,
	)
	args = append(args, limit, params.Offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("listing certifications: %w", err)
	}
	defer rows.Close()

	var certs []Certification
	for rows.Next() {
		var c Certification
		if err := rows.Scan(
			&c.ID, &c.UserID, &c.Name, &c.Provider, &c.Status,
			&c.EarnedDate, &c.ExpiryDate, &c.Cost, &c.Currency,
			&c.CredentialURL, &c.StudyNotes, &c.StudyProgress,
			&c.CreatedAt, &c.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scanning certification: %w", err)
		}
		certs = append(certs, c)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterating certifications: %w", err)
	}

	return certs, total, nil
}

func (r *Repository) GetByID(ctx context.Context, userID, certID uuid.UUID) (Certification, error) {
	var c Certification
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, name, provider, status, earned_date, expiry_date,
			cost, currency, credential_url, study_notes, study_progress,
			created_at, updated_at
		FROM certifications WHERE id = $1 AND user_id = $2`,
		certID, userID,
	).Scan(
		&c.ID, &c.UserID, &c.Name, &c.Provider, &c.Status,
		&c.EarnedDate, &c.ExpiryDate, &c.Cost, &c.Currency,
		&c.CredentialURL, &c.StudyNotes, &c.StudyProgress,
		&c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Certification{}, fmt.Errorf("certification not found")
		}
		return Certification{}, fmt.Errorf("querying certification: %w", err)
	}

	return c, nil
}

func (r *Repository) Create(ctx context.Context, c *Certification) error {
	err := r.pool.QueryRow(ctx,
		`INSERT INTO certifications (id, user_id, name, provider, status, earned_date, expiry_date,
			cost, currency, credential_url, study_notes, study_progress)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING created_at, updated_at`,
		c.ID, c.UserID, c.Name, c.Provider, c.Status, c.EarnedDate, c.ExpiryDate,
		c.Cost, c.Currency, c.CredentialURL, c.StudyNotes, c.StudyProgress,
	).Scan(&c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return fmt.Errorf("inserting certification: %w", err)
	}

	return nil
}

func (r *Repository) Update(ctx context.Context, c *Certification) error {
	result, err := r.pool.Exec(ctx,
		`UPDATE certifications SET name = $3, provider = $4, status = $5, earned_date = $6,
			expiry_date = $7, cost = $8, currency = $9, credential_url = $10,
			study_notes = $11, study_progress = $12, updated_at = NOW()
		WHERE id = $1 AND user_id = $2`,
		c.ID, c.UserID, c.Name, c.Provider, c.Status, c.EarnedDate, c.ExpiryDate,
		c.Cost, c.Currency, c.CredentialURL, c.StudyNotes, c.StudyProgress,
	)
	if err != nil {
		return fmt.Errorf("updating certification: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("certification not found")
	}

	return nil
}

func (r *Repository) UpdateStatus(ctx context.Context, userID, certID uuid.UUID, status string) error {
	result, err := r.pool.Exec(ctx,
		`UPDATE certifications SET status = $3, updated_at = NOW()
		WHERE id = $1 AND user_id = $2`,
		certID, userID, status,
	)
	if err != nil {
		return fmt.Errorf("updating certification status: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("certification not found")
	}

	return nil
}

func (r *Repository) Delete(ctx context.Context, userID, certID uuid.UUID) error {
	result, err := r.pool.Exec(ctx,
		`DELETE FROM certifications WHERE id = $1 AND user_id = $2`,
		certID, userID,
	)
	if err != nil {
		return fmt.Errorf("deleting certification: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("certification not found")
	}
	return nil
}
