// Package repository roda as agregações do dashboard como várias queries
// simples (COUNT/GROUP BY diretos), não uma camada de BI — decisão de
// arquitetura documentada em backend/ARCHITECTURE.md §6: volume do MVP
// (~2000 msgs/dia) não justifica materialized view nem pré-agregação.
package repository

import (
	"context"

	"github.com/cia-da-vacina/crm/backend/pkg/database"
)

type Repository struct {
	db *database.DB
}

func New(db *database.DB) *Repository {
	return &Repository{db: db}
}

// Counts agrupa os campos que dá pra tirar de uma única passada pela tabela
// conversations com COUNT(*) FILTER — evita 9 roundtrips pro banco.
type Counts struct {
	OpenConversations int `db:"open_conversations"`
	Closed            int `db:"closed"`
	NotClosed         int `db:"not_closed"`
	AITriage          int `db:"ai_triage"`
	Human             int `db:"human"`
	Unclaimed         int `db:"unclaimed"`
	StaleOpen         int `db:"stale_open"`
	AwaitingPhone     int `db:"awaiting_phone"`
	WindowExpiring    int `db:"window_expiring"`
}

func (r *Repository) GetCounts(ctx context.Context, unscoped bool, unitIDs []string) (Counts, error) {
	var c Counts
	err := r.db.GetContext(ctx, &c, `
		SELECT
			COUNT(*) FILTER (WHERE pipeline_stage NOT IN ('fechado','nao_fechado')) AS open_conversations,
			COUNT(*) FILTER (WHERE pipeline_stage = 'fechado') AS closed,
			COUNT(*) FILTER (WHERE pipeline_stage = 'nao_fechado') AS not_closed,
			COUNT(*) FILTER (WHERE mode = 'ai_triage') AS ai_triage,
			COUNT(*) FILTER (WHERE mode = 'human') AS human,
			COUNT(*) FILTER (WHERE owner_id IS NULL AND pipeline_stage NOT IN ('fechado','nao_fechado')) AS unclaimed,
			COUNT(*) FILTER (WHERE pipeline_stage NOT IN ('fechado','nao_fechado') AND last_message_at < NOW() - INTERVAL '24 hours') AS stale_open,
			COUNT(*) FILTER (WHERE pipeline_stage NOT IN ('fechado','nao_fechado') AND phone_gate IN ('required','pending_verification')) AS awaiting_phone,
			COUNT(*) FILTER (WHERE window_expires_at IS NOT NULL AND window_expires_at BETWEEN NOW() AND NOW() + INTERVAL '4 hours') AS window_expiring
		FROM conversations
		WHERE ($1 OR unit_id = ANY($2))
	`, unscoped, unitIDsOrEmpty(unitIDs))
	return c, err
}

// GetAwaitingReply conta conversas abertas cuja última mensagem é inbound —
// exige olhar a mensagem mais recente de cada conversa (docs/BACKEND-CONTRACT.md §6).
func (r *Repository) GetAwaitingReply(ctx context.Context, unscoped bool, unitIDs []string) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count, `
		SELECT COUNT(*) FROM conversations c
		WHERE ($1 OR c.unit_id = ANY($2))
		  AND c.pipeline_stage NOT IN ('fechado', 'nao_fechado')
		  AND (SELECT m.direction FROM messages m WHERE m.conversation_id = c.id ORDER BY m.created_at DESC LIMIT 1) = 'in'
	`, unscoped, unitIDsOrEmpty(unitIDs))
	return count, err
}

// GetOpenEngagements conta social_engagements pendentes de triagem — tabela
// própria (fase 8), então é uma query separada de GetCounts, mesmo padrão de
// GetAwaitingReply/GetFollowUpCounts.
func (r *Repository) GetOpenEngagements(ctx context.Context, unscoped bool, unitIDs []string) (int, error) {
	var count int
	err := r.db.GetContext(ctx, &count, `
		SELECT COUNT(*) FROM social_engagements
		WHERE status = 'open' AND ($1 OR unit_id = ANY($2))
	`, unscoped, unitIDsOrEmpty(unitIDs))
	return count, err
}

type groupCount struct {
	Key   string `db:"key"`
	Count int    `db:"count"`
}

func (r *Repository) GetByStage(ctx context.Context, unscoped bool, unitIDs []string) (map[string]int, error) {
	return r.groupBy(ctx, `
		SELECT pipeline_stage AS key, COUNT(*) AS count FROM conversations
		WHERE ($1 OR unit_id = ANY($2)) GROUP BY pipeline_stage
	`, unscoped, unitIDs)
}

func (r *Repository) GetByChannel(ctx context.Context, unscoped bool, unitIDs []string) (map[string]int, error) {
	return r.groupBy(ctx, `
		SELECT channel AS key, COUNT(*) AS count FROM conversations
		WHERE ($1 OR unit_id = ANY($2)) GROUP BY channel
	`, unscoped, unitIDs)
}

// GetByIntent só conta abertas — docs/BACKEND-CONTRACT.md §6: "só abertas;
// intent nulo conta em outro".
func (r *Repository) GetByIntent(ctx context.Context, unscoped bool, unitIDs []string) (map[string]int, error) {
	return r.groupBy(ctx, `
		SELECT COALESCE(intent, 'outro') AS key, COUNT(*) AS count FROM conversations
		WHERE ($1 OR unit_id = ANY($2)) AND pipeline_stage NOT IN ('fechado', 'nao_fechado')
		GROUP BY COALESCE(intent, 'outro')
	`, unscoped, unitIDs)
}

func (r *Repository) GetClosedByChannel(ctx context.Context, unscoped bool, unitIDs []string) (map[string]int, error) {
	return r.groupBy(ctx, `
		SELECT channel AS key, COUNT(*) AS count FROM conversations
		WHERE ($1 OR unit_id = ANY($2)) AND pipeline_stage = 'fechado' GROUP BY channel
	`, unscoped, unitIDs)
}

func (r *Repository) GetNotClosedByChannel(ctx context.Context, unscoped bool, unitIDs []string) (map[string]int, error) {
	return r.groupBy(ctx, `
		SELECT channel AS key, COUNT(*) AS count FROM conversations
		WHERE ($1 OR unit_id = ANY($2)) AND pipeline_stage = 'nao_fechado' GROUP BY channel
	`, unscoped, unitIDs)
}

func (r *Repository) groupBy(ctx context.Context, query string, unscoped bool, unitIDs []string) (map[string]int, error) {
	var rows []groupCount
	if err := r.db.SelectContext(ctx, &rows, query, unscoped, unitIDsOrEmpty(unitIDs)); err != nil {
		return nil, err
	}
	result := make(map[string]int, len(rows))
	for _, row := range rows {
		result[row.Key] = row.Count
	}
	return result, nil
}

func (r *Repository) GetFollowUpCounts(ctx context.Context, unscoped bool, unitIDs []string) (awaiting, overdue int, err error) {
	var row struct {
		Awaiting int `db:"awaiting"`
		Overdue  int `db:"overdue"`
	}
	err = r.db.GetContext(ctx, &row, `
		SELECT
			COUNT(*) FILTER (WHERE status = 'open') AS awaiting,
			COUNT(*) FILTER (WHERE status = 'open' AND due_at < NOW()) AS overdue
		FROM follow_ups
		WHERE ($1 OR unit_id = ANY($2))
	`, unscoped, unitIDsOrEmpty(unitIDs))
	return row.Awaiting, row.Overdue, err
}

type UnitRow struct {
	UnitID    string `db:"unit_id"`
	UnitName  string `db:"unit_name"`
	Open      int    `db:"open"`
	Closed    int    `db:"closed"`
	NotClosed int    `db:"not_closed"`
	Unclaimed int    `db:"unclaimed"`
}

// GetUnitBreakdown sempre cobre unitIDs por inteiro (as unidades acessíveis
// ao usuário) — não é afetado pelo filtro unit_id? da query principal
// (docs/BACKEND-CONTRACT.md §6: "sempre as unidades acessíveis ao usuário").
func (r *Repository) GetUnitBreakdown(ctx context.Context, unitIDs []string) ([]UnitRow, error) {
	var rows []UnitRow
	err := r.db.SelectContext(ctx, &rows, `
		SELECT u.id AS unit_id, u.name AS unit_name,
			COUNT(c.id) FILTER (WHERE c.pipeline_stage NOT IN ('fechado', 'nao_fechado')) AS open,
			COUNT(c.id) FILTER (WHERE c.pipeline_stage = 'fechado') AS closed,
			COUNT(c.id) FILTER (WHERE c.pipeline_stage = 'nao_fechado') AS not_closed,
			COUNT(c.id) FILTER (WHERE c.owner_id IS NULL AND c.pipeline_stage NOT IN ('fechado', 'nao_fechado')) AS unclaimed
		FROM units u
		LEFT JOIN conversations c ON c.unit_id = u.id
		WHERE u.id = ANY($1)
		GROUP BY u.id, u.name
		ORDER BY u.name
	`, unitIDsOrEmpty(unitIDs))
	return rows, err
}

func (r *Repository) GetAwaitingFollowupByUnit(ctx context.Context, unitIDs []string) (map[string]int, error) {
	var rows []struct {
		UnitID string `db:"unit_id"`
		Count  int    `db:"count"`
	}
	if err := r.db.SelectContext(ctx, &rows, `
		SELECT unit_id, COUNT(*) FILTER (WHERE status = 'open') AS count
		FROM follow_ups WHERE unit_id = ANY($1) GROUP BY unit_id
	`, unitIDsOrEmpty(unitIDs)); err != nil {
		return nil, err
	}
	result := make(map[string]int, len(rows))
	for _, row := range rows {
		result[row.UnitID] = row.Count
	}
	return result, nil
}

func (r *Repository) ListAllUnitIDs(ctx context.Context) ([]string, error) {
	var ids []string
	err := r.db.SelectContext(ctx, &ids, `SELECT id FROM units`)
	return ids, err
}

// unitIDsOrEmpty evita passar nil pro driver — ANY(NULL) no Postgres não dá
// erro, mas é mais claro/seguro sempre mandar um slice concreto.
func unitIDsOrEmpty(ids []string) []string {
	if ids == nil {
		return []string{}
	}
	return ids
}
