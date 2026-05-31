package certification

import (
	"context"
	"strings"

	"github.com/google/uuid"
)

// certSearcher is the interface the search logic requires from the repository,
// allowing test doubles to be injected.
type certSearcher interface {
	SearchUserCertNames(ctx context.Context, userID uuid.UUID, query string) ([]CertSearchResult, error)
}

// Search returns certification suggestions by combining the user's own
// certifications (from their data) with matches from the curated common
// certifications list. Duplicates are removed, preferring the user's own
// entries.
func (s *Service) Search(ctx context.Context, userID uuid.UUID, query string) (*CertSearchResponse, error) {
	return searchCerts(ctx, s.repo, userID, query)
}

// searchCerts contains the core search logic, accepting the certSearcher
// interface so it can be tested with mocks.
func searchCerts(ctx context.Context, repo certSearcher, userID uuid.UUID, query string) (*CertSearchResponse, error) {
	// Fetch the user's own certification names and providers.
	userResults, err := repo.SearchUserCertNames(ctx, userID, query)
	if err != nil {
		return nil, err
	}

	seen := make(map[string]bool, len(userResults))
	for _, r := range userResults {
		seen[strings.ToLower(r.Name)] = true
	}

	// Filter the static common certifications list by case-insensitive
	// substring match on name.
	queryLower := strings.ToLower(query)
	var commonResults []CertSearchResult
	for _, cert := range commonCerts {
		if len(commonResults) >= 10 {
			break
		}
		if seen[strings.ToLower(cert.Name)] {
			continue
		}
		if strings.Contains(strings.ToLower(cert.Name), queryLower) {
			commonResults = append(commonResults, CertSearchResult{
				Name:     cert.Name,
				Provider: cert.Provider,
				Source:   "common",
			})
		}
	}

	// Return combined: user results first, then common results.
	combined := make([]CertSearchResult, 0, len(userResults)+len(commonResults))
	combined = append(combined, userResults...)
	combined = append(combined, commonResults...)

	return &CertSearchResponse{Results: combined}, nil
}
