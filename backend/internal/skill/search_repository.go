package skill

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// SearchUserSkillNames returns distinct skill name/category pairs from
// the user's skills that match the given query (case-insensitive
// substring match on name), limited to 5 results.
func (r *Repository) SearchUserSkillNames(ctx context.Context, userID uuid.UUID, query string) ([]SkillSearchResult, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT DISTINCT name, COALESCE(category, '') AS category
		FROM skills
		WHERE user_id = $1 AND name ILIKE '%' || $2 || '%'
		ORDER BY name
		LIMIT 5`,
		userID, query,
	)
	if err != nil {
		return nil, fmt.Errorf("searching user skill names: %w", err)
	}
	defer rows.Close()

	var results []SkillSearchResult
	for rows.Next() {
		var r SkillSearchResult
		if err := rows.Scan(&r.Name, &r.Category); err != nil {
			return nil, fmt.Errorf("scanning skill search result: %w", err)
		}
		r.Source = "user"
		results = append(results, r)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating skill search results: %w", err)
	}

	return results, nil
}
