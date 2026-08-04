package model

import "time"

type AuditLog struct {
	ID           string         `json:"id"`
	ActorUserID  *string        `json:"actor_user_id,omitempty"`
	ActorName    *string        `json:"actor_name,omitempty"`
	Action       string         `json:"action"`
	ResourceType string         `json:"resource_type"`
	ResourceID   string         `json:"resource_id"`
	UnitID       *string        `json:"unit_id,omitempty"`
	Metadata     map[string]any `json:"metadata"`
	CreatedAt    time.Time      `json:"created_at"`
}

type CursorPage struct {
	Items      []AuditLog `json:"items"`
	NextCursor *string    `json:"next_cursor"`
}

type ListFilter struct {
	Action       string
	ResourceType string
	ActorUserID  string
	UnitID       string
	Cursor       string
	Limit        int
}
