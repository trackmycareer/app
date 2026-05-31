// Package password provides Argon2id password hashing with transparent bcrypt
// backwards compatibility for gradual migration.
package password

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/bcrypt"
)

// Argon2id parameters (OWASP recommendations).
const (
	argon2Memory      = 64 * 1024 // 64 MB
	argon2Iterations  = 3
	argon2Parallelism = 4
	argon2KeyLength   = 32
	argon2SaltLength  = 16
)

// Hash creates an Argon2id hash of the password.
func Hash(password string) (string, error) {
	salt := make([]byte, argon2SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generating salt: %w", err)
	}

	key := argon2.IDKey([]byte(password), salt, argon2Iterations, argon2Memory, argon2Parallelism, argon2KeyLength)

	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		argon2Memory, argon2Iterations, argon2Parallelism,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(key),
	), nil
}

// Verify checks a password against a hash. Supports both Argon2id and bcrypt
// formats, enabling transparent migration from bcrypt to Argon2id.
func Verify(password, hash string) (bool, error) {
	if strings.HasPrefix(hash, "$argon2id$") {
		return verifyArgon2id(password, hash)
	}
	// Assume bcrypt for $2a$, $2b$, etc.
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return false, nil
	}
	return err == nil, err
}

// NeedsRehash returns true if the hash is not Argon2id (i.e., it is bcrypt
// and should be migrated on next successful login).
func NeedsRehash(hash string) bool {
	return !strings.HasPrefix(hash, "$argon2id$")
}

func verifyArgon2id(password, hash string) (bool, error) {
	// Parse: $argon2id$v=19$m=65536,t=3,p=4$<salt>$<key>
	parts := strings.Split(hash, "$")
	if len(parts) != 6 {
		return false, fmt.Errorf("invalid argon2id hash format")
	}

	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil {
		return false, fmt.Errorf("parsing version: %w", err)
	}

	var memory uint32
	var iterations uint32
	var parallelism uint8
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &iterations, &parallelism); err != nil {
		return false, fmt.Errorf("parsing parameters: %w", err)
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, fmt.Errorf("decoding salt: %w", err)
	}

	expectedKey, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, fmt.Errorf("decoding key: %w", err)
	}

	key := argon2.IDKey([]byte(password), salt, iterations, memory, parallelism, uint32(len(expectedKey)))
	return subtle.ConstantTimeCompare(key, expectedKey) == 1, nil
}
