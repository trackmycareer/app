package jobtitle

import (
	"context"
	"strings"

	"github.com/google/uuid"
)

// titleSearcher is the interface the service requires from the repository,
// allowing test doubles to be injected.
type titleSearcher interface {
	SearchUserTitles(ctx context.Context, userID uuid.UUID, query string) ([]string, error)
}

// Service coordinates job title search across the user's own data and the
// curated static list of common titles.
type Service struct {
	repo titleSearcher
}

// NewService creates a new job title Service.
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// Search returns job title suggestions by combining the user's own titles
// (from their jobs) with matches from the curated common titles list.
// Duplicates are removed, preferring the user's own entries.
func (s *Service) Search(ctx context.Context, userID uuid.UUID, query string) (*SearchResponse, error) {
	// Fetch the user's own job titles.
	userTitles, err := s.repo.SearchUserTitles(ctx, userID, query)
	if err != nil {
		return nil, err
	}

	userResults := make([]JobTitleResult, 0, len(userTitles))
	seen := make(map[string]bool, len(userTitles))
	for _, title := range userTitles {
		userResults = append(userResults, JobTitleResult{
			Title:  title,
			Source: "user",
		})
		seen[strings.ToLower(title)] = true
	}

	// Filter the static common titles list by case-insensitive substring match.
	queryLower := strings.ToLower(query)
	var commonResults []JobTitleResult
	for _, title := range commonTitles {
		if len(commonResults) >= 10 {
			break
		}
		if seen[strings.ToLower(title)] {
			continue
		}
		if strings.Contains(strings.ToLower(title), queryLower) {
			commonResults = append(commonResults, JobTitleResult{
				Title:  title,
				Source: "common",
			})
		}
	}

	// Return combined: user titles first, then common titles.
	combined := make([]JobTitleResult, 0, len(userResults)+len(commonResults))
	combined = append(combined, userResults...)
	combined = append(combined, commonResults...)

	return &SearchResponse{Results: combined}, nil
}
