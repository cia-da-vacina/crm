package domain

import (
	"time"

	"github.com/google/uuid"
)

type Role string

const (
	RoleAdmin      Role = "admin"
	RoleManager    Role = "manager"
	RoleSupervisor Role = "supervisor"
	RoleAgent      Role = "agent"
)

type User struct {
	ID           uuid.UUID   `json:"id"`
	Email        string      `json:"email"`
	PasswordHash string      `json:"-"`
	Name         string      `json:"name"`
	Role         Role        `json:"role"`
	Active       bool        `json:"active"`
	UnitIDs      []uuid.UUID `json:"unit_ids,omitempty"`
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at"`
}

type Unit struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Code      string    `json:"code"`
	Timezone  string    `json:"timezone"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
}

type MeResponse struct {
	User
	Units []Unit `json:"units"`
}
