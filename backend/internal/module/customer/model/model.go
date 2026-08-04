package model

import (
	"time"

	"github.com/cia-da-vacina/crm/backend/internal/domain/entity"
)

// Customer é o shape exposto pela API — o contrato usa a mesma forma tanto
// na listagem quanto no detalhe, sempre com identities embutido
// (docs/BACKEND-CONTRACT.md §3).
type Customer struct {
	ID             string                    `json:"id"`
	DisplayName    string                    `json:"display_name"`
	Identification string                    `json:"identification"`
	PrimaryPhone   *string                   `json:"primary_phone"`
	UnitID         *string                   `json:"unit_id,omitempty"`
	Identities     []entity.CustomerIdentity `json:"identities"`
	CreatedAt      time.Time                 `json:"created_at"`
	UpdatedAt      time.Time                 `json:"updated_at"`
}

type ListResult struct {
	Items []Customer `json:"items"`
	Total int        `json:"total"`
}

// ListFilter carrega tanto os filtros de busca (Query/Identification/UnitID)
// quanto o escopo de visibilidade do requester (Unscoped/RequesterUnitIDs) —
// mesma regra de "visão de unidade" já aplicada a users/units.
type ListFilter struct {
	Query            string
	Identification   string
	UnitID           string
	Unscoped         bool
	RequesterUnitIDs []string
	Page             int
	PageSize         int
}
