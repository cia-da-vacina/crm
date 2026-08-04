// Package model define o shape de TriageSummary (docs/BACKEND-CONTRACT.md §4).
package model

type TriageSummary struct {
	ConversationID     string         `json:"conversation_id"`
	Intent             *string        `json:"intent,omitempty"`
	Confidence         *float64       `json:"confidence,omitempty"`
	Summary            string         `json:"summary"`
	SuggestedPops      []string       `json:"suggested_pops,omitempty"`
	ReadyForHandoff    bool           `json:"ready_for_handoff"`
	PhoneGate          string         `json:"phone_gate"`
	PendingPhoneMasked *string        `json:"pending_phone_masked,omitempty"`
	CollectedFields    map[string]any `json:"collected_fields,omitempty"`
}
