package store

import (
	"context"
	"errors"

	"github.com/cia-vacina/crm/backend/internal/domain"
	"github.com/google/uuid"
)

var (
	ErrNotFound      = errors.New("not found")
	ErrConflict      = errors.New("conflict")
	ErrInvalidInput  = errors.New("invalid input")
	ErrUnauthorized  = errors.New("unauthorized")
	ErrForbidden     = errors.New("forbidden")
	ErrInactiveUser  = errors.New("inactive user")
)

type Store interface {
	// Users
	CreateUser(ctx context.Context, user domain.User) (domain.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (domain.User, error)
	GetUserByEmail(ctx context.Context, email string) (domain.User, error)
	ListUsers(ctx context.Context) ([]domain.User, error)
	UpdateUser(ctx context.Context, user domain.User) (domain.User, error)
	DeleteUser(ctx context.Context, id uuid.UUID) error
	SetUserUnits(ctx context.Context, userID uuid.UUID, unitIDs []uuid.UUID) error

	// Units
	CreateUnit(ctx context.Context, unit domain.Unit) (domain.Unit, error)
	GetUnitByID(ctx context.Context, id uuid.UUID) (domain.Unit, error)
	ListUnits(ctx context.Context) ([]domain.Unit, error)
	UpdateUnit(ctx context.Context, unit domain.Unit) (domain.Unit, error)

	// Auth helpers
	ListUnitsForUser(ctx context.Context, userID uuid.UUID) ([]domain.Unit, error)
}
