// Package crypto cifra segredos em repouso (tokens de canal Meta) com
// AES-256-GCM. Nunca usado pra senha de usuário (isso é argon2id em
// internal/domain/vo) — só pra dado que precisa ser recuperado em texto
// plano depois (aqui: pra assinar chamadas à API da Meta).
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/cia-da-vacina/crm/backend/pkg/env"
)

type Cipher struct {
	gcm cipher.AEAD
}

// NewCipherFromEnv lê APP_ENCRYPTION_KEY (hex de 32 bytes = 64 chars, AES-256).
func NewCipherFromEnv() (*Cipher, error) {
	hexKey := env.GetOrDefault("APP_ENCRYPTION_KEY", "")
	if hexKey == "" {
		return nil, errors.New("APP_ENCRYPTION_KEY is not set")
	}
	key, err := hex.DecodeString(hexKey)
	if err != nil {
		return nil, fmt.Errorf("APP_ENCRYPTION_KEY must be hex-encoded: %w", err)
	}
	if len(key) != 32 {
		return nil, fmt.Errorf("APP_ENCRYPTION_KEY must decode to 32 bytes (AES-256), got %d", len(key))
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return &Cipher{gcm: gcm}, nil
}

// Encrypt retorna nonce||ciphertext||tag num único slice — o nonce vai junto
// pra não precisar de uma coluna separada no banco.
func (c *Cipher) Encrypt(plaintext string) ([]byte, error) {
	nonce := make([]byte, c.gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}
	return c.gcm.Seal(nonce, nonce, []byte(plaintext), nil), nil
}

func (c *Cipher) Decrypt(data []byte) (string, error) {
	nonceSize := c.gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("ciphertext too short")
	}
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := c.gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("decrypt failed: %w", err)
	}
	return string(plaintext), nil
}
