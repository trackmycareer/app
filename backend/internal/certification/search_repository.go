package certification

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// SearchUserCertNames returns distinct certification name/provider pairs from
// the user's certifications that match the given query (case-insensitive
// substring match on name), limited to 5 results.
func (r *Repository) SearchUserCertNames(ctx context.Context, userID uuid.UUID, query string) ([]CertSearchResult, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT DISTINCT name, provider
		FROM certifications
		WHERE user_id = $1 AND name ILIKE '%' || $2 || '%'
		ORDER BY name
		LIMIT 5`,
		userID, query,
	)
	if err != nil {
		return nil, fmt.Errorf("searching user certification names: %w", err)
	}
	defer rows.Close()

	var results []CertSearchResult
	for rows.Next() {
		var r CertSearchResult
		if err := rows.Scan(&r.Name, &r.Provider); err != nil {
			return nil, fmt.Errorf("scanning certification search result: %w", err)
		}
		r.Source = "user"
		results = append(results, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating certification search results: %w", err)
	}

	return results, nil
}
