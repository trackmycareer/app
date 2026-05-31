package skill

import (
	"context"
	"strings"

	"github.com/google/uuid"
)

// skillSearcher is the interface the search logic requires from the repository,
// allowing test doubles to be injected.
type skillSearcher interface {
	SearchUserSkillNames(ctx context.Context, userID uuid.UUID, query string) ([]SkillSearchResult, error)
}

// Search returns skill suggestions by combining the user's own
// skills (from their data) with matches from the curated common
// skills list. Duplicates are removed, preferring the user's own
// entries.
func (s *Service) Search(ctx context.Context, userID uuid.UUID, query string) (*SkillSearchResponse, error) {
	return searchSkills(ctx, s.repo, userID, query)
}

// searchSkills contains the core search logic, accepting the skillSearcher
// interface so it can be tested with mocks.
func searchSkills(ctx context.Context, repo skillSearcher, userID uuid.UUID, query string) (*SkillSearchResponse, error) {
	// Fetch the user's own skill names and categories.
	userResults, err := repo.SearchUserSkillNames(ctx, userID, query)
	if err != nil {
		return nil, err
	}

	seen := make(map[string]bool, len(userResults))
	for _, r := range userResults {
		seen[strings.ToLower(r.Name)] = true
	}

	// Filter the static common skills list by case-insensitive
	// substring match on name.
	queryLower := strings.ToLower(query)
	var commonResults []SkillSearchResult
	for _, sk := range commonSkills {
		if len(commonResults) >= 10 {
			break
		}
		if seen[strings.ToLower(sk.Name)] {
			continue
		}
		if strings.Contains(strings.ToLower(sk.Name), queryLower) {
			commonResults = append(commonResults, SkillSearchResult{
				Name:     sk.Name,
				Category: sk.Category,
				Source:   "common",
			})
		}
	}

	// Return combined: user results first, then common results.
	combined := make([]SkillSearchResult, 0, len(userResults)+len(commonResults))
	combined = append(combined, userResults...)
	combined = append(combined, commonResults...)

	return &SkillSearchResponse{Results: combined}, nil
}
