package crypto

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
)

func HashBackupCode(key []byte, code string) string {
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(code))
	return hex.EncodeToString(mac.Sum(nil))
}

func VerifyBackupCode(key []byte, code, expectedHash string) bool {
	actualHash := HashBackupCode(key, code)
	return subtle.ConstantTimeCompare([]byte(actualHash), []byte(expectedHash)) == 1
}
