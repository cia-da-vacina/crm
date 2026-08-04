// Package model define os DTOs de request/response do módulo user — nunca
// expõe entity.User diretamente (ela carrega password_hash).
package model

import "time"

// User é o shape de usuário exposto pela API. UnitIDs fica de fora do
// payload de login (docs/BACKEND-CONTRACT.md §1) mas presente em /me e no
// CRUD de /users — por isso é `omitempty` e cada usecase decide se preenche.
type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Role      string    `json:"role"`
	Active    bool      `json:"active"`
	UnitIDs   []string  `json:"unit_ids,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateUserRequest struct {
	Email    string   `json:"email"    validate:"required,email"`
	Password string   `json:"password" validate:"required,min=8"`
	Name     string   `json:"name"     validate:"required"`
	Role     string   `json:"role"     validate:"required,oneof=admin manager supervisor agent"`
	UnitIDs  []string `json:"unit_ids"`
}

// UpdateUserRequest é parcial: campos ausentes (nil) não são alterados.
// O usecase decide quais campos um requester não-admin pode setar em si
// mesmo (ver internal/module/user/usecase).
type UpdateUserRequest struct {
	Name     *string `json:"name"`
	Role     *string `json:"role"     validate:"omitempty,oneof=admin manager supervisor agent"`
	Active   *bool   `json:"active"`
	Password *string `json:"password" validate:"omitempty,min=8"`
}

// SetUnitsRequest substitui integralmente o vínculo usuário×unidade — não é
// incremental (docs/BACKEND-CONTRACT.md §2, PUT /users/:id/units).
type SetUnitsRequest struct {
	UnitIDs []string `json:"unit_ids" validate:"required"`
}

type ListUsersQuery struct {
	Page     int
	PageSize int
}

type ListUsersResult struct {
	Items []User `json:"items"`
	Total int    `json:"total"`
}
