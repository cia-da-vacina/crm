// Package vo contém tipos de valor de domínio — hoje só hashing de senha.
package vo

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/cia-da-vacina/crm/backend/pkg/env"
	"golang.org/x/crypto/argon2"
)

type PasswordConfig struct {
	Pepper      string
	MemoryCost  uint32
	TimeCost    uint32
	ThreadCount uint8
	KeyLength   uint32
	SaltLength  uint32
}

func DefaultPasswordConfig() *PasswordConfig {
	return &PasswordConfig{
		Pepper:      env.GetOrDefault("PASSWORD_HASH_PEPPER", "cia-da-vacina-2026"),
		MemoryCost:  uint32(env.GetOrDefault("PASSWORD_HASH_MEMORY_COST", 65536)),
		TimeCost:    uint32(env.GetOrDefault("PASSWORD_HASH_TIME_COST", 4)),
		ThreadCount: uint8(env.GetOrDefault("PASSWORD_HASH_THREADS", 4)),
		KeyLength:   32,
		SaltLength:  16,
	}
}

func HashPassword(password string, cfg *PasswordConfig) (string, error) {
	salt, err := generateSalt(cfg.SaltLength)
	if err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	hash := argon2.IDKey(
		[]byte(password+cfg.Pepper),
		salt,
		cfg.TimeCost,
		cfg.MemoryCost,
		cfg.ThreadCount,
		cfg.KeyLength,
	)

	encodedSalt := base64.RawStdEncoding.EncodeToString(salt)
	encodedHash := base64.RawStdEncoding.EncodeToString(hash)

	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, cfg.MemoryCost, cfg.TimeCost, cfg.ThreadCount, encodedSalt, encodedHash), nil
}

func VerifyPassword(password, hash string, cfg *PasswordConfig) bool {
	parts := strings.Split(hash, "$")
	if len(parts) != 6 {
		return false
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false
	}

	storedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false
	}

	computedHash := argon2.IDKey(
		[]byte(password+cfg.Pepper),
		salt,
		cfg.TimeCost,
		cfg.MemoryCost,
		cfg.ThreadCount,
		cfg.KeyLength,
	)

	return subtle.ConstantTimeCompare(storedHash, computedHash) == 1
}

func generateSalt(length uint32) ([]byte, error) {
	salt := make([]byte, length)
	_, err := rand.Read(salt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate random salt: %w", err)
	}
	return salt, nil
}
