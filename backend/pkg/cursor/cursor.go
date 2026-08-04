// Package cursor implementa paginação por cursor opaco (base64 de
// "timestamp|id") — o padrão CursorPage<T> usado pelo contrato em todo
// endpoint de alto volume/tempo real (inbox, mensagens, engagements,
// follow-ups), em vez de paginação por página/offset.
package cursor

import (
	"encoding/base64"
	"errors"
	"strings"
	"time"
)

var ErrInvalid = errors.New("cursor: invalid cursor")

// Encode produz um cursor opaco a partir do timestamp e id do último item da
// página atual — o client não deve interpretar o conteúdo, só devolver de
// volta no próximo `cursor=`.
func Encode(t time.Time, id string) string {
	raw := t.UTC().Format(time.RFC3339Nano) + "|" + id
	return base64.URLEncoding.EncodeToString([]byte(raw))
}

// Decode reverte um cursor gerado por Encode.
func Decode(raw string) (time.Time, string, error) {
	if raw == "" {
		return time.Time{}, "", nil
	}

	decoded, err := base64.URLEncoding.DecodeString(raw)
	if err != nil {
		return time.Time{}, "", ErrInvalid
	}

	parts := strings.SplitN(string(decoded), "|", 2)
	if len(parts) != 2 {
		return time.Time{}, "", ErrInvalid
	}

	t, err := time.Parse(time.RFC3339Nano, parts[0])
	if err != nil {
		return time.Time{}, "", ErrInvalid
	}

	return t, parts[1], nil
}
