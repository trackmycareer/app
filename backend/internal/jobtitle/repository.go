package jobtitle

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository provides access to job title data stored in the user's existing
// jobs table.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a new job title Repository backed by the given
// connection pool.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// SearchUserTitles returns distinct job titles from the user's jobs that match
// the given query (case-insensitive substring match), limited to 5 results.
func (r *Repository) SearchUserTitles(ctx context.Context, userID uuid.UUID, query string) ([]string, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT DISTINCT title FROM jobs WHERE user_id = $1 AND title ILIKE '%' || $2 || '%' ORDER BY title LIMIT 5`,
		userID, query,
	)
	if err != nil {
		return nil, fmt.Errorf("searching user job titles: %w", err)
	}
	defer rows.Close()

	var titles []string
	for rows.Next() {
		var title string
		if err := rows.Scan(&title); err != nil {
			return nil, fmt.Errorf("scanning job title: %w", err)
		}
		titles = append(titles, title)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating job titles: %w", err)
	}

	return titles, nil
}
