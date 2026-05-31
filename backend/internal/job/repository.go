package job

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

func (r *Repository) List(ctx context.Context, userID uuid.UUID) ([]Job, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, user_id, company, title, start_date, end_date, employment_type,
			transition_type, location, work_mode, responsibilities, notes, sort_order,
			created_at, updated_at
		FROM jobs
		WHERE user_id = $1
		ORDER BY start_date DESC, created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("listing jobs: %w", err)
	}
	defer rows.Close()

	var jobs []Job
	for rows.Next() {
		var j Job
		if err := rows.Scan(
			&j.ID, &j.UserID, &j.Company, &j.Title, &j.StartDate, &j.EndDate,
			&j.EmploymentType, &j.TransitionType, &j.Location, &j.WorkMode,
			&j.Responsibilities, &j.Notes, &j.SortOrder, &j.CreatedAt, &j.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning job: %w", err)
		}
		jobs = append(jobs, j)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating jobs: %w", err)
	}

	return jobs, nil
}

func (r *Repository) GetByID(ctx context.Context, userID, jobID uuid.UUID) (Job, error) {
	var j Job
	err := r.pool.QueryRow(ctx,
		`SELECT id, user_id, company, title, start_date, end_date, employment_type,
			transition_type, location, work_mode, responsibilities, notes, sort_order,
			created_at, updated_at
		FROM jobs
		WHERE id = $1 AND user_id = $2`,
		jobID, userID,
	).Scan(
		&j.ID, &j.UserID, &j.Company, &j.Title, &j.StartDate, &j.EndDate,
		&j.EmploymentType, &j.TransitionType, &j.Location, &j.WorkMode,
		&j.Responsibilities, &j.Notes, &j.SortOrder, &j.CreatedAt, &j.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Job{}, fmt.Errorf("job not found")
		}
		return Job{}, fmt.Errorf("querying job: %w", err)
	}

	return j, nil
}

func (r *Repository) Create(ctx context.Context, j *Job) error {
	err := r.pool.QueryRow(ctx,
		`INSERT INTO jobs (id, user_id, company, title, start_date, end_date, employment_type,
			transition_type, location, work_mode, responsibilities, notes, sort_order)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING created_at, updated_at`,
		j.ID, j.UserID, j.Company, j.Title, j.StartDate, j.EndDate, j.EmploymentType,
		j.TransitionType, j.Location, j.WorkMode, j.Responsibilities, j.Notes, j.SortOrder,
	).Scan(&j.CreatedAt, &j.UpdatedAt)
	if err != nil {
		return fmt.Errorf("inserting job: %w", err)
	}

	return nil
}

func (r *Repository) Update(ctx context.Context, j *Job) error {
	result, err := r.pool.Exec(ctx,
		`UPDATE jobs SET company = $3, title = $4, start_date = $5, end_date = $6,
			employment_type = $7, transition_type = $8, location = $9, work_mode = $10,
			responsibilities = $11, notes = $12, sort_order = $13, updated_at = NOW()
		WHERE id = $1 AND user_id = $2`,
		j.ID, j.UserID, j.Company, j.Title, j.StartDate, j.EndDate, j.EmploymentType,
		j.TransitionType, j.Location, j.WorkMode, j.Responsibilities, j.Notes, j.SortOrder,
	)
	if err != nil {
		return fmt.Errorf("updating job: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("job not found")
	}

	return nil
}

func (r *Repository) Delete(ctx context.Context, userID, jobID uuid.UUID) error {
	result, err := r.pool.Exec(ctx,
		`DELETE FROM jobs WHERE id = $1 AND user_id = $2`,
		jobID, userID,
	)
	if err != nil {
		return fmt.Errorf("deleting job: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("job not found")
	}
	return nil
}
