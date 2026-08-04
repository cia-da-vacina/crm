package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	authmodel "github.com/cia-da-vacina/crm/backend/internal/module/auth/model"
	"github.com/cia-da-vacina/crm/backend/internal/module/middleware"
	"github.com/cia-da-vacina/crm/backend/pkg/jwt"
)

// This is the exact chain GET /audit-logs is registered behind
// (internal/module/auditlog/module.go: RequireAuth then
// RequireRole(admin)) — the only thing standing between "any authenticated
// user" and "audit trail of every sensitive action in the system" is this
// middleware working correctly, so it's worth pinning down precisely rather
// than trusting the wiring.

func testJWT(t *testing.T) *jwt.Service {
	svc, err := jwt.NewService("test-secret-at-least-this-long", "test", 15*time.Minute, 168*time.Hour)
	if err != nil {
		t.Fatalf("new jwt service: %v", err)
	}
	return svc
}

func tokenFor(t *testing.T, svc *jwt.Service, role string) string {
	t.Helper()
	reg, err := svc.NewRegisteredClaims("user-id", svc.Expiration())
	if err != nil {
		t.Fatalf("new registered claims: %v", err)
	}
	claims := authmodel.UserClaims{RegisteredClaims: reg, Role: role, UnitIDs: []string{}}
	signed, err := svc.Sign(claims)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	return signed
}

func okHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
}

func adminOnlyChain(t *testing.T) http.Handler {
	mw := middleware.New(testJWT(t))
	return mw.RequireAuth(middleware.RequireRole("admin")(okHandler()))
}

func TestAuditLogChain_NoToken_Unauthorized(t *testing.T) {
	handler := adminOnlyChain(t)
	req := httptest.NewRequest(http.MethodGet, "/audit-logs", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without a token, got %d", rec.Code)
	}
}

func TestAuditLogChain_MalformedBearer_Unauthorized(t *testing.T) {
	handler := adminOnlyChain(t)
	req := httptest.NewRequest(http.MethodGet, "/audit-logs", nil)
	req.Header.Set("Authorization", "NotBearer sometoken")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for a malformed Authorization header, got %d", rec.Code)
	}
}

func TestAuditLogChain_ValidTokenWrongRole_Forbidden(t *testing.T) {
	svc := testJWT(t)
	mw := middleware.New(svc)
	handler := mw.RequireAuth(middleware.RequireRole("admin")(okHandler()))

	for _, role := range []string{"agent", "supervisor", "manager"} {
		t.Run(role, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/audit-logs", nil)
			req.Header.Set("Authorization", "Bearer "+tokenFor(t, svc, role))
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusForbidden {
				t.Fatalf("expected 403 for role=%s, got %d", role, rec.Code)
			}
		})
	}
}

func TestAuditLogChain_ValidAdminToken_Allowed(t *testing.T) {
	svc := testJWT(t)
	mw := middleware.New(svc)
	handler := mw.RequireAuth(middleware.RequireRole("admin")(okHandler()))

	req := httptest.NewRequest(http.MethodGet, "/audit-logs", nil)
	req.Header.Set("Authorization", "Bearer "+tokenFor(t, svc, "admin"))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for an admin token, got %d", rec.Code)
	}
}

// A token signed with a different secret (e.g. forged, or issued by a
// misconfigured second instance) must be rejected outright.
func TestRequireAuth_TokenSignedWithWrongSecret_Unauthorized(t *testing.T) {
	other, err := jwt.NewService("a-completely-different-secret", "test", 15*time.Minute, 168*time.Hour)
	if err != nil {
		t.Fatalf("new other jwt service: %v", err)
	}
	forged := tokenFor(t, other, "admin")

	handler := adminOnlyChain(t)
	req := httptest.NewRequest(http.MethodGet, "/audit-logs", nil)
	req.Header.Set("Authorization", "Bearer "+forged)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for a token signed with the wrong secret, got %d", rec.Code)
	}
}
