package model

import "time"

type FollowUp struct {
	ID             string     `json:"id"`
	ConversationID string     `json:"conversation_id"`
	CustomerID     string     `json:"customer_id"`
	CustomerName   string     `json:"customer_name"`
	CustomerPhone  *string    `json:"customer_phone,omitempty"`
	UnitID         string     `json:"unit_id"`
	PipelineStage  string     `json:"pipeline_stage"`
	DueAt          time.Time  `json:"due_at"`
	Status         string     `json:"status"`
	Note           string     `json:"note"`
	CreatedAt      time.Time  `json:"created_at"`
	CompletedAt    *time.Time `json:"completed_at,omitempty"`
}

type CursorPage struct {
	Items      []FollowUp `json:"items"`
	NextCursor *string    `json:"next_cursor"`
}

type ListFilter struct {
	UnitID           string
	Status           string
	Stage            string
	Cursor           string
	Limit            int
	Unscoped         bool
	RequesterUnitIDs []string
}
