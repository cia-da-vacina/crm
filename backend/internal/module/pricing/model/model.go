package model

import "time"

type Rate struct {
	Category  string    `json:"category"`
	RateBRL   float64   `json:"rate_brl"`
	Billable  bool      `json:"billable"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UpdateRateRequest é parcial: RateBRL/Billable ausentes mantêm o valor
// atual — mesma convenção de ponteiro-como-sinal-de-presença usada em
// UpdatePopRequest/UpdateMetaSettingsPayload.
type UpdateRateRequest struct {
	RateBRL  *float64 `json:"rate_brl"  validate:"omitempty,gte=0"`
	Billable *bool    `json:"billable"`
}
