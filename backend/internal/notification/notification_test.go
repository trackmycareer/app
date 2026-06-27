package notification

import (
	"strings"
	"testing"
	"time"

	"github.com/trackmycareer/app/pkg/types"
)

func TestNextRunAt(t *testing.T) {
	const hour = 8

	tests := []struct {
		name string
		now  time.Time
		want time.Time
	}{
		{
			name: "before target hour same day",
			now:  time.Date(2026, 6, 27, 6, 0, 0, 0, time.UTC),
			want: time.Date(2026, 6, 27, 8, 0, 0, 0, time.UTC),
		},
		{
			name: "exactly target hour rolls to next day",
			now:  time.Date(2026, 6, 27, 8, 0, 0, 0, time.UTC),
			want: time.Date(2026, 6, 28, 8, 0, 0, 0, time.UTC),
		},
		{
			name: "after target hour rolls to next day",
			now:  time.Date(2026, 6, 27, 9, 30, 0, 0, time.UTC),
			want: time.Date(2026, 6, 28, 8, 0, 0, 0, time.UTC),
		},
		{
			name: "one second before target hour",
			now:  time.Date(2026, 6, 27, 7, 59, 59, 0, time.UTC),
			want: time.Date(2026, 6, 27, 8, 0, 0, 0, time.UTC),
		},
		{
			name: "month boundary",
			now:  time.Date(2026, 6, 30, 23, 30, 0, 0, time.UTC),
			want: time.Date(2026, 7, 1, 8, 0, 0, 0, time.UTC),
		},
		{
			name: "year boundary",
			now:  time.Date(2026, 12, 31, 23, 30, 0, 0, time.UTC),
			want: time.Date(2027, 1, 1, 8, 0, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := nextRunAt(tt.now, hour)
			if !got.Equal(tt.want) {
				t.Errorf("nextRunAt(%v) = %v, want %v", tt.now, got, tt.want)
			}
			if !got.After(tt.now) {
				t.Errorf("nextRunAt(%v) = %v is not strictly after now", tt.now, got)
			}
		})
	}
}

func TestApplicableType(t *testing.T) {
	tests := []struct {
		name                       string
		days                       int
		remind90, remind30, remind7 bool
		want                       string
	}{
		{"outside window above 90", 91, true, true, true, ""},
		{"at 90 days", 90, true, true, true, TypeExpiry90},
		{"between 30 and 90", 31, true, true, true, TypeExpiry90},
		{"at 30 days", 30, true, true, true, TypeExpiry30},
		{"between 7 and 30", 8, true, true, true, TypeExpiry30},
		{"at 7 days", 7, true, true, true, TypeExpiry7},
		{"at 1 day", 1, true, true, true, TypeExpiry7},
		{"expires today", 0, true, true, true, TypeExpiry7},
		{"one day expired", -1, true, true, true, TypeExpired},
		{"thirty days expired", -30, true, true, true, TypeExpired},
		{"7-band disabled falls through to 30", 7, true, true, false, TypeExpiry30},
		{"30-band disabled falls through to 90", 30, true, false, true, TypeExpiry90},
		{"90-band disabled yields none", 90, false, true, true, ""},
		{"all bands disabled within window yields none", 5, false, false, false, ""},
		{"expired ignores disabled bands", -3, false, false, false, TypeExpired},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := applicableType(tt.days, tt.remind90, tt.remind30, tt.remind7)
			if got != tt.want {
				t.Errorf("applicableType(%d, %t, %t, %t) = %q, want %q",
					tt.days, tt.remind90, tt.remind30, tt.remind7, got, tt.want)
			}
		})
	}
}

func TestComposeReminder(t *testing.T) {
	expiry := types.NewDate(time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC))

	tests := []struct {
		name           string
		due            DueReminder
		wantTitle      string
		bodyContains   []string
		bodyExcludes   []string
	}{
		{
			name: "plural days",
			due: DueReminder{
				CertName: "AWS Solutions Architect", Provider: "Amazon",
				ExpiryDate: expiry, DaysRemaining: 30, ApplicableType: TypeExpiry30,
			},
			wantTitle:    "AWS Solutions Architect expires in 30 days",
			bodyContains: []string{"from Amazon", "1 September 2026", "30 days"},
		},
		{
			name: "singular day",
			due: DueReminder{
				CertName: "CKA", Provider: "CNCF",
				ExpiryDate: expiry, DaysRemaining: 1, ApplicableType: TypeExpiry7,
			},
			wantTitle:    "CKA expires in 1 day",
			bodyContains: []string{"1 day from now"},
			bodyExcludes: []string{"1 days"},
		},
		{
			name: "expired",
			due: DueReminder{
				CertName: "CKA", Provider: "CNCF",
				ExpiryDate: expiry, DaysRemaining: -2, ApplicableType: TypeExpired,
			},
			wantTitle:    "CKA has expired",
			bodyContains: []string{"expired on 1 September 2026"},
		},
		{
			name: "no provider omits from clause",
			due: DueReminder{
				CertName: "Self-study", Provider: "",
				ExpiryDate: expiry, DaysRemaining: 7, ApplicableType: TypeExpiry7,
			},
			wantTitle:    "Self-study expires in 7 days",
			bodyContains: []string{"Self-study certification expires"},
			bodyExcludes: []string{"certification from "},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			title, body := composeReminder(tt.due)
			if title != tt.wantTitle {
				t.Errorf("title = %q, want %q", title, tt.wantTitle)
			}
			for _, sub := range tt.bodyContains {
				if !strings.Contains(body, sub) {
					t.Errorf("body %q does not contain %q", body, sub)
				}
			}
			for _, sub := range tt.bodyExcludes {
				if strings.Contains(body, sub) {
					t.Errorf("body %q should not contain %q", body, sub)
				}
			}
		})
	}
}
