package model

import "time"

type SocialEngagement struct {
	ID               string     `json:"id"`
	CustomerID       *string    `json:"customer_id,omitempty"`
	CustomerName     *string    `json:"customer_name,omitempty"`
	Channel          string     `json:"channel"`
	Type             string     `json:"type"`
	Status           string     `json:"status"`
	UnitID           string     `json:"unit_id"`
	MediaID          *string    `json:"media_id,omitempty"`
	MediaURL         *string    `json:"media_url,omitempty"`
	MediaCaption     *string    `json:"media_caption,omitempty"`
	Body             string     `json:"body"`
	ExternalID       string     `json:"external_id"`
	AuthorExternalID string     `json:"author_external_id"`
	ConversationID   *string    `json:"conversation_id,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	RepliedAt        *time.Time `json:"replied_at,omitempty"`
}

type CursorPage struct {
	Items      []SocialEngagement `json:"items"`
	NextCursor *string            `json:"next_cursor"`
}

type ListFilter struct {
	UnitID           string
	Channel          string
	Type             string
	Status           string
	Cursor           string
	Limit            int
	Unscoped         bool
	RequesterUnitIDs []string
}

type ReplyRequest struct {
	Body string `json:"body" validate:"required"`
}
