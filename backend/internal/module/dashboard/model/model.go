// Package model define o shape de DashboardSummary (docs/BACKEND-CONTRACT.md §6).
package model

type Summary struct {
	OpenConversations int            `json:"open_conversations"`
	ByStage           map[string]int `json:"by_stage"`
	ByChannel         map[string]int `json:"by_channel"`
	Closed            int            `json:"closed"`
	NotClosed         int            `json:"not_closed"`
	Decided           int            `json:"decided"`
	ConversionRate    float64        `json:"conversion_rate"`
	AITriage          int            `json:"ai_triage"`
	Human             int            `json:"human"`

	Unclaimed        int `json:"unclaimed"`
	AwaitingReply    int `json:"awaiting_reply"`
	StaleOpen        int `json:"stale_open"`
	AwaitingPhone    int `json:"awaiting_phone"`
	WindowExpiring   int `json:"window_expiring"`
	AwaitingFollowup int `json:"awaiting_followup"`
	OverdueFollowups int `json:"overdue_followups"`
	OpenEngagements  int `json:"open_engagements"`

	ByIntent           map[string]int `json:"by_intent"`
	ClosedByChannel    map[string]int `json:"closed_by_channel"`
	NotClosedByChannel map[string]int `json:"not_closed_by_channel"`

	Units []UnitSummary `json:"units"`
}

type UnitSummary struct {
	UnitID           string  `json:"unit_id"`
	UnitName         string  `json:"unit_name"`
	Open             int     `json:"open"`
	Closed           int     `json:"closed"`
	NotClosed        int     `json:"not_closed"`
	ConversionRate   float64 `json:"conversion_rate"`
	Unclaimed        int     `json:"unclaimed"`
	AwaitingFollowup int     `json:"awaiting_followup"`
}

// CostSummary é o dashboard de custo do plano de adaptação WhatsApp 2026
// (Frente A) — GET /dashboard/costs, endpoint novo (não estende
// GET /dashboard/summary, que docs/BACKEND-CONTRACT.md §6 documenta
// explicitamente como "sem receita/ticket/período"). PricedMessages/
// ConfirmedMessages ajudam a UI a distinguir estimativa de valor
// reconciliado com a Meta — ver Message.PricingConfirmed.
type CostSummary struct {
	TotalCostBRL      float64            `json:"total_cost_brl"`
	TotalOutMessages  int                `json:"total_out_messages"`
	PricedMessages    int                `json:"priced_messages"`
	ConfirmedMessages int                `json:"confirmed_messages"`
	ByCategory        []CostCategoryItem `json:"by_category"`
}

type CostCategoryItem struct {
	Category     string  `json:"category"`
	MessageCount int     `json:"message_count"`
	CostBRL      float64 `json:"cost_brl"`
}
