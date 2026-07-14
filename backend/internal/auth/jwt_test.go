package auth

import (
	"testing"
	"time"

	"github.com/cia-vacina/crm/backend/internal/domain"
	"github.com/google/uuid"
)

func TestIssueAndParseAccessToken(t *testing.T) {
	svc := NewTokenService("test-secret", time.Minute, time.Hour)
	user := domain.User{
		ID:    uuid.New(),
		Email: "admin@ciadavacina.com.br",
		Role:  domain.RoleAdmin,
	}
	pair, err := svc.Issue(user)
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	claims, err := svc.Parse(pair.AccessToken, "access")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if claims.UserID != user.ID || claims.Role != domain.RoleAdmin {
		t.Fatalf("unexpected claims: %+v", claims)
	}
	if _, err := svc.Parse(pair.AccessToken, "refresh"); err == nil {
		t.Fatal("expected type mismatch error")
	}
}
