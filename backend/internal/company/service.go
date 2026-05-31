package company

import (
	"context"
	"log/slog"
	"strings"

	"github.com/google/uuid"
	"github.com/patrickmn/go-cache"
)

// Service coordinates company search across local user data and the
// Companies House external API, with caching for external results.
type Service struct {
	repo   *Repository
	cache  *cache.Cache
	client *CompaniesHouseClient
}

// NewService creates a new company Service.
func NewService(repo *Repository, cache *cache.Cache, client *CompaniesHouseClient) *Service {
	return &Service{repo: repo, cache: cache, client: client}
}

// Search returns company suggestions by combining the user's own companies
// (from jobs and certifications) with results from the Companies House API.
// Duplicates are removed, preferring the user's own entries.
func (s *Service) Search(ctx context.Context, userID uuid.UUID, query string) (*SearchResponse, error) {
	// Fetch the user's own company names from jobs and certifications.
	userNames, err := s.repo.SearchUserCompanies(ctx, userID, query)
	if err != nil {
		return nil, err
	}

	userResults := make([]CompanyResult, 0, len(userNames))
	seen := make(map[string]bool, len(userNames))
	for _, name := range userNames {
		userResults = append(userResults, CompanyResult{
			Name:   name,
			Source: "user",
		})
		seen[strings.ToLower(name)] = true
	}

	// Look up Companies House results, using the cache where possible.
	// Cache key is query-only (no user ID) because Companies House data is a public register.
	cacheKey := strings.ToLower(query)
	var chResults []CompanyResult

	if cached, found := s.cache.Get(cacheKey); found {
		chResults, _ = cached.([]CompanyResult)
	} else {
		fetched, fetchErr := s.client.Search(ctx, query)
		if fetchErr != nil {
			// Graceful degradation: log and continue with user results only.
			slog.Error("companies house search failed",
				"error", fetchErr.Error(),
				"query", query,
			)
		} else {
			chResults = fetched
			s.cache.SetDefault(cacheKey, chResults)
		}
	}

	// Merge, deduplicating against the user's own companies.
	combined := userResults
	for _, r := range chResults {
		if !seen[strings.ToLower(r.Name)] {
			combined = append(combined, r)
			seen[strings.ToLower(r.Name)] = true
		}
	}

	return &SearchResponse{Results: combined}, nil
}
