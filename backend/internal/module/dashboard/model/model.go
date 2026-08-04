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
