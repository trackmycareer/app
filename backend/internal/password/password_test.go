package password

import (
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestHash(t *testing.T) {
	t.Run("produces argon2id hash", func(t *testing.T) {
		hash, err := Hash("testpassword")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.HasPrefix(hash, "$argon2id$") {
			t.Errorf("expected argon2id prefix, got: %s", hash)
		}
	})

	t.Run("produces different hashes for same password", func(t *testing.T) {
		hash1, err := Hash("testpassword")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		hash2, err := Hash("testpassword")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if hash1 == hash2 {
			t.Error("expected different hashes due to random salt")
		}
	})

	t.Run("hash has correct format", func(t *testing.T) {
		hash, err := Hash("testpassword")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		parts := strings.Split(hash, "$")
		if len(parts) != 6 {
			t.Errorf("expected 6 parts, got %d: %s", len(parts), hash)
		}
		if parts[1] != "argon2id" {
			t.Errorf("expected argon2id algorithm, got: %s", parts[1])
		}
	})
}

func TestVerify(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(t *testing.T) string
		password  string
		wantMatch bool
		wantErr   bool
	}{
		{
			name: "argon2id correct password",
			setup: func(t *testing.T) string {
				t.Helper()
				hash, err := Hash("correctpassword")
				if err != nil {
					t.Fatalf("hashing: %v", err)
				}
				return hash
			},
			password:  "correctpassword",
			wantMatch: true,
			wantErr:   false,
		},
		{
			name: "argon2id wrong password",
			setup: func(t *testing.T) string {
				t.Helper()
				hash, err := Hash("correctpassword")
				if err != nil {
					t.Fatalf("hashing: %v", err)
				}
				return hash
			},
			password:  "wrongpassword",
			wantMatch: false,
			wantErr:   false,
		},
		{
			name: "bcrypt correct password",
			setup: func(t *testing.T) string {
				t.Helper()
				hash, err := bcrypt.GenerateFromPassword([]byte("bcryptpassword"), bcrypt.DefaultCost)
				if err != nil {
					t.Fatalf("bcrypt hashing: %v", err)
				}
				return string(hash)
			},
			password:  "bcryptpassword",
			wantMatch: true,
			wantErr:   false,
		},
		{
			name: "bcrypt wrong password",
			setup: func(t *testing.T) string {
				t.Helper()
				hash, err := bcrypt.GenerateFromPassword([]byte("bcryptpassword"), bcrypt.DefaultCost)
				if err != nil {
					t.Fatalf("bcrypt hashing: %v", err)
				}
				return string(hash)
			},
			password:  "wrongpassword",
			wantMatch: false,
			wantErr:   false,
		},
		{
			name: "invalid argon2id hash format",
			setup: func(t *testing.T) string {
				t.Helper()
				return "$argon2id$invalid"
			},
			password:  "anything",
			wantMatch: false,
			wantErr:   true,
		},
		{
			name: "empty password against argon2id hash",
			setup: func(t *testing.T) string {
				t.Helper()
				hash, err := Hash("notempty")
				if err != nil {
					t.Fatalf("hashing: %v", err)
				}
				return hash
			},
			password:  "",
			wantMatch: false,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash := tt.setup(t)
			match, err := Verify(tt.password, hash)
			if (err != nil) != tt.wantErr {
				t.Errorf("Verify() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if match != tt.wantMatch {
				t.Errorf("Verify() = %v, want %v", match, tt.wantMatch)
			}
		})
	}
}

func TestNeedsRehash(t *testing.T) {
	tests := []struct {
		name string
		hash string
		want bool
	}{
		{
			name: "bcrypt hash needs rehash",
			hash: "$2a$10$abcdefghijklmnopqrstuuABCDEFGHIJKLMNOPQRSTUVWXYZ012",
			want: true,
		},
		{
			name: "argon2id hash does not need rehash",
			hash: "$argon2id$v=19$m=65536,t=3,p=4$c2FsdHNhbHQ$a2V5a2V5",
			want: false,
		},
		{
			name: "empty hash needs rehash",
			hash: "",
			want: true,
		},
		{
			name: "random string needs rehash",
			hash: "notahash",
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NeedsRehash(tt.hash); got != tt.want {
				t.Errorf("NeedsRehash() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRoundTrip(t *testing.T) {
	passwords := []string{
		"simple",
		"c0mpl3x!P@ssw0rd#2024",
		"unicode-пароль-密码",
		strings.Repeat("a", 128),
	}

	for _, pw := range passwords {
		t.Run(pw[:min(len(pw), 20)], func(t *testing.T) {
			hash, err := Hash(pw)
			if err != nil {
				t.Fatalf("Hash() error: %v", err)
			}

			match, err := Verify(pw, hash)
			if err != nil {
				t.Fatalf("Verify() error: %v", err)
			}
			if !match {
				t.Error("expected password to match its own hash")
			}

			match, err = Verify(pw+"wrong", hash)
			if err != nil {
				t.Fatalf("Verify() error: %v", err)
			}
			if match {
				t.Error("expected modified password not to match")
			}
		})
	}
}
