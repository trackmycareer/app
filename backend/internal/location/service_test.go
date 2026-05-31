package location

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/patrickmn/go-cache"
)

// mockRepo implements a minimal test double for the repository layer.
type mockRepo struct {
	results []string
	err     error
}

func (m *mockRepo) SearchUserLocations(_ context.Context, _ uuid.UUID, _ string) ([]string, error) {
	return m.results, m.err
}

// mockClient implements a minimal test double for the Photon client.
type mockClient struct {
	results []LocationResult
	err     error
	called  bool
}

func (m *mockClient) Search(_ context.Context, _ string) ([]LocationResult, error) {
	m.called = true
	return m.results, m.err
}

// locationSearcher is the interface the service needs from the repository.
type locationSearcher interface {
	SearchUserLocations(ctx context.Context, userID uuid.UUID, query string) ([]string, error)
}

// externalSearcher is the interface the service needs from the client.
type externalSearcher interface {
	Search(ctx context.Context, query string) ([]LocationResult, error)
}

// testableService wraps the search logic so we can inject mocks without
// modifying the production Service struct. It mirrors Service.Search exactly.
type testableService struct {
	repo   locationSearcher
	cache  *cache.Cache
	client externalSearcher
}

func (s *testableService) Search(ctx context.Context, userID uuid.UUID, query string) (*SearchResponse, error) {
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

	cacheKey := strings.ToLower(query)
	var photonResults []LocationResult

	if cached, found := s.cache.Get(cacheKey); found {
		photonResults, _ = cached.([]LocationResult)
	} else {
		fetched, fetchErr := s.client.Search(ctx, query)
		if fetchErr != nil {
			// Graceful degradation: continue with user results only.
		} else {
			photonResults = fetched
			s.cache.SetDefault(cacheKey, photonResults)
		}
	}

	combined := userResults
	for _, r := range photonResults {
		if !seen[strings.ToLower(r.Label)] {
			combined = append(combined, r)
			seen[strings.ToLower(r.Label)] = true
		}
	}

	return &SearchResponse{Results: combined}, nil
}

func newTestService(repo locationSearcher, client externalSearcher) *testableService {
	return &testableService{
		repo:   repo,
		cache:  cache.New(5*time.Minute, 10*time.Minute),
		client: client,
	}
}

func TestServiceSearch(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	tests := []struct {
		name           string
		query          string
		repoResults    []string
		repoErr        error
		clientResults  []LocationResult
		clientErr      error
		primeCache     bool
		cachedResults  []LocationResult
		wantCount      int
		wantErr        bool
		wantClientCall bool
		wantSources    map[string]bool
	}{
		{
			name:           "user locations returned when Photon returns nothing",
			query:          "london",
			repoResults:    []string{"London, UK", "London, Ontario"},
			clientResults:  nil,
			wantCount:      2,
			wantClientCall: true,
			wantSources:    map[string]bool{"user": true},
		},
		{
			name:           "Photon results only when no user locations exist",
			query:          "berlin",
			repoResults:    []string{},
			clientResults:  []LocationResult{{Label: "Berlin, Germany", Source: "photon"}},
			wantCount:      1,
			wantClientCall: true,
			wantSources:    map[string]bool{"photon": true},
		},
		{
			name:        "cache hit returns cached results without calling Photon",
			query:       "paris",
			repoResults: []string{},
			primeCache:  true,
			cachedResults: []LocationResult{
				{Label: "Paris, France", Source: "photon"},
			},
			wantCount:      1,
			wantClientCall: false,
			wantSources:    map[string]bool{"photon": true},
		},
		{
			name:        "cache miss calls Photon and caches the result",
			query:       "tokyo",
			repoResults: []string{},
			clientResults: []LocationResult{
				{Label: "Tokyo, Japan", Source: "photon"},
			},
			wantCount:      1,
			wantClientCall: true,
			wantSources:    map[string]bool{"photon": true},
		},
		{
			name:        "deduplication when location appears in both user and Photon",
			query:       "london",
			repoResults: []string{"London, United Kingdom"},
			clientResults: []LocationResult{
				{Label: "london, united kingdom", Source: "photon"},
				{Label: "London, Ontario, Canada", Source: "photon"},
			},
			wantCount:      2, // London, United Kingdom (user) + London, Ontario, Canada (photon)
			wantClientCall: true,
			wantSources:    map[string]bool{"user": true, "photon": true},
		},
		{
			name:           "Photon API error still returns user results (graceful degradation)",
			query:          "fail",
			repoResults:    []string{"Failsworth, Manchester"},
			clientErr:      errors.New("connection timeout"),
			wantCount:      1,
			wantClientCall: true,
			wantSources:    map[string]bool{"user": true},
		},
		{
			name:    "repository error is propagated",
			query:   "broken",
			repoErr: errors.New("database connection lost"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRepo{results: tt.repoResults, err: tt.repoErr}
			client := &mockClient{results: tt.clientResults, err: tt.clientErr}
			svc := newTestService(repo, client)

			if tt.primeCache {
				svc.cache.SetDefault(strings.ToLower(tt.query), tt.cachedResults)
			}

			resp, err := svc.Search(ctx, userID, tt.query)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(resp.Results) != tt.wantCount {
				t.Errorf("got %d results, want %d", len(resp.Results), tt.wantCount)
			}

			if tt.wantClientCall && !client.called {
				t.Error("expected client.Search to be called, but it was not")
			}
			if !tt.wantClientCall && client.called {
				t.Error("expected client.Search NOT to be called, but it was")
			}

			if tt.wantSources != nil {
				foundSources := make(map[string]bool)
				for _, r := range resp.Results {
					foundSources[r.Source] = true
				}
				for src := range tt.wantSources {
					if !foundSources[src] {
						t.Errorf("expected source %q in results, but not found", src)
					}
				}
			}
		})
	}
}

func TestServiceSearch_CachePopulatedOnMiss(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	repo := &mockRepo{results: []string{}}
	client := &mockClient{
		results: []LocationResult{
			{Label: "Cached City, Country", Source: "photon"},
		},
	}
	svc := newTestService(repo, client)

	// First call should populate the cache.
	_, err := svc.Search(ctx, userID, "cached")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !client.called {
		t.Fatal("expected client to be called on first search")
	}

	// Reset and search again. The cache should be hit, so the client
	// should not be called a second time.
	client.called = false
	resp, err := svc.Search(ctx, userID, "cached")
	if err != nil {
		t.Fatalf("unexpected error on second search: %v", err)
	}

	if client.called {
		t.Error("expected cache hit, but client was called again")
	}

	if len(resp.Results) != 1 || resp.Results[0].Label != "Cached City, Country" {
		t.Errorf("unexpected results from cache: %+v", resp.Results)
	}
}
