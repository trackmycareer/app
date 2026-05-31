package company

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository provides access to company-related data stored in the user's
// existing jobs and certifications tables.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a new company Repository backed by the given connection pool.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// SearchUserCompanies returns distinct company/provider names from the user's
// jobs and certifications that match the given query (case-insensitive).
func (r *Repository) SearchUserCompanies(ctx context.Context, userID uuid.UUID, query string) ([]string, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT DISTINCT name FROM (
			SELECT company AS name FROM jobs WHERE user_id = $1 AND company ILIKE '%' || $2 || '%'
			UNION
			SELECT provider AS name FROM certifications WHERE user_id = $1 AND provider ILIKE '%' || $2 || '%'
		) AS user_companies
		ORDER BY name
		LIMIT 5`,
		userID, query,
	)
	if err != nil {
		return nil, fmt.Errorf("searching user companies: %w", err)
	}
	defer rows.Close()

	var names []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("scanning company name: %w", err)
		}
		names = append(names, name)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating company names: %w", err)
	}

	return names, nil
}
