package http

import (
	"net/http"
	"os"
	"strings"
)

// CORS retorna um middleware que adiciona os headers de CORS em todas as respostas
// e responde imediatamente com 204 em requisições OPTIONS (preflight).
//
// A origem permitida é lida de CORS_ALLOWED_ORIGINS (separadas por vírgula).
// Se não definida, usa "*" — adequado apenas para desenvolvimento. Em produção
// deve ser travada só para o domínio do BFF Next.js (ver docs/BACKEND-CONTRACT.md
// §9: nenhum browser deve conseguir falar direto com o backend).
func CORS(next http.Handler) http.Handler {
	rawOrigins := os.Getenv("CORS_ALLOWED_ORIGINS")

	var allowedOrigins []string
	if rawOrigins != "" {
		for _, o := range strings.Split(rawOrigins, ",") {
			if trimmed := strings.TrimSpace(o); trimmed != "" {
				allowedOrigins = append(allowedOrigins, trimmed)
			}
		}
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		allowed := resolveOrigin(origin, allowedOrigins)
		w.Header().Set("Access-Control-Allow-Origin", allowed)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Request-ID")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "86400")

		// Responde o preflight sem passar pelo router (evita 405).
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// resolveOrigin devolve a origem a permitir.
// Se não houver lista configurada retorna "*".
// Se houver lista, reflete a origem da requisição caso esteja na lista,
// ou retorna a primeira origem permitida como fallback.
func resolveOrigin(requestOrigin string, allowed []string) string {
	if len(allowed) == 0 {
		return "*"
	}
	for _, o := range allowed {
		if o == "*" || o == requestOrigin {
			return o
		}
	}
	return allowed[0]
}
