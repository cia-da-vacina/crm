// Package middleware fornece os middlewares de autorização compartilhados
// por todos os módulos de domínio: autenticação (RequireAuth) e checagem de
// papel (RequireRole). Fica fora de qualquer módulo específico porque toda
// rota autenticada da API depende dele.
package middleware

import (
	"context"
	"net/http"
	"strings"

	authmodel "github.com/cia-da-vacina/crm/backend/internal/module/auth/model"
	httppkg "github.com/cia-da-vacina/crm/backend/pkg/http"
	"github.com/cia-da-vacina/crm/backend/pkg/jwt"
)

type contextKey string

const claimsContextKey contextKey = "auth_claims"

type Middleware struct {
	jwt *jwt.Service
}

func New(jwtSvc *jwt.Service) *Middleware {
	return &Middleware{jwt: jwtSvc}
}

// RequireAuth valida o Bearer token e injeta os claims no contexto da
// request. Handlers/usecases downstream leem via ClaimsFromContext.
func (m *Middleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := extractBearerToken(r)
		if token == "" {
			httppkg.Unauthorized(w, "")
			return
		}

		var claims authmodel.UserClaims
		if err := m.jwt.Validate(token, &claims); err != nil {
			httppkg.Unauthorized(w, "")
			return
		}

		ctx := context.WithValue(r.Context(), claimsContextKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireRole só deixa passar se o papel do usuário autenticado estiver na
// lista. Deve vir depois de RequireAuth na cadeia de middlewares.
func RequireRole(roles ...string) func(http.Handler) http.Handler {
	allowed := make(map[string]struct{}, len(roles))
	for _, role := range roles {
		allowed[role] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := ClaimsFromContext(r.Context())
			if !ok {
				httppkg.Unauthorized(w, "")
				return
			}
			if _, ok := allowed[claims.Role]; !ok {
				httppkg.Forbidden(w, "")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// ClaimsFromContext lê os claims injetados por RequireAuth.
func ClaimsFromContext(ctx context.Context) (authmodel.UserClaims, bool) {
	claims, ok := ctx.Value(claimsContextKey).(authmodel.UserClaims)
	return claims, ok
}

func extractBearerToken(r *http.Request) string {
	header := r.Header.Get("Authorization")
	if header == "" {
		return ""
	}
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}
