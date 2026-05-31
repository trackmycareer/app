package certification

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// mockCertSearcher implements a minimal test double for the certSearcher
// interface.
type mockCertSearcher struct {
	results []CertSearchResult
	err     error
}

func (m *mockCertSearcher) SearchUserCertNames(_ context.Context, _ uuid.UUID, _ string) ([]CertSearchResult, error) {
	return m.results, m.err
}

func TestSearchCerts(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	tests := []struct {
		name         string
		query        string
		repoResults  []CertSearchResult
		repoErr      error
		wantMinCount int
		wantMaxCount int
		wantErr      bool
		checkFunc    func(t *testing.T, resp *CertSearchResponse)
	}{
		{
			name:  "user certs returned with correct name and provider",
			query: "aws",
			repoResults: []CertSearchResult{
				{Name: "AWS Solutions Architect", Provider: "Amazon", Source: "user"},
			},
			wantMinCount: 1,
			checkFunc: func(t *testing.T, resp *CertSearchResponse) {
				if resp.Results[0].Source != "user" {
					t.Errorf("first result source = %q, want %q", resp.Results[0].Source, "user")
				}
				if resp.Results[0].Name != "AWS Solutions Architect" {
					t.Errorf("first result name = %q, want %q", resp.Results[0].Name, "AWS Solutions Architect")
				}
				if resp.Results[0].Provider != "Amazon" {
					t.Errorf("first result provider = %q, want %q", resp.Results[0].Provider, "Amazon")
				}
			},
		},
		{
			name:         "common certs returned when matching",
			query:        "terraform",
			repoResults:  nil,
			wantMinCount: 1,
			checkFunc: func(t *testing.T, resp *CertSearchResponse) {
				for _, r := range resp.Results {
					if r.Source != "common" {
						t.Errorf("expected all results to be common, got source %q", r.Source)
					}
					if !strings.Contains(strings.ToLower(r.Name), "terraform") {
						t.Errorf("result %q does not contain 'terraform'", r.Name)
					}
					if r.Provider == "" {
						t.Errorf("result %q has empty provider", r.Name)
					}
				}
			},
		},
		{
			name:  "deduplication: cert in both user results and common list shows only once as user source",
			query: "certified kubernetes administrator",
			repoResults: []CertSearchResult{
				{Name: "Certified Kubernetes Administrator (CKA)", Provider: "My Provider", Source: "user"},
			},
			wantMinCount: 1,
			checkFunc: func(t *testing.T, resp *CertSearchResponse) {
				count := 0
				for _, r := range resp.Results {
					if strings.ToLower(r.Name) == "certified kubernetes administrator (cka)" {
						count++
						if r.Source != "user" {
							t.Errorf("duplicated cert should have source 'user', got %q", r.Source)
						}
					}
				}
				if count != 1 {
					t.Errorf("'Certified Kubernetes Administrator (CKA)' appeared %d times, want exactly 1", count)
				}
			},
		},
		{
			name:         "case-insensitive matching works",
			query:        "CISSP",
			repoResults:  nil,
			wantMinCount: 1,
			checkFunc: func(t *testing.T, resp *CertSearchResponse) {
				found := false
				for _, r := range resp.Results {
					if strings.Contains(r.Name, "CISSP") {
						found = true
						break
					}
				}
				if !found {
					t.Error("expected a CISSP result for query 'CISSP'")
				}
			},
		},
		{
			name:         "common results limited to 10",
			query:        "certified",
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
			name:  "provider is included in all results",
			query: "aws",
			repoResults: []CertSearchResult{
				{Name: "AWS Custom Cert", Provider: "My Company", Source: "user"},
			},
			wantMinCount: 1,
			checkFunc: func(t *testing.T, resp *CertSearchResponse) {
				for _, r := range resp.Results {
					if r.Provider == "" {
						t.Errorf("result %q has empty provider", r.Name)
					}
				}
			},
		},
		{
			name:  "user results appear before common results",
			query: "security",
			repoResults: []CertSearchResult{
				{Name: "My Custom Security Cert", Provider: "Acme Corp", Source: "user"},
			},
			checkFunc: func(t *testing.T, resp *CertSearchResponse) {
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
			query: "cloud",
			repoResults: []CertSearchResult{
				{Name: "Cloud Cert 1", Provider: "P1", Source: "user"},
				{Name: "Cloud Cert 2", Provider: "P2", Source: "user"},
				{Name: "Cloud Cert 3", Provider: "P3", Source: "user"},
				{Name: "Cloud Cert 4", Provider: "P4", Source: "user"},
				{Name: "Cloud Cert 5", Provider: "P5", Source: "user"},
			},
			wantMinCount: 5,
			wantMaxCount: 15, // 5 user + up to 10 common
			checkFunc: func(t *testing.T, resp *CertSearchResponse) {
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
			repo := &mockCertSearcher{results: tt.repoResults, err: tt.repoErr}
			resp, err := searchCerts(ctx, repo, userID, tt.query)

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

func TestCommonCertsAreSorted(t *testing.T) {
	for i := 1; i < len(commonCerts); i++ {
		if strings.ToLower(commonCerts[i-1].Name) >= strings.ToLower(commonCerts[i].Name) {
			t.Errorf("commonCerts not sorted: %q >= %q at index %d",
				commonCerts[i-1].Name, commonCerts[i].Name, i)
		}
	}
}

func TestCommonCertsNoDuplicates(t *testing.T) {
	seen := make(map[string]bool, len(commonCerts))
	for _, cert := range commonCerts {
		lower := strings.ToLower(cert.Name)
		if seen[lower] {
			t.Errorf("duplicate certification in commonCerts: %q", cert.Name)
		}
		seen[lower] = true
	}
}

func TestCommonCertsAllHaveProviders(t *testing.T) {
	for i, cert := range commonCerts {
		if cert.Provider == "" {
			t.Errorf("commonCerts[%d] %q has empty provider", i, cert.Name)
		}
		if cert.Name == "" {
			t.Errorf("commonCerts[%d] has empty name", i)
		}
	}
}
