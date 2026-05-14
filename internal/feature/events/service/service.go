package service

import (
	"context"
	"strings"
	"time"

	eventsRepository "hrms/internal/feature/events/repository"

	"github.com/google/uuid"
)

type Service struct {
	repo *eventsRepository.Repository
}

func NewService(repo *eventsRepository.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListUpcoming(ctx context.Context, userIDRaw string) ([]EventResponse, error) {
	actor, err := s.actor(ctx, userIDRaw)
	if err != nil {
		return nil, err
	}

	events, err := s.repo.ListUpcoming(ctx, actor, time.Now().UTC())
	if err != nil {
		return nil, err
	}

	return s.toResponses(actor, events), nil
}

func (s *Service) ListMy(ctx context.Context, userIDRaw string) ([]EventResponse, error) {
	actor, err := s.actor(ctx, userIDRaw)
	if err != nil {
		return nil, err
	}

	events, err := s.repo.ListMyEvents(ctx, actor, time.Now().UTC())
	if err != nil {
		return nil, err
	}

	return s.toResponses(actor, events), nil
}

func (s *Service) Create(ctx context.Context, userIDRaw string, req CreateEventRequest) (*EventResponse, error) {
	actor, err := s.actor(ctx, userIDRaw)
	if err != nil {
		return nil, err
	}

	title, description, startsAt, endsAt, departmentID, scope, err := validateCreateRequest(req)
	if err != nil {
		return nil, err
	}

	if err := validateCreatePermissions(actor, scope, departmentID); err != nil {
		return nil, err
	}

	event, err := s.repo.CreateEvent(ctx, eventsRepository.CreateEventParams{
		ID:             uuid.New(),
		Title:          title,
		Description:    description,
		StartsAt:       startsAt,
		EndsAt:         endsAt,
		Scope:          scope,
		DepartmentID:   departmentID,
		CreatedBy:      actor.UserID,
		CreatedByRole:  canonicalRole(actor.Role),
		OrganizationID: actor.Organization,
	})
	if err != nil {
		return nil, err
	}

	response := toResponse(actor, event)
	return &response, nil
}

func (s *Service) Update(ctx context.Context, userIDRaw, eventIDRaw string, req UpdateEventRequest) (*EventResponse, error) {
	actor, err := s.actor(ctx, userIDRaw)
	if err != nil {
		return nil, err
	}

	eventID, err := parseEventID(eventIDRaw)
	if err != nil {
		return nil, err
	}

	current, err := s.repo.GetEventByID(ctx, eventID)
	if err != nil {
		return nil, ErrEventNotFound
	}
	if !canManage(actor, current) {
		return nil, ErrPermissionDenied
	}

	title, description, startsAt, endsAt, departmentID, err := validateUpdateRequest(req, current.Scope)
	if err != nil {
		return nil, err
	}
	if current.Scope == eventsRepository.ScopeDepartment && actor.DepartmentID != nil && departmentID != nil && *departmentID != *actor.DepartmentID {
		return nil, ErrPermissionDenied
	}

	event, err := s.repo.UpdateEvent(ctx, eventsRepository.UpdateEventParams{
		ID:           eventID,
		Title:        title,
		Description:  description,
		StartsAt:     startsAt,
		EndsAt:       endsAt,
		DepartmentID: departmentID,
	})
	if err != nil {
		return nil, err
	}

	response := toResponse(actor, event)
	return &response, nil
}

func (s *Service) Delete(ctx context.Context, userIDRaw, eventIDRaw string) error {
	actor, err := s.actor(ctx, userIDRaw)
	if err != nil {
		return err
	}

	eventID, err := parseEventID(eventIDRaw)
	if err != nil {
		return err
	}

	current, err := s.repo.GetEventByID(ctx, eventID)
	if err != nil {
		return ErrEventNotFound
	}
	if !canManage(actor, current) {
		return ErrPermissionDenied
	}

	return s.repo.DeleteEvent(ctx, eventID)
}

func (s *Service) actor(ctx context.Context, userIDRaw string) (eventsRepository.ActorContext, error) {
	userID, err := uuid.Parse(strings.TrimSpace(userIDRaw))
	if err != nil {
		return eventsRepository.ActorContext{}, ErrPermissionDenied
	}
	actor, err := s.repo.GetActorContext(ctx, userID)
	if err != nil {
		return eventsRepository.ActorContext{}, err
	}
	return actor, nil
}

func validateCreateRequest(req CreateEventRequest) (string, string, time.Time, time.Time, *uuid.UUID, string, error) {
	title := strings.TrimSpace(req.Title)
	if title == "" {
		return "", "", time.Time{}, time.Time{}, nil, "", ErrTitleRequired
	}
	description := strings.TrimSpace(req.Description)
	if description == "" {
		return "", "", time.Time{}, time.Time{}, nil, "", ErrDescriptionRequired
	}
	scope := strings.ToLower(strings.TrimSpace(req.Scope))
	if scope == "" {
		return "", "", time.Time{}, time.Time{}, nil, "", ErrScopeRequired
	}
	if scope != eventsRepository.ScopeGlobal && scope != eventsRepository.ScopeDepartment {
		return "", "", time.Time{}, time.Time{}, nil, "", ErrInvalidScope
	}
	startsAt, endsAt, err := parseTimeRange(req.StartsAt, req.EndsAt)
	if err != nil {
		return "", "", time.Time{}, time.Time{}, nil, "", err
	}
	departmentID, err := validateDepartmentByScope(scope, req.DepartmentID)
	if err != nil {
		return "", "", time.Time{}, time.Time{}, nil, "", err
	}
	return title, description, startsAt, endsAt, departmentID, scope, nil
}

func validateUpdateRequest(req UpdateEventRequest, scope string) (string, string, time.Time, time.Time, *uuid.UUID, error) {
	title := strings.TrimSpace(req.Title)
	if title == "" {
		return "", "", time.Time{}, time.Time{}, nil, ErrTitleRequired
	}
	description := strings.TrimSpace(req.Description)
	if description == "" {
		return "", "", time.Time{}, time.Time{}, nil, ErrDescriptionRequired
	}
	startsAt, endsAt, err := parseTimeRange(req.StartsAt, req.EndsAt)
	if err != nil {
		return "", "", time.Time{}, time.Time{}, nil, err
	}
	departmentID, err := validateDepartmentByScope(scope, req.DepartmentID)
	if err != nil {
		return "", "", time.Time{}, time.Time{}, nil, err
	}
	return title, description, startsAt, endsAt, departmentID, nil
}

func validateCreatePermissions(actor eventsRepository.ActorContext, scope string, departmentID *uuid.UUID) error {
	switch actor.Role {
	case strings.ToLower(eventsRepository.RoleSuperAdmin):
		if scope != eventsRepository.ScopeGlobal {
			return ErrPermissionDenied
		}
		return nil
	case strings.ToLower(eventsRepository.RoleAdmin):
		if scope != eventsRepository.ScopeDepartment {
			return ErrPermissionDenied
		}
		if actor.DepartmentID == nil {
			return ErrDepartmentNotBound
		}
		if departmentID == nil || *departmentID != *actor.DepartmentID {
			return ErrPermissionDenied
		}
		return nil
	default:
		return ErrPermissionDenied
	}
}

func validateDepartmentByScope(scope string, raw *string) (*uuid.UUID, error) {
	switch scope {
	case eventsRepository.ScopeGlobal:
		if raw != nil && strings.TrimSpace(*raw) != "" {
			return nil, ErrDepartmentForbidden
		}
		return nil, nil
	case eventsRepository.ScopeDepartment:
		if raw == nil || strings.TrimSpace(*raw) == "" {
			return nil, ErrDepartmentRequired
		}
		parsed, err := uuid.Parse(strings.TrimSpace(*raw))
		if err != nil {
			return nil, ErrInvalidDepartmentID
		}
		return &parsed, nil
	default:
		return nil, ErrInvalidScope
	}
}

func parseTimeRange(startsAtRaw, endsAtRaw string) (time.Time, time.Time, error) {
	if strings.TrimSpace(startsAtRaw) == "" {
		return time.Time{}, time.Time{}, ErrStartsAtRequired
	}
	if strings.TrimSpace(endsAtRaw) == "" {
		return time.Time{}, time.Time{}, ErrEndsAtRequired
	}
	startsAt, err := time.Parse(time.RFC3339, startsAtRaw)
	if err != nil {
		return time.Time{}, time.Time{}, ErrInvalidStartsAt
	}
	endsAt, err := time.Parse(time.RFC3339, endsAtRaw)
	if err != nil {
		return time.Time{}, time.Time{}, ErrInvalidEndsAt
	}
	if !endsAt.After(startsAt) {
		return time.Time{}, time.Time{}, ErrInvalidEventTime
	}
	return startsAt, endsAt, nil
}

func parseEventID(raw string) (uuid.UUID, error) {
	id, err := uuid.Parse(strings.TrimSpace(raw))
	if err != nil {
		return uuid.Nil, ErrInvalidEventID
	}
	return id, nil
}

func toResponse(actor eventsRepository.ActorContext, event eventsRepository.Event) EventResponse {
	resp := EventResponse{
		ID:             event.ID.String(),
		Title:          event.Title,
		Description:    event.Description,
		StartsAt:       event.StartsAt,
		EndsAt:         event.EndsAt,
		Scope:          event.Scope,
		DepartmentName: event.DepartmentName,
		CreatedBy:      event.CreatedBy.String(),
		CreatedByRole:  canonicalRole(event.CreatedByRole),
		OrganizationID: event.OrganizationID.String(),
		CreatedAt:      event.CreatedAt,
		CanEdit:        canManage(actor, event),
		CanDelete:      canManage(actor, event),
	}
	if event.DepartmentID != nil {
		departmentID := event.DepartmentID.String()
		resp.DepartmentID = &departmentID
	}
	return resp
}

func (s *Service) toResponses(actor eventsRepository.ActorContext, events []eventsRepository.Event) []EventResponse {
	response := make([]EventResponse, 0, len(events))
	for _, event := range events {
		response = append(response, toResponse(actor, event))
	}
	return response
}

func canManage(actor eventsRepository.ActorContext, event eventsRepository.Event) bool {
	switch actor.Role {
	case strings.ToLower(eventsRepository.RoleSuperAdmin):
		return event.Scope == eventsRepository.ScopeGlobal &&
			event.CreatedBy == actor.UserID &&
			canonicalRole(event.CreatedByRole) == eventsRepository.RoleSuperAdmin
	case strings.ToLower(eventsRepository.RoleAdmin):
		if actor.DepartmentID == nil || event.DepartmentID == nil {
			return false
		}
		return event.Scope == eventsRepository.ScopeDepartment &&
			event.CreatedBy == actor.UserID &&
			*event.DepartmentID == *actor.DepartmentID
	default:
		return false
	}
}

func canonicalRole(role string) string {
	switch strings.ToLower(strings.TrimSpace(role)) {
	case "sysadmin", "superadmin":
		return eventsRepository.RoleSuperAdmin
	case "admin":
		return eventsRepository.RoleAdmin
	case "employee":
		return eventsRepository.RoleEmployee
	default:
		return strings.TrimSpace(role)
	}
}
