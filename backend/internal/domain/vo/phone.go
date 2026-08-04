package vo

import (
	"regexp"
	"strings"
)

// e164Pattern: "+" seguido de 8 a 15 dígitos, primeiro dígito != 0.
// docs/BACKEND-CONTRACT.md §3 usa E.164 (ex.: +5511999998888) como chave de
// negócio pra merge cross-canal — validação estrita aqui evita telefone mal
// formado virar chave de merge.
var e164Pattern = regexp.MustCompile(`^\+[1-9]\d{7,14}$`)

func IsE164(phone string) bool {
	return e164Pattern.MatchString(phone)
}

// MaskPhone mostra só o prefixo de país+DDD e os últimos 4 dígitos — usado
// em pending_phone_masked (docs/BACKEND-CONTRACT.md §3) pra não expor o
// número completo antes da posse ser confirmada por OTP.
func MaskPhone(e164 string) string {
	if len(e164) < 9 {
		return strings.Repeat("*", len(e164))
	}
	visibleStart := e164[:5]
	visibleEnd := e164[len(e164)-4:]
	maskedLen := len(e164) - len(visibleStart) - len(visibleEnd)
	return visibleStart + strings.Repeat("*", maskedLen) + visibleEnd
}
