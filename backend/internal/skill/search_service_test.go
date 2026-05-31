package skill

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// mockSkillSearcher implements a minimal test double for the skillSearcher
// interface.
type mockSkillSearcher struct {
	results []SkillSearchResult
	err     error
}

func (m *mockSkillSearcher) SearchUserSkillNames(_ context.Context, _ uuid.UUID, _ string) ([]SkillSearchResult, error) {
	return m.results, m.err
}

func TestSearchSkills(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	tests := []struct {
		name         string
		query        string
		repoResults  []SkillSearchResult
		repoErr      error
		wantMinCount int
		wantMaxCount int
		wantErr      bool
		checkFunc    func(t *testing.T, resp *SkillSearchResponse)
	}{
		{
			name:  "user skills returned with correct name and category",
			query: "go",
			repoResults: []SkillSearchResult{
				{Name: "Go", Category: "Programming Languages", Source: "user"},
			},
			wantMinCount: 1,
			checkFunc: func(t *testing.T, resp *SkillSearchResponse) {
				if resp.Results[0].Source != "user" {
					t.Errorf("first result source = %q, want %q", resp.Results[0].Source, "user")
				}
				if resp.Results[0].Name != "Go" {
					t.Errorf("first result name = %q, want %q", resp.Results[0].Name, "Go")
				}
				if resp.Results[0].Category != "Programming Languages" {
					t.Errorf("first result category = %q, want %q", resp.Results[0].Category, "Programming Languages")
				}
			},
		},
		{
			name:         "common skills returned when matching",
			query:        "terraform",
			repoResults:  nil,
			wantMinCount: 1,
			checkFunc: func(t *testing.T, resp *SkillSearchResponse) {
				for _, r := range resp.Results {
					if r.Source != "common" {
						t.Errorf("expected all results to be common, got source %q", r.Source)
					}
					if !strings.Contains(strings.ToLower(r.Name), "terraform") {
						t.Errorf("result %q does not contain 'terraform'", r.Name)
					}
					if r.Category == "" {
						t.Errorf("result %q has empty category", r.Name)
					}
				}
			},
		},
		{
			name:  "deduplication: skill in both user results and common list shows only once as user source",
			query: "react",
			repoResults: []SkillSearchResult{
				{Name: "React", Category: "Frameworks & Libraries", Source: "user"},
			},
			wantMinCount: 1,
			checkFunc: func(t *testing.T, resp *SkillSearchResponse) {
				count := 0
				for _, r := range resp.Results {
					if strings.ToLower(r.Name) == "react" {
						count++
						if r.Source != "user" {
							t.Errorf("duplicated skill should have source 'user', got %q", r.Source)
						}
					}
				}
				if count != 1 {
					t.Errorf("'React' appeared %d times, want exactly 1", count)
				}
			},
		},
		{
			name:         "case-insensitive matching works",
			query:        "PYTHON",
			repoResults:  nil,
			wantMinCount: 1,
			checkFunc: func(t *testing.T, resp *SkillSearchResponse) {
				found := false
				for _, r := range resp.Results {
					if strings.Contains(r.Name, "Python") {
						found = true
						break
					}
				}
				if !found {
					t.Error("expected a Python result for query 'PYTHON'")
				}
			},
		},
		{
			name:         "common results limited to 10",
			query:        "a",
			repoResults:  nil,
			wantMinCount: 10,
			wantMaxCount: 10,
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
			name:  "category included in all results",
			query: "docker",
			repoResults: []SkillSearchResult{
				{Name: "Docker Compose", Category: "DevOps & CI/CD", Source: "user"},
			},
			wantMinCount: 1,
			checkFunc: func(t *testing.T, resp *SkillSearchResponse) {
				for _, r := range resp.Results {
					if r.Category == "" {
						t.Errorf("result %q has empty category", r.Name)
					}
				}
			},
		},
		{
			name:  "user results appear before common results",
			query: "java",
			repoResults: []SkillSearchResult{
				{Name: "My Custom Java Skill", Category: "Programming Languages", Source: "user"},
			},
			checkFunc: func(t *testing.T, resp *SkillSearchResponse) {
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
		{
			name:  "user results limited to 5 (by repository) plus up to 10 common",
			query: "script",
			repoResults: []SkillSearchResult{
				{Name: "Script Skill 1", Category: "C1", Source: "user"},
				{Name: "Script Skill 2", Category: "C2", Source: "user"},
				{Name: "Script Skill 3", Category: "C3", Source: "user"},
				{Name: "Script Skill 4", Category: "C4", Source: "user"},
				{Name: "Script Skill 5", Category: "C5", Source: "user"},
			},
			wantMinCount: 5,
			wantMaxCount: 15, // 5 user + up to 10 common
			checkFunc: func(t *testing.T, resp *SkillSearchResponse) {
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockSkillSearcher{results: tt.repoResults, err: tt.repoErr}
			resp, err := searchSkills(ctx, repo, userID, tt.query)

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

func TestCommonSkillsAreSorted(t *testing.T) {
	for i := 1; i < len(commonSkills); i++ {
		if strings.ToLower(commonSkills[i-1].Name) >= strings.ToLower(commonSkills[i].Name) {
			t.Errorf("commonSkills not sorted: %q >= %q at index %d",
				commonSkills[i-1].Name, commonSkills[i].Name, i)
		}
	}
}

func TestCommonSkillsNoDuplicates(t *testing.T) {
	seen := make(map[string]bool, len(commonSkills))
	for _, sk := range commonSkills {
		lower := strings.ToLower(sk.Name)
		if seen[lower] {
			t.Errorf("duplicate skill in commonSkills: %q", sk.Name)
		}
		seen[lower] = true
	}
}

func TestCommonSkillsAllHaveCategories(t *testing.T) {
	for i, sk := range commonSkills {
		if sk.Category == "" {
			t.Errorf("commonSkills[%d] %q has empty category", i, sk.Name)
		}
		if sk.Name == "" {
			t.Errorf("commonSkills[%d] has empty name", i)
		}
	}
}
