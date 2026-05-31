package linkedaccount

import "testing"

func TestExtractDomain(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"full https URL", "https://sarahchen.dev", "sarahchen.dev", false},
		{"with path", "https://sarahchen.dev/blog", "sarahchen.dev", false},
		{"with port", "https://sarahchen.dev:8080", "sarahchen.dev", false},
		{"no scheme", "sarahchen.dev", "sarahchen.dev", false},
		{"http scheme", "http://example.com", "example.com", false},
		{"subdomain", "https://blog.sarahchen.dev", "blog.sarahchen.dev", false},
		{"empty string", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExtractDomain(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ExtractDomain(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("ExtractDomain(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
