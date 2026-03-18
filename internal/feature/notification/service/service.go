package service

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"hrms/internal/feature/notification/repository"

	"github.com/google/uuid"
)

const (
	defaultLimit = 20
	maxLimit     = 100
)

type Service struct {
	repo repository.NotificationRepository
}

func NewService(repo repository.NotificationRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreatePayrollNotification(ctx context.Context, userIDRaw string, req CreateNotificationRequest) (*NotificationResponse, error) {
	return s.createNotification(ctx, userIDRaw, repository.TypePayroll, req)
}

func (s *Service) CreateSalaryNotification(ctx context.Context, userIDRaw string, req CreateNotificationRequest) (*NotificationResponse, error) {
	return s.createNotification(ctx, userIDRaw, repository.TypeSalary, req)
}

func (s *Service) CreateSystemNotification(ctx context.Context, userIDRaw string, req CreateNotificationRequest) (*NotificationResponse, error) {
	return s.createNotification(ctx, userIDRaw, repository.TypeSystem, req)
}

func (s *Service) ListNotifications(ctx context.Context, userIDRaw string, req ListNotificationsRequest) ([]NotificationResponse, error) {
	userID, err := parseUserID(userIDRaw)
	if err != nil {
		return nil, err
	}

	limit := req.Limit
	if limit == 0 {
		limit = defaultLimit
	}
	if limit < 1 || limit > maxLimit {
		return nil, ErrInvalidPaginationLimit
	}
	if req.Offset < 0 {
		return nil, ErrInvalidPaginationShift
	}

	notifications, err := s.repo.ListByUserID(ctx, repository.ListNotificationsParams{
		UserID:     userID,
		UnreadOnly: req.UnreadOnly,
		Limit:      limit,
		Offset:     req.Offset,
	})
	if err != nil {
		return nil, err
	}

	response := make([]NotificationResponse, 0, len(notifications))
	for _, notification := range notifications {
		response = append(response, toNotificationResponse(notification))
	}

	return response, nil
}

func (s *Service) MarkAsRead(ctx context.Context, userIDRaw, notificationIDRaw string) error {
	userID, err := parseUserID(userIDRaw)
	if err != nil {
		return err
	}

	notificationID, err := uuid.Parse(strings.TrimSpace(notificationIDRaw))
	if err != nil {
		return ErrInvalidNotificationID
	}

	updated, err := s.repo.MarkAsRead(ctx, notificationID, userID, time.Now().UTC())
	if err != nil {
		return err
	}
	if !updated {
		return ErrNotificationNotFound
	}

	return nil
}

func (s *Service) MarkAllAsRead(ctx context.Context, userIDRaw string) (*MarkAllAsReadResponse, error) {
	userID, err := parseUserID(userIDRaw)
	if err != nil {
		return nil, err
	}

	updated, err := s.repo.MarkAllAsRead(ctx, userID, time.Now().UTC())
	if err != nil {
		return nil, err
	}

	return &MarkAllAsReadResponse{Updated: updated}, nil
}

func (s *Service) createNotification(ctx context.Context, userIDRaw, notificationType string, req CreateNotificationRequest) (*NotificationResponse, error) {
	userID, err := parseUserID(userIDRaw)
	if err != nil {
		return nil, err
	}

	orgID, err := parseOptionalOrgID(req.OrgID)
	if err != nil {
		return nil, err
	}

	title := strings.TrimSpace(req.Title)
	if title == "" {
		return nil, ErrTitleRequired
	}

	message := strings.TrimSpace(req.Message)
	if message == "" {
		return nil, ErrMessageRequired
	}

	notification, err := s.repo.Create(ctx, repository.CreateNotificationParams{
		ID:       uuid.New(),
		UserID:   userID,
		OrgID:    orgID,
		Type:     notificationType,
		Title:    title,
		Message:  message,
		Metadata: normalizeMetadata(req.Metadata),
	})
	if err != nil {
		return nil, err
	}

	response := toNotificationResponse(notification)
	return &response, nil
}

func parseUserID(raw string) (uuid.UUID, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return uuid.Nil, ErrUserIDRequired
	}

	userID, err := uuid.Parse(trimmed)
	if err != nil {
		return uuid.Nil, ErrInvalidUserID
	}

	return userID, nil
}

func parseOptionalOrgID(raw *string) (*uuid.UUID, error) {
	if raw == nil {
		return nil, nil
	}

	trimmed := strings.TrimSpace(*raw)
	if trimmed == "" {
		return nil, nil
	}

	orgID, err := uuid.Parse(trimmed)
	if err != nil {
		return nil, ErrInvalidOrgID
	}

	return &orgID, nil
}

func normalizeMetadata(metadata json.RawMessage) json.RawMessage {
	trimmed := strings.TrimSpace(string(metadata))
	if trimmed == "" {
		return json.RawMessage(`{}`)
	}

	return json.RawMessage(trimmed)
}

func toNotificationResponse(notification repository.Notification) NotificationResponse {
	response := NotificationResponse{
		ID:        notification.ID.String(),
		UserID:    notification.UserID.String(),
		Type:      notification.Type,
		Title:     notification.Title,
		Message:   notification.Message,
		Metadata:  notification.Metadata,
		IsRead:    notification.IsRead,
		ReadAt:    notification.ReadAt,
		CreatedAt: notification.CreatedAt,
	}

	if notification.OrgID != nil {
		orgID := notification.OrgID.String()
		response.OrgID = &orgID
	}

	return response
}
