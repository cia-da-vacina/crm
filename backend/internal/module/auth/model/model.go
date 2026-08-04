package model

import "github.com/golang-jwt/jwt/v5"

// UserClaims é o payload assinado no access token. Carrega role e unit_ids
// pra middlewares/usecases decidirem autorização sem bater no banco a cada
// request — só é recalculado no login/refresh (ver auth/usecase).
type UserClaims struct {
	jwt.RegisteredClaims
	Role    string   `json:"role"`
	UnitIDs []string `json:"unit_ids"`
}
