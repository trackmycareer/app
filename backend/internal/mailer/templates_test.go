package mailer

import "testing"

func TestCertReminderSubject(t *testing.T) {
	tests := []struct {
		name string
		data CertReminderData
		want string
	}{
		{
			name: "plural days",
			data: CertReminderData{CertName: "AWS SAA", DaysRemaining: 30},
			want: "Your AWS SAA certification expires in 30 days",
		},
		{
			name: "singular day",
			data: CertReminderData{CertName: "AWS SAA", DaysRemaining: 1},
			want: "Your AWS SAA certification expires in 1 day",
		},
		{
			name: "expired",
			data: CertReminderData{CertName: "AWS SAA", DaysRemaining: -2, Expired: true},
			want: "Your AWS SAA certification has expired",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := certReminderSubject(tt.data); got != tt.want {
				t.Errorf("certReminderSubject() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestCertReminderTemplatesRender ensures both reminder templates render for the
// expiring and expired cases without error and include the key fields.
func TestCertReminderTemplatesRender(t *testing.T) {
	cred := "https://example.com/renew"
	cases := []certReminderTemplateData{
		{Name: "Sam", CertName: "CKA", Provider: "CNCF", ExpiryDate: "1 September 2026", DaysRemaining: 7, Link: "https://app/certifications/1"},
		{Name: "Sam", CertName: "CKA", Provider: "CNCF", ExpiryDate: "1 September 2026", Expired: true, Link: "https://app/certifications/1", CredentialURL: &cred},
	}

	for _, td := range cases {
		if _, err := renderTemplate(certReminderHTMLTmpl, td); err != nil {
			t.Errorf("rendering HTML (expired=%t): %v", td.Expired, err)
		}
		text, err := renderTemplate(certReminderTextTmpl, td)
		if err != nil {
			t.Errorf("rendering text (expired=%t): %v", td.Expired, err)
		}
		if !contains(text, td.CertName) {
			t.Errorf("text body missing cert name for expired=%t", td.Expired)
		}
	}
}

func TestSanitiseHeader(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"plain", "AWS Solutions Architect", "AWS Solutions Architect"},
		{"strips CRLF injection", "AWS\r\nBcc: attacker@evil.test", "AWSBcc: attacker@evil.test"},
		{"strips lone LF", "line1\nline2", "line1line2"},
		{"strips lone CR", "line1\rline2", "line1line2"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sanitiseHeader(tt.in); got != tt.want {
				t.Errorf("sanitiseHeader(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func contains(haystack, needle string) bool {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}
