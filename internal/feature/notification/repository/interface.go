package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

const (
	TypePayroll = "payroll"
	TypeSalary  = "salary"
	TypeSystem  = "system"
)

type Notification struct {
	ID        uuid.UUID       `json:"id"`
	UserID    uuid.UUID       `json:"user_id"`
	OrgID     *uuid.UUID      `json:"org_id,omitempty"`
	Type      string          `json:"type"`
	Title     string          `json:"title"`
	Message   string          `json:"message"`
	Metadata  json.RawMessage `json:"metadata"`
	IsRead    bool            `json:"is_read"`
	ReadAt    *time.Time      `json:"read_at,omitempty"`
	CreatedAt time.Time       `json:"created_at"`
}

type CreateNotificationParams struct {
	ID       uuid.UUID
	UserID   uuid.UUID
	OrgID    *uuid.UUID
	Type     string
	Title    string
	Message  string
	Metadata json.RawMessage
}

type ListNotificationsParams struct {
	UserID     uuid.UUID
	UnreadOnly bool
	Limit      int
	Offset     int
}

type NotificationRepository interface {
	Create(ctx context.Context, params CreateNotificationParams) (Notification, error)
	CreateBulk(ctx context.Context, params []CreateNotificationParams) error
	ListUserIDsByRole(ctx context.Context, role string) ([]uuid.UUID, error)
	ListUserIDsByOrgAndRole(ctx context.Context, orgID uuid.UUID, role string) ([]uuid.UUID, error)
	ListByUserID(ctx context.Context, params ListNotificationsParams) ([]Notification, error)
	MarkAsRead(ctx context.Context, notificationID, userID uuid.UUID, readAt time.Time) (bool, error)
	MarkAllAsRead(ctx context.Context, userID uuid.UUID, readAt time.Time) (int64, error)
}
