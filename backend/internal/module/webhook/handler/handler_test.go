package handler_test

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cia-da-vacina/crm/backend/internal/domain/entity"
	"github.com/cia-da-vacina/crm/backend/internal/module/webhook/handler"
	"github.com/go-chi/chi/v5"
)

const testAppSecret = "test-app-secret-for-hmac"

type fakeUseCase struct {
	called  bool
	channel entity.Channel
	body    []byte
}

func (f *fakeUseCase) IngestPayload(ctx context.Context, channel entity.Channel, rawBody []byte) error {
	f.called = true
	f.channel = channel
	f.body = rawBody
	return nil
}

func withChannelParam(r *http.Request, channel string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("channel", channel)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

func sign(secret string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

// Receive must reject a payload before it ever reaches IngestPayload when
// the HMAC signature doesn't check out — "payload não assinado corretamente
// nunca é persistido, sem exceção mesmo em dev" (handler.go comment,
// docs/BACKEND-CONTRACT.md §8/§9). This is the one place the entire webhook
// surface is authenticated (there's no Bearer here), so it's worth pinning
// down precisely: wrong secret, and a byte-for-byte tampered body under an
// otherwise-valid signature, must both be rejected with 403 and never touch
// the usecase.
func TestReceive_InvalidSignature_RejectedBefore403WithoutIngesting(t *testing.T) {
	t.Setenv("META_APP_SECRET", testAppSecret)

	body := []byte(`{"entry":[]}`)
	uc := &fakeUseCase{}
	h := handler.New(uc)

	req := httptest.NewRequest(http.MethodPost, "/webhooks/meta/whatsapp", strings.NewReader(string(body)))
	req.Header.Set("X-Hub-Signature-256", sign("wrong-secret", body))
	req = withChannelParam(req, "whatsapp")
	rec := httptest.NewRecorder()

	h.Receive(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for a signature computed with the wrong secret, got %d", rec.Code)
	}
	if uc.called {
		t.Fatal("expected IngestPayload to never be called when signature verification fails")
	}
}

func TestReceive_TamperedBodyUnderValidSignature_Rejected(t *testing.T) {
	t.Setenv("META_APP_SECRET", testAppSecret)

	original := []byte(`{"entry":[{"changes":[]}]}`)
	tampered := []byte(`{"entry":[{"changes":[],"injected":true}]}`)
	uc := &fakeUseCase{}
	h := handler.New(uc)

	req := httptest.NewRequest(http.MethodPost, "/webhooks/meta/whatsapp", strings.NewReader(string(tampered)))
	req.Header.Set("X-Hub-Signature-256", sign(testAppSecret, original)) // signed for the ORIGINAL body, not the one actually sent
	req = withChannelParam(req, "whatsapp")
	rec := httptest.NewRecorder()

	h.Receive(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for a body that doesn't match its signature, got %d", rec.Code)
	}
	if uc.called {
		t.Fatal("expected IngestPayload to never be called for a tampered body")
	}
}

func TestReceive_ValidSignature_IngestsAndReturns200(t *testing.T) {
	t.Setenv("META_APP_SECRET", testAppSecret)

	body := []byte(`{"entry":[{"changes":[]}]}`)
	uc := &fakeUseCase{}
	h := handler.New(uc)

	req := httptest.NewRequest(http.MethodPost, "/webhooks/meta/whatsapp", strings.NewReader(string(body)))
	req.Header.Set("X-Hub-Signature-256", sign(testAppSecret, body))
	req = withChannelParam(req, "whatsapp")
	rec := httptest.NewRecorder()

	h.Receive(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for a validly-signed payload, got %d", rec.Code)
	}
	if !uc.called {
		t.Fatal("expected IngestPayload to be called for a validly-signed payload")
	}
	if uc.channel != entity.ChannelWhatsApp {
		t.Fatalf("expected channel=whatsapp to be resolved from the URL param, got %q", uc.channel)
	}
}

// Even a genuine ingestion failure (e.g. malformed payload) must still
// answer 200 — the Meta retries webhooks that don't get a 200 back, and a
// message that's malformed once will be malformed forever, so retrying
// would just loop (see handler.go comment on Receive).
func TestReceive_IngestionErrorStillReturns200(t *testing.T) {
	t.Setenv("META_APP_SECRET", testAppSecret)

	body := []byte(`not-json-at-all`)
	uc := &erroringUseCase{}
	h := handler.New(uc)

	req := httptest.NewRequest(http.MethodPost, "/webhooks/meta/whatsapp", strings.NewReader(string(body)))
	req.Header.Set("X-Hub-Signature-256", sign(testAppSecret, body))
	req = withChannelParam(req, "whatsapp")
	rec := httptest.NewRecorder()

	h.Receive(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 even when ingestion fails internally, got %d", rec.Code)
	}
}

type erroringUseCase struct{}

func (erroringUseCase) IngestPayload(ctx context.Context, channel entity.Channel, rawBody []byte) error {
	return context.DeadlineExceeded
}

func TestReceive_UnknownChannel_NotFound(t *testing.T) {
	t.Setenv("META_APP_SECRET", testAppSecret)

	body := []byte(`{}`)
	uc := &fakeUseCase{}
	h := handler.New(uc)

	req := httptest.NewRequest(http.MethodPost, "/webhooks/meta/telegram", strings.NewReader(string(body)))
	req.Header.Set("X-Hub-Signature-256", sign(testAppSecret, body))
	req = withChannelParam(req, "telegram")
	rec := httptest.NewRecorder()

	h.Receive(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for an unsupported channel, got %d", rec.Code)
	}
}

// --- Verify (handshake) -----------------------------------------------------

func TestVerify_CorrectToken_EchoesChallenge(t *testing.T) {
	t.Setenv("META_WEBHOOK_VERIFY_TOKEN", "the-verify-token")
	h := handler.New(&fakeUseCase{})

	req := httptest.NewRequest(http.MethodGet, "/webhooks/meta/whatsapp?hub.mode=subscribe&hub.verify_token=the-verify-token&hub.challenge=abc123", nil)
	rec := httptest.NewRecorder()

	h.Verify(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if rec.Body.String() != "abc123" {
		t.Fatalf("expected the hub.challenge to be echoed back verbatim, got %q", rec.Body.String())
	}
}

func TestVerify_WrongToken_Forbidden(t *testing.T) {
	t.Setenv("META_WEBHOOK_VERIFY_TOKEN", "the-verify-token")
	h := handler.New(&fakeUseCase{})

	req := httptest.NewRequest(http.MethodGet, "/webhooks/meta/whatsapp?hub.mode=subscribe&hub.verify_token=wrong&hub.challenge=abc123", nil)
	rec := httptest.NewRecorder()

	h.Verify(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for the wrong verify_token, got %d", rec.Code)
	}
}
