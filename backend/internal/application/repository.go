package application

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

const applicationColumns = `id, user_id, company, title, status, location, work_mode,
	job_url, source, salary, applied_date, notes, sort_order, created_at, updated_at`

func scanApplication(row pgx.Row, a *Application) error {
	return row.Scan(
		&a.ID, &a.UserID, &a.Company, &a.Title, &a.Status, &a.Location, &a.WorkMode,
		&a.JobURL, &a.Source, &a.Salary, &a.AppliedDate, &a.Notes, &a.SortOrder,
		&a.CreatedAt, &a.UpdatedAt,
	)
}

func (r *Repository) List(ctx context.Context, userID uuid.UUID) ([]Application, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+applicationColumns+`
		FROM applications
		WHERE user_id = $1
		ORDER BY status, sort_order, created_at`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("listing applications: %w", err)
	}
	defer rows.Close()

	var applications []Application
	for rows.Next() {
		var a Application
		if err := scanApplication(rows, &a); err != nil {
			return nil, fmt.Errorf("scanning application: %w", err)
		}
		applications = append(applications, a)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating applications: %w", err)
	}

	return applications, nil
}

func (r *Repository) GetByID(ctx context.Context, userID, applicationID uuid.UUID) (Application, error) {
	var a Application
	err := scanApplication(
		r.pool.QueryRow(ctx,
			`SELECT `+applicationColumns+`
			FROM applications
			WHERE id = $1 AND user_id = $2`,
			applicationID, userID,
		),
		&a,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Application{}, fmt.Errorf("application not found")
		}
		return Application{}, fmt.Errorf("querying application: %w", err)
	}

	return a, nil
}

func (r *Repository) Create(ctx context.Context, a *Application) error {
	err := r.pool.QueryRow(ctx,
		`INSERT INTO applications (id, user_id, company, title, status, location, work_mode,
			job_url, source, salary, applied_date, notes, sort_order)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING created_at, updated_at`,
		a.ID, a.UserID, a.Company, a.Title, a.Status, a.Location, a.WorkMode,
		a.JobURL, a.Source, a.Salary, a.AppliedDate, a.Notes, a.SortOrder,
	).Scan(&a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		return fmt.Errorf("inserting application: %w", err)
	}

	return nil
}

func (r *Repository) Update(ctx context.Context, a *Application) error {
	result, err := r.pool.Exec(ctx,
		`UPDATE applications SET company = $3, title = $4, status = $5, location = $6,
			work_mode = $7, job_url = $8, source = $9, salary = $10, applied_date = $11,
			notes = $12, sort_order = $13, updated_at = NOW()
		WHERE id = $1 AND user_id = $2`,
		a.ID, a.UserID, a.Company, a.Title, a.Status, a.Location, a.WorkMode,
		a.JobURL, a.Source, a.Salary, a.AppliedDate, a.Notes, a.SortOrder,
	)
	if err != nil {
		return fmt.Errorf("updating application: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("application not found")
	}

	return nil
}

// Move changes a single application's stage and position. Used when dropping a
// card without an explicit ordering (it is appended to the end of the column).
func (r *Repository) Move(ctx context.Context, userID, applicationID uuid.UUID, status string, sortOrder int) error {
	result, err := r.pool.Exec(ctx,
		`UPDATE applications SET status = $3, sort_order = $4, updated_at = NOW()
		WHERE id = $1 AND user_id = $2`,
		applicationID, userID, status, sortOrder,
	)
	if err != nil {
		return fmt.Errorf("moving application: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("application not found")
	}

	return nil
}

// MaxSortOrder returns the highest sort_order in a status column, or -1 when the
// column is empty, so callers can append with MaxSortOrder + 1.
func (r *Repository) MaxSortOrder(ctx context.Context, userID uuid.UUID, status string) (int, error) {
	var maxOrder int
	err := r.pool.QueryRow(ctx,
		`SELECT COALESCE(MAX(sort_order), -1) FROM applications WHERE user_id = $1 AND status = $2`,
		userID, status,
	).Scan(&maxOrder)
	if err != nil {
		return 0, fmt.Errorf("querying max sort order: %w", err)
	}
	return maxOrder, nil
}

// Reorder rewrites a whole column in one transaction: every id in orderedIDs is
// set to the given status with sort_order equal to its position. This handles
// both same-column reordering and a cross-column drop (the moved card is part of
// the destination column's orderedIDs) deterministically, with no collisions.
func (r *Repository) Reorder(ctx context.Context, userID uuid.UUID, status string, orderedIDs []uuid.UUID) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("beginning reorder transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	for i, id := range orderedIDs {
		if _, err := tx.Exec(ctx,
			`UPDATE applications SET status = $3, sort_order = $4, updated_at = NOW()
			WHERE id = $1 AND user_id = $2`,
			id, userID, status, i,
		); err != nil {
			return fmt.Errorf("reordering application %s: %w", id, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("committing reorder transaction: %w", err)
	}

	return nil
}

func (r *Repository) Delete(ctx context.Context, userID, applicationID uuid.UUID) error {
	result, err := r.pool.Exec(ctx,
		`DELETE FROM applications WHERE id = $1 AND user_id = $2`,
		applicationID, userID,
	)
	if err != nil {
		return fmt.Errorf("deleting application: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("application not found")
	}
	return nil
}
