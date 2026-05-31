package location

import (
	"context"
	"log/slog"
	"strings"

	"github.com/google/uuid"
	"github.com/patrickmn/go-cache"
)

// Service coordinates location search across local user data and the Photon
// geocoder API, with caching for external results.
type Service struct {
	repo   *Repository
	cache  *cache.Cache
	client *PhotonClient
}

// NewService creates a new location Service.
func NewService(repo *Repository, cache *cache.Cache, client *PhotonClient) *Service {
	return &Service{repo: repo, cache: cache, client: client}
}

// Search returns location suggestions by combining the user's own locations
// (from jobs and profile) with results from the Photon geocoder.
// Duplicates are removed, preferring the user's own entries.
func (s *Service) Search(ctx context.Context, userID uuid.UUID, query string) (*SearchResponse, error) {
	// Fetch the user's own location names from jobs and profile.
	userNames, err := s.repo.SearchUserLocations(ctx, userID, query)
	if err != nil {
		return nil, err
	}

	userResults := make([]LocationResult, 0, len(userNames))
	seen := make(map[string]bool, len(userNames))
	for _, name := range userNames {
		userResults = append(userResults, LocationResult{
			Label:  name,
			Source: "user",
		})
		seen[strings.ToLower(name)] = true
	}

	// Look up Photon results, using the cache where possible.
	// Cache key is query-only (no user ID) because Photon data is public.
	cacheKey := strings.ToLower(query)
	var photonResults []LocationResult

	if cached, found := s.cache.Get(cacheKey); found {
		photonResults, _ = cached.([]LocationResult)
	} else {
		fetched, fetchErr := s.client.Search(ctx, query)
		if fetchErr != nil {
			// Graceful degradation: log and continue with user results only.
			slog.Error("photon geocoder search failed",
				"error", fetchErr.Error(),
			)
		} else {
			photonResults = fetched
			s.cache.SetDefault(cacheKey, photonResults)
		}
	}

	// Merge, deduplicating against the user's own locations.
	combined := userResults
	for _, r := range photonResults {
		if !seen[strings.ToLower(r.Label)] {
			combined = append(combined, r)
			seen[strings.ToLower(r.Label)] = true
		}
	}

	return &SearchResponse{Results: combined}, nil
}
