package crypto

import (
	"bytes"
	"crypto/rand"
	"testing"
)

func validKey(t *testing.T) []byte {
	t.Helper()
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		t.Fatalf("generating key: %v", err)
	}
	return key
}

func TestNewEncryptor(t *testing.T) {
	tests := []struct {
		name    string
		keyLen  int
		wantErr bool
	}{
		{"valid 32-byte key", 32, false},
		{"too short", 16, true},
		{"too long", 64, true},
		{"empty", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := make([]byte, tt.keyLen)
			_, err := NewEncryptor(key)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewEncryptor() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEncryptDecryptRoundtrip(t *testing.T) {
	tests := []struct {
		name      string
		plaintext []byte
	}{
		{"short string", []byte("hello")},
		{"empty", []byte{}},
		{"binary data", []byte{0x00, 0xFF, 0xAB, 0x12}},
		{"TOTP secret", []byte("JBSWY3DPEHPK3PXP")},
		{"long string", bytes.Repeat([]byte("a"), 1024)},
	}

	enc, err := NewEncryptor(validKey(t))
	if err != nil {
		t.Fatalf("NewEncryptor: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ciphertext, nonce, err := enc.Encrypt(tt.plaintext)
			if err != nil {
				t.Fatalf("Encrypt: %v", err)
			}

			got, err := enc.Decrypt(ciphertext, nonce)
			if err != nil {
				t.Fatalf("Decrypt: %v", err)
			}

			if !bytes.Equal(got, tt.plaintext) {
				t.Errorf("roundtrip mismatch: got %q, want %q", got, tt.plaintext)
			}
		})
	}
}

func TestDecryptWithWrongKey(t *testing.T) {
	enc1, _ := NewEncryptor(validKey(t))
	enc2, _ := NewEncryptor(validKey(t))

	ciphertext, nonce, err := enc1.Encrypt([]byte("secret"))
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}

	_, err = enc2.Decrypt(ciphertext, nonce)
	if err == nil {
		t.Error("expected error decrypting with wrong key, got nil")
	}
}

func TestDecryptWithWrongNonce(t *testing.T) {
	enc, _ := NewEncryptor(validKey(t))

	ciphertext, _, err := enc.Encrypt([]byte("secret"))
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}

	wrongNonce := make([]byte, nonceLength)
	_, err = enc.Decrypt(ciphertext, wrongNonce)
	if err == nil {
		t.Error("expected error decrypting with wrong nonce, got nil")
	}
}

func TestDecryptWithInvalidNonceLength(t *testing.T) {
	enc, _ := NewEncryptor(validKey(t))

	_, err := enc.Decrypt([]byte("ciphertext"), []byte("short"))
	if err == nil {
		t.Error("expected error with invalid nonce length, got nil")
	}
}

func TestEncryptProducesUniqueOutput(t *testing.T) {
	enc, _ := NewEncryptor(validKey(t))
	plaintext := []byte("same input")

	ct1, n1, _ := enc.Encrypt(plaintext)
	ct2, n2, _ := enc.Encrypt(plaintext)

	if bytes.Equal(n1, n2) {
		t.Error("expected unique nonces, got identical")
	}
	if bytes.Equal(ct1, ct2) {
		t.Error("expected unique ciphertexts, got identical")
	}
}

func TestDecryptTamperedCiphertext(t *testing.T) {
	enc, _ := NewEncryptor(validKey(t))

	ciphertext, nonce, _ := enc.Encrypt([]byte("secret"))

	ciphertext[0] ^= 0xFF

	_, err := enc.Decrypt(ciphertext, nonce)
	if err == nil {
		t.Error("expected error decrypting tampered ciphertext, got nil")
	}
}
