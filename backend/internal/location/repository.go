package location

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository provides access to location data stored in the user's existing
// jobs and users tables.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a new location Repository backed by the given connection pool.
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// SearchUserLocations returns distinct location names from the user's jobs and
// profile that match the given query (case-insensitive).
func (r *Repository) SearchUserLocations(ctx context.Context, userID uuid.UUID, query string) ([]string, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT DISTINCT name FROM (
			SELECT location AS name FROM jobs WHERE user_id = $1 AND location IS NOT NULL AND location ILIKE '%' || $2 || '%'
			UNION
			SELECT location AS name FROM users WHERE id = $1 AND location IS NOT NULL AND location ILIKE '%' || $2 || '%'
		) AS user_locations
		ORDER BY name
		LIMIT 5`,
		userID, query,
	)
	if err != nil {
		return nil, fmt.Errorf("searching user locations: %w", err)
	}
	defer rows.Close()

	var names []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("scanning location name: %w", err)
		}
		names = append(names, name)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating location names: %w", err)
	}

	return names, nil
}
