package meta

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

// VerifySignature valida o header X-Hub-Signature-256 (HMAC-SHA256 com o App
// Secret) — obrigatório em 100% dos webhooks Meta, sem exceção, mesmo em dev
// (docs/BACKEND-CONTRACT.md §9). body deve ser os bytes crus da request,
// lidos antes de qualquer parse de JSON (o HMAC é sobre o payload exato que
// chegou na rede, não sobre uma re-serialização).
func VerifySignature(appSecret string, body []byte, signatureHeader string) bool {
	const prefix = "sha256="
	if !strings.HasPrefix(signatureHeader, prefix) {
		return false
	}

	expected, err := hex.DecodeString(strings.TrimPrefix(signatureHeader, prefix))
	if err != nil {
		return false
	}

	mac := hmac.New(sha256.New, []byte(appSecret))
	mac.Write(body)
	return hmac.Equal(mac.Sum(nil), expected)
}
