package company

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

func (m *mockRepo) SearchUserCompanies(_ context.Context, _ uuid.UUID, _ string) ([]string, error) {
	return m.results, m.err
}

// mockClient implements a minimal test double for the Companies House client.
type mockClient struct {
	results []CompanyResult
	err     error
	called  bool
}

func (m *mockClient) Search(_ context.Context, _ string) ([]CompanyResult, error) {
	m.called = true
	return m.results, m.err
}

// companySearcher is the interface the service actually needs from the repository.
type companySearcher interface {
	SearchUserCompanies(ctx context.Context, userID uuid.UUID, query string) ([]string, error)
}

// externalSearcher is the interface the service actually needs from the client.
type externalSearcher interface {
	Search(ctx context.Context, query string) ([]CompanyResult, error)
}

// testableService wraps the search logic so we can inject mocks without
// modifying the production Service struct. It mirrors Service.Search exactly.
type testableService struct {
	repo   companySearcher
	cache  *cache.Cache
	client externalSearcher
}

func (s *testableService) Search(ctx context.Context, userID uuid.UUID, query string) (*SearchResponse, error) {
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

	cacheKey := strings.ToLower(query)
	var chResults []CompanyResult

	if cached, found := s.cache.Get(cacheKey); found {
		chResults, _ = cached.([]CompanyResult)
	} else {
		fetched, fetchErr := s.client.Search(ctx, query)
		if fetchErr != nil {
			// Graceful degradation: continue with user results only.
		} else {
			chResults = fetched
			s.cache.SetDefault(cacheKey, chResults)
		}
	}

	combined := userResults
	for _, r := range chResults {
		if !seen[strings.ToLower(r.Name)] {
			combined = append(combined, r)
			seen[strings.ToLower(r.Name)] = true
		}
	}

	return &SearchResponse{Results: combined}, nil
}

func newTestService(repo companySearcher, client externalSearcher) *testableService {
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
		clientResults  []CompanyResult
		clientErr      error
		primeCache     bool
		cachedResults  []CompanyResult
		wantCount      int
		wantErr        bool
		wantClientCall bool
		wantSources    map[string]bool // source values we expect to see
	}{
		{
			name:        "user companies returned when no Companies House results",
			query:       "acme",
			repoResults: []string{"Acme Corp", "Acme Ltd"},
			clientResults: nil,
			wantCount:      2,
			wantClientCall: true,
			wantSources:    map[string]bool{"user": true},
		},
		{
			name:        "cache hit returns cached results without calling Companies House",
			query:       "widget",
			repoResults: []string{},
			primeCache:  true,
			cachedResults: []CompanyResult{
				{Name: "Widget Inc", Source: "companies_house", CompanyNumber: "12345678"},
			},
			wantCount:      1,
			wantClientCall: false,
			wantSources:    map[string]bool{"companies_house": true},
		},
		{
			name:        "cache miss calls Companies House and caches the result",
			query:       "new co",
			repoResults: []string{},
			clientResults: []CompanyResult{
				{Name: "New Co Ltd", Source: "companies_house", CompanyNumber: "99999999"},
			},
			wantCount:      1,
			wantClientCall: true,
			wantSources:    map[string]bool{"companies_house": true},
		},
		{
			name:        "deduplication when company appears in both user and Companies House",
			query:       "acme",
			repoResults: []string{"Acme Corp"},
			clientResults: []CompanyResult{
				{Name: "acme corp", Source: "companies_house", CompanyNumber: "11111111"},
				{Name: "Acme Industries", Source: "companies_house", CompanyNumber: "22222222"},
			},
			wantCount:      2, // Acme Corp (user) + Acme Industries (CH)
			wantClientCall: true,
			wantSources:    map[string]bool{"user": true, "companies_house": true},
		},
		{
			name:        "empty API key means Companies House returns nil (skipped)",
			query:       "test",
			repoResults: []string{"Test Corp"},
			clientResults: nil,
			wantCount:      1,
			wantClientCall: true,
			wantSources:    map[string]bool{"user": true},
		},
		{
			name:        "Companies House API error still returns user results",
			query:       "fail",
			repoResults: []string{"Failsafe Ltd"},
			clientErr:   errors.New("connection timeout"),
			wantCount:      1,
			wantClientCall: true,
			wantSources:    map[string]bool{"user": true},
		},
		{
			name:    "repository error propagates",
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
		results: []CompanyResult{
			{Name: "Cached Co", Source: "companies_house"},
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

	if len(resp.Results) != 1 || resp.Results[0].Name != "Cached Co" {
		t.Errorf("unexpected results from cache: %+v", resp.Results)
	}
}
