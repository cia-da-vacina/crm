package model

import "time"

type Template struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Category       string    `json:"category"`
	LanguageCode   string    `json:"language_code"`
	Body           string    `json:"body"`
	VariableCount  int       `json:"variable_count"`
	ApprovalStatus string    `json:"approval_status"`
	Active         bool      `json:"active"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type CreateTemplateRequest struct {
	Name          string `json:"name"           validate:"required"`
	Category      string `json:"category"       validate:"required,oneof=marketing utility authentication"`
	LanguageCode  string `json:"language_code"`
	Body          string `json:"body"           validate:"required"`
	VariableCount int    `json:"variable_count" validate:"gte=0"`
}

// UpdateTemplateRequest é parcial — mesma convenção de ponteiro-como-sinal-
// de-presença de UpdatePopRequest.
type UpdateTemplateRequest struct {
	Name           *string `json:"name"`
	Category       *string `json:"category"        validate:"omitempty,oneof=marketing utility authentication"`
	LanguageCode   *string `json:"language_code"`
	Body           *string `json:"body"`
	VariableCount  *int    `json:"variable_count"  validate:"omitempty,gte=0"`
	ApprovalStatus *string `json:"approval_status" validate:"omitempty,oneof=pending approved rejected"`
	Active         *bool   `json:"active"`
}
