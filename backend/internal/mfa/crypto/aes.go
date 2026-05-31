package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
)

const (
	keyLength   = 32
	nonceLength = 12
)

type Encryptor struct {
	gcm cipher.AEAD
}

func NewEncryptor(key []byte) (*Encryptor, error) {
	if len(key) != keyLength {
		return nil, fmt.Errorf("encryption key must be exactly %d bytes, got %d", keyLength, len(key))
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("creating AES cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("creating GCM: %w", err)
	}

	return &Encryptor{gcm: gcm}, nil
}

func (e *Encryptor) Encrypt(plaintext []byte) (ciphertext, nonce []byte, err error) {
	nonce = make([]byte, nonceLength)
	if _, err := rand.Read(nonce); err != nil {
		return nil, nil, fmt.Errorf("generating nonce: %w", err)
	}

	ciphertext = e.gcm.Seal(nil, nonce, plaintext, nil)
	return ciphertext, nonce, nil
}

func (e *Encryptor) Decrypt(ciphertext, nonce []byte) ([]byte, error) {
	if len(nonce) != nonceLength {
		return nil, fmt.Errorf("nonce must be exactly %d bytes, got %d", nonceLength, len(nonce))
	}

	plaintext, err := e.gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decrypting: %w", err)
	}

	return plaintext, nil
}
