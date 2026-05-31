package crypto

import (
	"crypto/rand"
	"testing"
)

func TestHashBackupCodeRoundtrip(t *testing.T) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		t.Fatalf("generating key: %v", err)
	}

	tests := []struct {
		name string
		code string
	}{
		{"alphanumeric", "abc12345"},
		{"short", "ab"},
		{"long", "abcdefghijklmnop"},
		{"with special chars", "a1!@#$%^"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash := HashBackupCode(key, tt.code)
			if !VerifyBackupCode(key, tt.code, hash) {
				t.Error("VerifyBackupCode returned false for correct code")
			}
		})
	}
}

func TestVerifyBackupCodeWrongCode(t *testing.T) {
	key := make([]byte, 32)
	rand.Read(key)

	hash := HashBackupCode(key, "correct")
	if VerifyBackupCode(key, "wrong", hash) {
		t.Error("VerifyBackupCode returned true for wrong code")
	}
}

func TestVerifyBackupCodeWrongKey(t *testing.T) {
	key1 := make([]byte, 32)
	key2 := make([]byte, 32)
	rand.Read(key1)
	rand.Read(key2)

	hash := HashBackupCode(key1, "mycode")
	if VerifyBackupCode(key2, "mycode", hash) {
		t.Error("VerifyBackupCode returned true for wrong key")
	}
}

func TestHashBackupCodeDeterministic(t *testing.T) {
	key := make([]byte, 32)
	rand.Read(key)

	h1 := HashBackupCode(key, "same")
	h2 := HashBackupCode(key, "same")
	if h1 != h2 {
		t.Errorf("same input produced different hashes: %q vs %q", h1, h2)
	}
}

func TestHashBackupCodeDifferentCodes(t *testing.T) {
	key := make([]byte, 32)
	rand.Read(key)

	h1 := HashBackupCode(key, "code1")
	h2 := HashBackupCode(key, "code2")
	if h1 == h2 {
		t.Error("different codes produced the same hash")
	}
}

func TestHashBackupCodeOutput(t *testing.T) {
	key := make([]byte, 32)
	rand.Read(key)

	hash := HashBackupCode(key, "test")
	if len(hash) != 64 {
		t.Errorf("expected 64-char hex hash, got %d chars", len(hash))
	}
}
