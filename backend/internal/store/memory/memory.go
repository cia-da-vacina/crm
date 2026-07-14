package memory

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/cia-vacina/crm/backend/internal/auth"
	"github.com/cia-vacina/crm/backend/internal/domain"
	"github.com/cia-vacina/crm/backend/internal/store"
	"github.com/google/uuid"
)

type Memory struct {
	mu        sync.RWMutex
	users     map[uuid.UUID]domain.User
	units     map[uuid.UUID]domain.Unit
	userUnits map[uuid.UUID]map[uuid.UUID]struct{}
}

func New() *Memory {
	m := &Memory{
		users:     make(map[uuid.UUID]domain.User),
		units:     make(map[uuid.UUID]domain.Unit),
		userUnits: make(map[uuid.UUID]map[uuid.UUID]struct{}),
	}
	m.seed()
	return m
}

func (m *Memory) seed() {
	now := time.Now().UTC()
	unitSeeds := []domain.Unit{
		{ID: uuid.MustParse("11111111-1111-1111-1111-111111111101"), Name: "Unidade Centro", Code: "centro", Timezone: "America/Sao_Paulo", Active: true, CreatedAt: now},
		{ID: uuid.MustParse("11111111-1111-1111-1111-111111111102"), Name: "Unidade Norte", Code: "norte", Timezone: "America/Sao_Paulo", Active: true, CreatedAt: now},
		{ID: uuid.MustParse("11111111-1111-1111-1111-111111111103"), Name: "Unidade Sul", Code: "sul", Timezone: "America/Sao_Paulo", Active: true, CreatedAt: now},
		{ID: uuid.MustParse("11111111-1111-1111-1111-111111111104"), Name: "Unidade Leste", Code: "leste", Timezone: "America/Sao_Paulo", Active: true, CreatedAt: now},
		{ID: uuid.MustParse("11111111-1111-1111-1111-111111111105"), Name: "Unidade Oeste", Code: "oeste", Timezone: "America/Sao_Paulo", Active: true, CreatedAt: now},
	}
	for _, u := range unitSeeds {
		m.units[u.ID] = u
	}

	hash, _ := auth.HashPassword("admin123")
	adminID := uuid.MustParse("22222222-2222-2222-2222-222222222201")
	admin := domain.User{
		ID:           adminID,
		Email:        "admin@ciadavacina.com.br",
		PasswordHash: hash,
		Name:         "Administrador",
		Role:         domain.RoleAdmin,
		Active:       true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	m.users[adminID] = admin
	m.userUnits[adminID] = make(map[uuid.UUID]struct{})
	for id := range m.units {
		m.userUnits[adminID][id] = struct{}{}
	}

	agentHash, _ := auth.HashPassword("agent123")
	agentID := uuid.MustParse("22222222-2222-2222-2222-222222222202")
	agent := domain.User{
		ID:           agentID,
		Email:        "atendente@ciadavacina.com.br",
		PasswordHash: agentHash,
		Name:         "Atendente Demo",
		Role:         domain.RoleAgent,
		Active:       true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	m.users[agentID] = agent
	centro := uuid.MustParse("11111111-1111-1111-1111-111111111101")
	m.userUnits[agentID] = map[uuid.UUID]struct{}{centro: {}}
}

func (m *Memory) CreateUser(_ context.Context, user domain.User) (domain.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	email := strings.ToLower(strings.TrimSpace(user.Email))
	for _, u := range m.users {
		if u.Email == email {
			return domain.User{}, store.ErrConflict
		}
	}
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}
	now := time.Now().UTC()
	user.Email = email
	user.CreatedAt = now
	user.UpdatedAt = now
	m.users[user.ID] = user
	m.userUnits[user.ID] = make(map[uuid.UUID]struct{})
	for _, id := range user.UnitIDs {
		if _, ok := m.units[id]; ok {
			m.userUnits[user.ID][id] = struct{}{}
		}
	}
	return m.withUnitsLocked(user), nil
}

func (m *Memory) GetUserByID(_ context.Context, id uuid.UUID) (domain.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	u, ok := m.users[id]
	if !ok {
		return domain.User{}, store.ErrNotFound
	}
	return m.withUnitsLocked(u), nil
}

func (m *Memory) GetUserByEmail(_ context.Context, email string) (domain.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	email = strings.ToLower(strings.TrimSpace(email))
	for _, u := range m.users {
		if u.Email == email {
			return m.withUnitsLocked(u), nil
		}
	}
	return domain.User{}, store.ErrNotFound
}

func (m *Memory) ListUsers(_ context.Context) ([]domain.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]domain.User, 0, len(m.users))
	for _, u := range m.users {
		out = append(out, m.withUnitsLocked(u))
	}
	return out, nil
}

func (m *Memory) UpdateUser(_ context.Context, user domain.User) (domain.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	existing, ok := m.users[user.ID]
	if !ok {
		return domain.User{}, store.ErrNotFound
	}
	user.Email = existing.Email
	user.CreatedAt = existing.CreatedAt
	if user.PasswordHash == "" {
		user.PasswordHash = existing.PasswordHash
	}
	user.UpdatedAt = time.Now().UTC()
	m.users[user.ID] = user
	return m.withUnitsLocked(user), nil
}

func (m *Memory) DeleteUser(_ context.Context, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.users[id]; !ok {
		return store.ErrNotFound
	}
	delete(m.users, id)
	delete(m.userUnits, id)
	return nil
}

func (m *Memory) SetUserUnits(_ context.Context, userID uuid.UUID, unitIDs []uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.users[userID]; !ok {
		return store.ErrNotFound
	}
	set := make(map[uuid.UUID]struct{})
	for _, id := range unitIDs {
		if _, ok := m.units[id]; !ok {
			return store.ErrInvalidInput
		}
		set[id] = struct{}{}
	}
	m.userUnits[userID] = set
	return nil
}

func (m *Memory) CreateUnit(_ context.Context, unit domain.Unit) (domain.Unit, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	code := strings.ToLower(strings.TrimSpace(unit.Code))
	for _, u := range m.units {
		if u.Code == code {
			return domain.Unit{}, store.ErrConflict
		}
	}
	if unit.ID == uuid.Nil {
		unit.ID = uuid.New()
	}
	if unit.Timezone == "" {
		unit.Timezone = "America/Sao_Paulo"
	}
	unit.Code = code
	unit.CreatedAt = time.Now().UTC()
	m.units[unit.ID] = unit
	return unit, nil
}

func (m *Memory) GetUnitByID(_ context.Context, id uuid.UUID) (domain.Unit, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	u, ok := m.units[id]
	if !ok {
		return domain.Unit{}, store.ErrNotFound
	}
	return u, nil
}

func (m *Memory) ListUnits(_ context.Context) ([]domain.Unit, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]domain.Unit, 0, len(m.units))
	for _, u := range m.units {
		out = append(out, u)
	}
	return out, nil
}

func (m *Memory) UpdateUnit(_ context.Context, unit domain.Unit) (domain.Unit, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	existing, ok := m.units[unit.ID]
	if !ok {
		return domain.Unit{}, store.ErrNotFound
	}
	unit.Code = existing.Code
	unit.CreatedAt = existing.CreatedAt
	m.units[unit.ID] = unit
	return unit, nil
}

func (m *Memory) ListUnitsForUser(_ context.Context, userID uuid.UUID) ([]domain.Unit, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if _, ok := m.users[userID]; !ok {
		return nil, store.ErrNotFound
	}
	set := m.userUnits[userID]
	out := make([]domain.Unit, 0, len(set))
	for id := range set {
		if u, ok := m.units[id]; ok {
			out = append(out, u)
		}
	}
	return out, nil
}

func (m *Memory) withUnitsLocked(u domain.User) domain.User {
	set := m.userUnits[u.ID]
	ids := make([]uuid.UUID, 0, len(set))
	for id := range set {
		ids = append(ids, id)
	}
	u.UnitIDs = ids
	return u
}
