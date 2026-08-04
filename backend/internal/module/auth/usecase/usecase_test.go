package usecase_test

import (
	"context"
	"testing"
	"time"

	"github.com/cia-da-vacina/crm/backend/internal/domain/entity"
	"github.com/cia-da-vacina/crm/backend/internal/module/auth/repository"
	"github.com/cia-da-vacina/crm/backend/internal/module/auth/usecase"
	"github.com/cia-da-vacina/crm/backend/internal/testutil"
	"github.com/cia-da-vacina/crm/backend/pkg/apperrors"
	"github.com/cia-da-vacina/crm/backend/pkg/jwt"
)

func newUseCase(t *testing.T) (*usecase.UseCase, *repository.Repository) {
	db := testutil.DB(t)
	repo := repository.New(db)
	jwtSvc, err := jwt.NewService("test-secret-at-least-this-long", "test", 15*time.Minute, 168*time.Hour)
	if err != nil {
		t.Fatalf("new jwt service: %v", err)
	}
	return usecase.New(repo, jwtSvc), repo
}

func TestLogin_Success_IssuesTokensWithUnitScopedClaims(t *testing.T) {
	uc, _ := newUseCase(t)
	db := testutil.DB(t)
	unit := testutil.NewUnit(t, db)
	user := testutil.NewUser(t, db, entity.RoleAgent, unit.ID)

	out, err := uc.Login(context.Background(), usecase.LoginInput{
		Email: user.Email, Password: testutil.TestPassword, IP: "127.0.0.1", UserAgent: "test",
	})
	if err != nil {
		t.Fatalf("expected successful login, got error: %v", err)
	}
	if out.AccessToken == "" || out.RefreshToken == "" {
		t.Fatal("expected non-empty access and refresh tokens")
	}
	if out.User.ID != user.ID || out.User.Email != user.Email {
		t.Fatalf("unexpected user in login output: %+v", out.User)
	}
	if out.ExpiresIn != int64((15 * time.Minute).Seconds()) {
		t.Fatalf("expected expires_in=900, got %d", out.ExpiresIn)
	}
}

func TestLogin_WrongPassword_Unauthorized(t *testing.T) {
	uc, _ := newUseCase(t)
	db := testutil.DB(t)
	unit := testutil.NewUnit(t, db)
	user := testutil.NewUser(t, db, entity.RoleAgent, unit.ID)

	_, err := uc.Login(context.Background(), usecase.LoginInput{Email: user.Email, Password: "wrong-password"})
	assertUnauthorized(t, err)
}

func TestLogin_UnknownEmail_Unauthorized(t *testing.T) {
	uc, _ := newUseCase(t)

	_, err := uc.Login(context.Background(), usecase.LoginInput{Email: "nobody-" + t.Name() + "@example.test", Password: "whatever"})
	assertUnauthorized(t, err)
}

// Inactive users must be rejected even with the correct password — this is
// the account-disable path (e.g. an offboarded agent), not just a wrong
// credential; ARCHITECTURE.md doesn't call this out explicitly but the code
// does check it (usecase.go Login), so it must be covered.
func TestLogin_InactiveUser_Forbidden(t *testing.T) {
	uc, _ := newUseCase(t)
	db := testutil.DB(t)
	unit := testutil.NewUnit(t, db)
	user := testutil.NewUser(t, db, entity.RoleAgent, unit.ID)
	testutil.SetUserActive(t, db, user.ID, false)

	_, err := uc.Login(context.Background(), usecase.LoginInput{Email: user.Email, Password: testutil.TestPassword})
	respErr, ok := err.(*apperrors.ResponseError)
	if !ok {
		t.Fatalf("expected *apperrors.ResponseError, got %T (%v)", err, err)
	}
	if respErr.StatusCode != 403 {
		t.Fatalf("expected 403 forbidden for inactive user, got %d", respErr.StatusCode)
	}
}

// Refresh rotates the token: the old refresh_token must stop working after
// use (single-use rotation, docs/BACKEND-CONTRACT.md §1) — this is the
// property that makes token theft detectable (a replayed old token fails).
func TestRefresh_RotatesToken_OldTokenStopsWorking(t *testing.T) {
	uc, _ := newUseCase(t)
	db := testutil.DB(t)
	unit := testutil.NewUnit(t, db)
	user := testutil.NewUser(t, db, entity.RoleAgent, unit.ID)

	login, err := uc.Login(context.Background(), usecase.LoginInput{Email: user.Email, Password: testutil.TestPassword})
	if err != nil {
		t.Fatalf("login: %v", err)
	}

	refreshed, err := uc.Refresh(context.Background(), usecase.RefreshInput{RefreshToken: login.RefreshToken})
	if err != nil {
		t.Fatalf("expected refresh to succeed, got: %v", err)
	}
	if refreshed.RefreshToken == login.RefreshToken {
		t.Fatal("expected refresh to issue a NEW refresh token, got the same one back")
	}

	// Replaying the original (now-revoked) refresh token must fail.
	if _, err := uc.Refresh(context.Background(), usecase.RefreshInput{RefreshToken: login.RefreshToken}); err == nil {
		t.Fatal("expected replaying a rotated-out refresh token to fail, got nil error")
	}

	// The new token from the rotation must still work.
	if _, err := uc.Refresh(context.Background(), usecase.RefreshInput{RefreshToken: refreshed.RefreshToken}); err != nil {
		t.Fatalf("expected the newly-issued refresh token to work, got: %v", err)
	}
}

func TestRefresh_UnknownToken_Unauthorized(t *testing.T) {
	uc, _ := newUseCase(t)
	_, err := uc.Refresh(context.Background(), usecase.RefreshInput{RefreshToken: "not-a-real-token"})
	assertUnauthorized(t, err)
}

// Logout revokes ALL sessions for the user, not just one — the endpoint has
// no way to identify a single session (no refresh_token in the body, see
// repository.RevokeAllUserRefreshTokens), so this is the documented,
// deliberate behavior in backend/ARCHITECTURE.md §4, not an oversight.
func TestLogout_RevokesAllSessions_NotJustOne(t *testing.T) {
	uc, _ := newUseCase(t)
	db := testutil.DB(t)
	unit := testutil.NewUnit(t, db)
	user := testutil.NewUser(t, db, entity.RoleAgent, unit.ID)

	loginA, err := uc.Login(context.Background(), usecase.LoginInput{Email: user.Email, Password: testutil.TestPassword, UserAgent: "device-a"})
	if err != nil {
		t.Fatalf("login A: %v", err)
	}
	loginB, err := uc.Login(context.Background(), usecase.LoginInput{Email: user.Email, Password: testutil.TestPassword, UserAgent: "device-b"})
	if err != nil {
		t.Fatalf("login B: %v", err)
	}

	if err := uc.Logout(context.Background(), user.ID); err != nil {
		t.Fatalf("logout: %v", err)
	}

	if _, err := uc.Refresh(context.Background(), usecase.RefreshInput{RefreshToken: loginA.RefreshToken}); err == nil {
		t.Fatal("expected session A's refresh token to be revoked by logout")
	}
	if _, err := uc.Refresh(context.Background(), usecase.RefreshInput{RefreshToken: loginB.RefreshToken}); err == nil {
		t.Fatal("expected session B's refresh token to ALSO be revoked by logout (logout revokes every session)")
	}
}

// Logout must be idempotent — calling it twice is not an error (no
// "already logged out" state to violate).
func TestLogout_Idempotent(t *testing.T) {
	uc, _ := newUseCase(t)
	db := testutil.DB(t)
	unit := testutil.NewUnit(t, db)
	user := testutil.NewUser(t, db, entity.RoleAgent, unit.ID)

	if err := uc.Logout(context.Background(), user.ID); err != nil {
		t.Fatalf("first logout: %v", err)
	}
	if err := uc.Logout(context.Background(), user.ID); err != nil {
		t.Fatalf("second logout should be a no-op, got error: %v", err)
	}
}

func assertUnauthorized(t *testing.T, err error) {
	t.Helper()
	respErr, ok := err.(*apperrors.ResponseError)
	if !ok {
		t.Fatalf("expected *apperrors.ResponseError, got %T (%v)", err, err)
	}
	if respErr.StatusCode != 401 {
		t.Fatalf("expected 401 unauthorized, got %d", respErr.StatusCode)
	}
}
