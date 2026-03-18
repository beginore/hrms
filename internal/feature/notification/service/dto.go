package service

import (
	"encoding/json"
	"time"
)

type CreateNotificationRequest struct {
	OrgID    *string         `json:"org_id,omitempty"`
	Title    string          `json:"title"`
	Message  string          `json:"message"`
	Metadata json.RawMessage `json:"metadata,omitempty"`
}

type NotificationResponse struct {
	ID        string          `json:"id"`
	UserID    string          `json:"user_id"`
	OrgID     *string         `json:"org_id,omitempty"`
	Type      string          `json:"type"`
	Title     string          `json:"title"`
	Message   string          `json:"message"`
	Metadata  json.RawMessage `json:"metadata"`
	IsRead    bool            `json:"is_read"`
	ReadAt    *time.Time      `json:"read_at,omitempty"`
	CreatedAt time.Time       `json:"created_at"`
}

type ListNotificationsRequest struct {
	UnreadOnly bool `form:"unread_only"`
	Limit      int  `form:"limit"`
	Offset     int  `form:"offset"`
}

type MarkAsReadRequest struct{}

type MarkAllAsReadRequest struct{}

type MarkAllAsReadResponse struct {
	Updated int64 `json:"updated"`
}
