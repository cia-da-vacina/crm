package meta

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

func sign(secret string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

func TestVerifySignature_Valid(t *testing.T) {
	body := []byte(`{"object":"whatsapp_business_account"}`)
	header := sign("app-secret", body)

	if !VerifySignature("app-secret", body, header) {
		t.Fatal("expected valid signature to verify")
	}
}

func TestVerifySignature_WrongSecret(t *testing.T) {
	body := []byte(`{"object":"whatsapp_business_account"}`)
	header := sign("app-secret", body)

	if VerifySignature("wrong-secret", body, header) {
		t.Fatal("expected signature with wrong secret to fail")
	}
}

func TestVerifySignature_TamperedBody(t *testing.T) {
	header := sign("app-secret", []byte(`{"object":"whatsapp_business_account"}`))
	tampered := []byte(`{"object":"tampered"}`)

	if VerifySignature("app-secret", tampered, header) {
		t.Fatal("expected signature to fail for tampered body")
	}
}

func TestVerifySignature_MissingPrefix(t *testing.T) {
	body := []byte(`{}`)
	if VerifySignature("app-secret", body, "not-a-valid-header") {
		t.Fatal("expected malformed header to fail")
	}
}

func TestVerifySignature_EmptyHeader(t *testing.T) {
	if VerifySignature("app-secret", []byte(`{}`), "") {
		t.Fatal("expected empty header to fail")
	}
}
