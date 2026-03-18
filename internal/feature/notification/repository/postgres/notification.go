package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"hrms/internal/feature/notification/repository"

	"github.com/google/uuid"
)

type notificationRepository struct {
	db *sql.DB
}

func NewNotificationRepository(db *sql.DB) repository.NotificationRepository {
	return &notificationRepository{db: db}
}

func (r *notificationRepository) Create(ctx context.Context, params repository.CreateNotificationParams) (repository.Notification, error) {
	const query = `
INSERT INTO notifications (
    id,
    user_id,
    org_id,
    type,
    title,
    message,
    metadata
)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, user_id, org_id, type, title, message, metadata, is_read, read_at, created_at
`

	row := r.db.QueryRowContext(
		ctx,
		query,
		params.ID,
		params.UserID,
		params.OrgID,
		params.Type,
		params.Title,
		params.Message,
		normalizeMetadata(params.Metadata),
	)

	return scanNotification(row)
}

func (r *notificationRepository) ListByUserID(ctx context.Context, params repository.ListNotificationsParams) ([]repository.Notification, error) {
	baseQuery := `
SELECT id, user_id, org_id, type, title, message, metadata, is_read, read_at, created_at
FROM notifications
WHERE user_id = $1
`

	args := []any{params.UserID}
	if params.UnreadOnly {
		baseQuery += " AND is_read = false"
	}

	baseQuery += `
ORDER BY created_at DESC
LIMIT $2 OFFSET $3
`
	args = append(args, params.Limit, params.Offset)

	rows, err := r.db.QueryContext(ctx, baseQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("query notifications: %w", err)
	}
	defer rows.Close()

	notifications := make([]repository.Notification, 0, params.Limit)
	for rows.Next() {
		notification, scanErr := scanNotification(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		notifications = append(notifications, notification)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate notifications: %w", err)
	}

	return notifications, nil
}

func (r *notificationRepository) MarkAsRead(ctx context.Context, notificationID, userID uuid.UUID, readAt time.Time) (bool, error) {
	const query = `
UPDATE notifications
SET is_read = true, read_at = $3
WHERE id = $1 AND user_id = $2 AND is_read = false
`

	result, err := r.db.ExecContext(ctx, query, notificationID, userID, readAt)
	if err != nil {
		return false, fmt.Errorf("mark notification as read: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("mark notification rows affected: %w", err)
	}

	return rowsAffected > 0, nil
}

func (r *notificationRepository) MarkAllAsRead(ctx context.Context, userID uuid.UUID, readAt time.Time) (int64, error) {
	const query = `
UPDATE notifications
SET is_read = true, read_at = $2
WHERE user_id = $1 AND is_read = false
`

	result, err := r.db.ExecContext(ctx, query, userID, readAt)
	if err != nil {
		return 0, fmt.Errorf("mark all notifications as read: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("mark all notifications rows affected: %w", err)
	}

	return rowsAffected, nil
}

func normalizeMetadata(metadata json.RawMessage) []byte {
	trimmed := strings.TrimSpace(string(metadata))
	if trimmed == "" {
		return []byte("{}")
	}

	return []byte(trimmed)
}

func scanNotification(scanner interface {
	Scan(dest ...any) error
}) (repository.Notification, error) {
	var notification repository.Notification
	var orgID sql.NullString
	var metadata []byte

	err := scanner.Scan(
		&notification.ID,
		&notification.UserID,
		&orgID,
		&notification.Type,
		&notification.Title,
		&notification.Message,
		&metadata,
		&notification.IsRead,
		&notification.ReadAt,
		&notification.CreatedAt,
	)
	if err != nil {
		return repository.Notification{}, fmt.Errorf("scan notification: %w", err)
	}

	if orgID.Valid {
		parsed, parseErr := uuid.Parse(orgID.String)
		if parseErr != nil {
			return repository.Notification{}, fmt.Errorf("parse notification org id: %w", parseErr)
		}
		notification.OrgID = &parsed
	}

	if len(metadata) == 0 {
		notification.Metadata = json.RawMessage(`{}`)
	} else {
		notification.Metadata = json.RawMessage(metadata)
	}

	return notification, nil
}
