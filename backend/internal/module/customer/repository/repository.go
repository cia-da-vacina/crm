package repository

import (
	"context"

	"github.com/cia-da-vacina/crm/backend/internal/domain/entity"
	"github.com/cia-da-vacina/crm/backend/internal/module/customer/model"
	"github.com/cia-da-vacina/crm/backend/pkg/database"
)

const customerColumns = `id, display_name, identification, primary_phone, unit_id, created_at, updated_at`

// filterClause é reaproveitado entre a query de itens e a de COUNT — os dois
// precisam filtrar exatamente igual, senão total diverge da página.
const filterClause = `
	($1 = '' OR display_name ILIKE '%' || $1 || '%' OR primary_phone ILIKE '%' || $1 || '%')
	AND ($2 = '' OR identification = $2)
	AND ($3::uuid IS NULL OR unit_id = $3::uuid)
	AND ($4 OR unit_id IS NULL OR unit_id = ANY($5))
`

type Repository struct {
	db *database.DB
}

func New(db *database.DB) *Repository {
	return &Repository{db: db}
}

// List aplica busca (nome/telefone), filtros de identification/unit_id, e a
// visão de unidade do requester (Unscoped=true pula o filtro de unidade —
// só admin). Clientes sem unit_id (ainda não triados) ficam visíveis pra
// todo mundo, não só admin.
func (r *Repository) List(ctx context.Context, f model.ListFilter) ([]entity.Customer, int, error) {
	var unitIDFilter *string
	if f.UnitID != "" {
		unitIDFilter = &f.UnitID
	}
	scopeIDs := f.RequesterUnitIDs
	if scopeIDs == nil {
		scopeIDs = []string{}
	}
	offset := (f.Page - 1) * f.PageSize

	var customers []entity.Customer
	if err := r.db.SelectContext(ctx, &customers, `
		SELECT `+customerColumns+`
		FROM customers
		WHERE `+filterClause+`
		ORDER BY created_at DESC
		LIMIT $6 OFFSET $7
	`, f.Query, f.Identification, unitIDFilter, f.Unscoped, scopeIDs, f.PageSize, offset); err != nil {
		return nil, 0, err
	}

	var total int
	if err := r.db.GetContext(ctx, &total, `
		SELECT COUNT(*) FROM customers WHERE `+filterClause,
		f.Query, f.Identification, unitIDFilter, f.Unscoped, scopeIDs); err != nil {
		return nil, 0, err
	}

	return customers, total, nil
}

func (r *Repository) GetByID(ctx context.Context, id string) (entity.Customer, error) {
	var c entity.Customer
	if err := r.db.GetContext(ctx, &c, `SELECT `+customerColumns+` FROM customers WHERE id = $1`, id); err != nil {
		return entity.Customer{}, err
	}
	return c, nil
}

func (r *Repository) GetIdentities(ctx context.Context, customerID string) ([]entity.CustomerIdentity, error) {
	var identities []entity.CustomerIdentity
	if err := r.db.SelectContext(ctx, &identities, `
		SELECT id, customer_id, channel, external_id, display_handle, phone_e164, verified_at, created_at
		FROM customer_identities WHERE customer_id = $1
		ORDER BY created_at
	`, customerID); err != nil {
		return nil, err
	}
	return identities, nil
}

// GetIdentitiesBulk evita N+1 em List — mesma técnica usada em
// internal/module/user/repository para unit_ids.
func (r *Repository) GetIdentitiesBulk(ctx context.Context, customerIDs []string) (map[string][]entity.CustomerIdentity, error) {
	result := make(map[string][]entity.CustomerIdentity, len(customerIDs))
	if len(customerIDs) == 0 {
		return result, nil
	}

	var identities []entity.CustomerIdentity
	if err := r.db.SelectContext(ctx, &identities, `
		SELECT id, customer_id, channel, external_id, display_handle, phone_e164, verified_at, created_at
		FROM customer_identities WHERE customer_id = ANY($1)
		ORDER BY created_at
	`, customerIDs); err != nil {
		return nil, err
	}

	for _, identity := range identities {
		result[identity.CustomerID] = append(result[identity.CustomerID], identity)
	}
	return result, nil
}
