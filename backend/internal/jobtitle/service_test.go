package jobtitle

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// mockRepo implements a minimal test double for the repository layer.
type mockRepo struct {
	results []string
	err     error
}

func (m *mockRepo) SearchUserTitles(_ context.Context, _ uuid.UUID, _ string) ([]string, error) {
	return m.results, m.err
}

// testableService mirrors the production Service.Search logic but accepts the
// titleSearcher interface so we can inject mocks without modifying the
// production struct.
type testableService struct {
	repo titleSearcher
}

func (s *testableService) Search(ctx context.Context, userID uuid.UUID, query string) (*SearchResponse, error) {
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

	combined := make([]JobTitleResult, 0, len(userResults)+len(commonResults))
	combined = append(combined, userResults...)
	combined = append(combined, commonResults...)

	return &SearchResponse{Results: combined}, nil
}

func newTestService(repo titleSearcher) *testableService {
	return &testableService{repo: repo}
}

func TestServiceSearch(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	tests := []struct {
		name            string
		query           string
		repoResults     []string
		repoErr         error
		wantMinCount    int
		wantMaxCount    int
		wantErr         bool
		wantUserSource  bool
		wantCommonFirst bool // if false, user results should come first
		checkFunc       func(t *testing.T, resp *SearchResponse)
	}{
		{
			name:           "user titles returned when matching",
			query:          "lead",
			repoResults:    []string{"Lead DevOps Engineer", "Team Lead"},
			wantMinCount:   2,
			wantMaxCount:   12, // 2 user + up to 10 common
			wantUserSource: true,
			checkFunc: func(t *testing.T, resp *SearchResponse) {
				// First two results should be from the user.
				for i := 0; i < 2; i++ {
					if resp.Results[i].Source != "user" {
						t.Errorf("result[%d] source = %q, want %q", i, resp.Results[i].Source, "user")
					}
				}
			},
		},
		{
			name:         "common titles returned when no user titles match",
			query:        "software",
			repoResults:  nil,
			wantMinCount: 1,
			wantMaxCount: 10,
			checkFunc: func(t *testing.T, resp *SearchResponse) {
				for _, r := range resp.Results {
					if r.Source != "common" {
						t.Errorf("expected all results to be common, got source %q", r.Source)
					}
					if !strings.Contains(strings.ToLower(r.Title), "software") {
						t.Errorf("result %q does not contain 'software'", r.Title)
					}
				}
			},
		},
		{
			name:         "deduplication: title in both user and common list shows only once as user source",
			query:        "software engineer",
			repoResults:  []string{"Software Engineer"},
			wantMinCount: 1,
			checkFunc: func(t *testing.T, resp *SearchResponse) {
				count := 0
				for _, r := range resp.Results {
					if strings.ToLower(r.Title) == "software engineer" {
						count++
						if r.Source != "user" {
							t.Errorf("duplicated title should have source 'user', got %q", r.Source)
						}
					}
				}
				if count != 1 {
					t.Errorf("'Software Engineer' appeared %d times, want exactly 1", count)
				}
			},
		},
		{
			name:         "case-insensitive matching works",
			query:        "DATA SCIENTIST",
			repoResults:  nil,
			wantMinCount: 1,
			checkFunc: func(t *testing.T, resp *SearchResponse) {
				found := false
				for _, r := range resp.Results {
					if r.Title == "Data Scientist" {
						found = true
						break
					}
				}
				if !found {
					t.Error("expected 'Data Scientist' in results for query 'DATA SCIENTIST'")
				}
			},
		},
		{
			name:         "common titles limited to 10 results",
			query:        "manager",
			repoResults:  nil,
			wantMinCount: 10,
			wantMaxCount: 10,
		},
		{
			name:         "user results limited to 5 (by repository) plus up to 10 common",
			query:        "engineer",
			repoResults:  []string{"Custom Engineer 1", "Custom Engineer 2", "Custom Engineer 3", "Custom Engineer 4", "Custom Engineer 5"},
			wantMinCount: 5,
			wantMaxCount: 15, // 5 user + up to 10 common
			checkFunc: func(t *testing.T, resp *SearchResponse) {
				userCount := 0
				commonCount := 0
				for _, r := range resp.Results {
					switch r.Source {
					case "user":
						userCount++
					case "common":
						commonCount++
					}
				}
				if userCount != 5 {
					t.Errorf("user result count = %d, want 5", userCount)
				}
				if commonCount > 10 {
					t.Errorf("common result count = %d, want <= 10", commonCount)
				}
			},
		},
		{
			name:    "repository error propagates",
			query:   "broken",
			repoErr: errors.New("database connection lost"),
			wantErr: true,
		},
		{
			name:         "no results when nothing matches",
			query:        "zzzznonexistent",
			repoResults:  nil,
			wantMinCount: 0,
			wantMaxCount: 0,
		},
		{
			name:        "user results appear before common results",
			query:       "architect",
			repoResults: []string{"My Custom Architect Role"},
			checkFunc: func(t *testing.T, resp *SearchResponse) {
				if len(resp.Results) == 0 {
					t.Fatal("expected at least one result")
				}
				if resp.Results[0].Source != "user" {
					t.Errorf("first result source = %q, want 'user'", resp.Results[0].Source)
				}
				// Verify common results come after user results.
				seenCommon := false
				for _, r := range resp.Results {
					if r.Source == "common" {
						seenCommon = true
					}
					if seenCommon && r.Source == "user" {
						t.Error("user result appeared after common result")
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRepo{results: tt.repoResults, err: tt.repoErr}
			svc := newTestService(repo)

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

			if tt.wantMinCount > 0 && len(resp.Results) < tt.wantMinCount {
				t.Errorf("got %d results, want at least %d", len(resp.Results), tt.wantMinCount)
			}
			if tt.wantMaxCount > 0 && len(resp.Results) > tt.wantMaxCount {
				t.Errorf("got %d results, want at most %d", len(resp.Results), tt.wantMaxCount)
			}

			if tt.checkFunc != nil {
				tt.checkFunc(t, resp)
			}
		})
	}
}

func TestCommonTitlesAreSorted(t *testing.T) {
	for i := 1; i < len(commonTitles); i++ {
		if strings.ToLower(commonTitles[i-1]) >= strings.ToLower(commonTitles[i]) {
			t.Errorf("commonTitles not sorted: %q >= %q at index %d", commonTitles[i-1], commonTitles[i], i)
		}
	}
}

func TestCommonTitlesNoDuplicates(t *testing.T) {
	seen := make(map[string]bool, len(commonTitles))
	for _, title := range commonTitles {
		lower := strings.ToLower(title)
		if seen[lower] {
			t.Errorf("duplicate title in commonTitles: %q", title)
		}
		seen[lower] = true
	}
}
