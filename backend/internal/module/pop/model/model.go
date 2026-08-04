package model

import "time"

type Pop struct {
	ID         string    `json:"id"`
	Title      string    `json:"title"`
	Body       string    `json:"body"`
	IntentTags []string  `json:"intent_tags"`
	Active     bool      `json:"active"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type CreatePopRequest struct {
	Title      string   `json:"title" validate:"required"`
	Body       string   `json:"body"  validate:"required"`
	IntentTags []string `json:"intent_tags"`
	Active     *bool    `json:"active"`
}

// UpdatePopRequest é parcial: um campo ausente no JSON não é tocado.
// IntentTags usa nil-vs-não-nil como sinal de presença (igual *string em
// outros módulos) — []string{} explícito no body limpa as tags; campo
// ausente mantém as atuais.
type UpdatePopRequest struct {
	Title      *string  `json:"title"`
	Body       *string  `json:"body"`
	IntentTags []string `json:"intent_tags"`
	Active     *bool    `json:"active"`
}
