// Package jwt assina e valida os access/refresh tokens do CRM.
//
// Usa HMAC-SHA256 (secret simétrico) em vez do par de chaves RSA que outros
// projetos da empresa usam: aqui é um monólito único e quem fala com o
// backend é sempre o BFF Next.js, que só repassa o Bearer sem validar
// assinatura — não há um segundo serviço que precise validar o token de
// forma independente com uma chave pública. Ver backend/ARCHITECTURE.md §5.
package jwt

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/cia-da-vacina/crm/backend/pkg/env"
	"github.com/golang-jwt/jwt/v5"
)

type Service struct {
	secret            []byte
	issuer            string
	expiration        time.Duration
	refreshExpiration time.Duration
}

func NewService(secret, issuer string, expiration, refreshExpiration time.Duration) (*Service, error) {
	if secret == "" {
		return nil, errors.New("jwt secret is required")
	}
	return &Service{
		secret:            []byte(secret),
		issuer:            issuer,
		expiration:        expiration,
		refreshExpiration: refreshExpiration,
	}, nil
}

func NewServiceFromEnv() (*Service, error) {
	secret := env.GetOrDefault("JWT_SECRET", "")
	issuer := env.GetOrDefault("API_DOMAIN", "localhost")

	expiration := parseDurationEnv("TOKEN_EXPIRATION", 15*time.Minute)
	refreshExpiration := parseDurationEnv("REFRESH_TOKEN_EXPIRATION", 168*time.Hour)

	return NewService(secret, issuer, expiration, refreshExpiration)
}

func (s *Service) Expiration() time.Duration        { return s.expiration }
func (s *Service) RefreshExpiration() time.Duration { return s.refreshExpiration }

// NewRegisteredClaims produz os claims padrão (iss/sub/iat/nbf/exp/jti).
// O payload final deve embutir o retorno, mantendo tudo na raiz do JWT.
func (s *Service) NewRegisteredClaims(subject string, expiration time.Duration) (jwt.RegisteredClaims, error) {
	raw := make([]byte, 16)
	if _, err := rand.Read(raw); err != nil {
		return jwt.RegisteredClaims{}, fmt.Errorf("failed to generate token ID: %w", err)
	}
	now := time.Now()
	return jwt.RegisteredClaims{
		Issuer:    s.issuer,
		Subject:   subject,
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(expiration)),
		NotBefore: jwt.NewNumericDate(now),
		ID:        hex.EncodeToString(raw),
	}, nil
}

// Sign assina claims que embutem jwt.RegisteredClaims.
func (s *Service) Sign(claims jwt.Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(s.secret)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}
	return signed, nil
}

// Validate parseia e valida o token, preenchendo dest (ponteiro para struct
// que embute jwt.RegisteredClaims).
func (s *Service) Validate(tokenString string, dest jwt.Claims) error {
	token, err := jwt.ParseWithClaims(tokenString, dest, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.secret, nil
	})
	if err != nil {
		return err
	}
	if !token.Valid {
		return errors.New("invalid token")
	}
	return nil
}

func parseDurationEnv(key string, fallback time.Duration) time.Duration {
	val := env.GetOrDefault(key, "")
	if val == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(val)
	if err != nil {
		return fallback
	}
	return parsed
}
