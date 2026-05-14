package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	ScopeGlobal     = "global"
	ScopeDepartment = "department"
	RoleSuperAdmin  = "SuperAdmin"
	RoleAdmin       = "Admin"
	RoleEmployee    = "Employee"
)

var ErrActorNotFound = errors.New("actor not found")

type ActorContext struct {
	UserID       uuid.UUID
	Organization uuid.UUID
	Role         string
	DepartmentID *uuid.UUID
}

type Event struct {
	ID             uuid.UUID
	Title          string
	Description    string
	StartsAt       time.Time
	EndsAt         time.Time
	Scope          string
	DepartmentID   *uuid.UUID
	DepartmentName *string
	CreatedBy      uuid.UUID
	CreatedByRole  string
	OrganizationID uuid.UUID
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type CreateEventParams struct {
	ID             uuid.UUID
	Title          string
	Description    string
	StartsAt       time.Time
	EndsAt         time.Time
	Scope          string
	DepartmentID   *uuid.UUID
	CreatedBy      uuid.UUID
	CreatedByRole  string
	OrganizationID uuid.UUID
}

type UpdateEventParams struct {
	ID           uuid.UUID
	Title        string
	Description  string
	StartsAt     time.Time
	EndsAt       time.Time
	DepartmentID *uuid.UUID
}

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetActorContext(ctx context.Context, userID uuid.UUID) (ActorContext, error) {
	const query = `
SELECT
    u.id,
    u.org_id,
    COALESCE(NULLIF(TRIM(e.role), ''), NULLIF(TRIM(u.role), '')) AS role,
    e.department_id
FROM users u
LEFT JOIN employees e ON e.user_id = u.id
WHERE u.id = $1
`

	var actor ActorContext
	var role sql.NullString
	var departmentID sql.NullString
	if err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&actor.UserID,
		&actor.Organization,
		&role,
		&departmentID,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ActorContext{}, ErrActorNotFound
		}
		return ActorContext{}, err
	}

	actor.Role = normalizeRole(role.String)
	if departmentID.Valid && strings.TrimSpace(departmentID.String) != "" {
		parsed, err := uuid.Parse(departmentID.String)
		if err != nil {
			return ActorContext{}, fmt.Errorf("parse actor department id: %w", err)
		}
		actor.DepartmentID = &parsed
	}

	return actor, nil
}

func (r *Repository) CreateEvent(ctx context.Context, params CreateEventParams) (Event, error) {
	const query = `
INSERT INTO events (
    id, title, description, starts_at, ends_at, scope, department_id, created_by, created_by_role, organization_id
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING id, title, description, starts_at, ends_at, scope, department_id, NULL::text AS department_name, created_by, created_by_role, organization_id, created_at, updated_at
`

	event, err := scanEvent(r.db.QueryRowContext(
		ctx,
		query,
		params.ID,
		params.Title,
		params.Description,
		params.StartsAt,
		params.EndsAt,
		params.Scope,
		params.DepartmentID,
		params.CreatedBy,
		params.CreatedByRole,
		params.OrganizationID,
	))
	if err != nil {
		return Event{}, err
	}

	if event.DepartmentID != nil {
		name, nameErr := r.getDepartmentName(ctx, *event.DepartmentID)
		if nameErr != nil {
			return Event{}, nameErr
		}
		event.DepartmentName = &name
	}

	return event, nil
}

func (r *Repository) GetEventByID(ctx context.Context, id uuid.UUID) (Event, error) {
	const query = `
SELECT
    e.id, e.title, e.description, e.starts_at, e.ends_at, e.scope, e.department_id,
    d.name, e.created_by, e.created_by_role, e.organization_id, e.created_at, e.updated_at
FROM events e
LEFT JOIN departments d ON d.id = e.department_id
WHERE e.id = $1
`

	return scanEvent(r.db.QueryRowContext(ctx, query, id))
}

func (r *Repository) UpdateEvent(ctx context.Context, params UpdateEventParams) (Event, error) {
	const query = `
UPDATE events
SET title = $2,
    description = $3,
    starts_at = $4,
    ends_at = $5,
    department_id = $6,
    updated_at = now()
WHERE id = $1
RETURNING id, title, description, starts_at, ends_at, scope, department_id, NULL::text AS department_name, created_by, created_by_role, organization_id, created_at, updated_at
`

	event, err := scanEvent(r.db.QueryRowContext(
		ctx,
		query,
		params.ID,
		params.Title,
		params.Description,
		params.StartsAt,
		params.EndsAt,
		params.DepartmentID,
	))
	if err != nil {
		return Event{}, err
	}

	if event.DepartmentID != nil {
		name, nameErr := r.getDepartmentName(ctx, *event.DepartmentID)
		if nameErr != nil {
			return Event{}, nameErr
		}
		event.DepartmentName = &name
	}

	return event, nil
}

func (r *Repository) DeleteEvent(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM events WHERE id = $1`, id)
	return err
}

func (r *Repository) ListUpcoming(ctx context.Context, actor ActorContext, now time.Time) ([]Event, error) {
	var (
		query string
		args  []any
	)

	switch actor.Role {
	case normalizeRole(RoleSuperAdmin):
		query = `
SELECT
    e.id, e.title, e.description, e.starts_at, e.ends_at, e.scope, e.department_id,
    d.name, e.created_by, e.created_by_role, e.organization_id, e.created_at, e.updated_at
FROM events e
LEFT JOIN departments d ON d.id = e.department_id
WHERE e.organization_id = $1
  AND e.scope = 'global'
  AND e.starts_at >= $2
ORDER BY e.starts_at ASC
`
		args = []any{actor.Organization, now}
	default:
		query = `
SELECT
    e.id, e.title, e.description, e.starts_at, e.ends_at, e.scope, e.department_id,
    d.name, e.created_by, e.created_by_role, e.organization_id, e.created_at, e.updated_at
FROM events e
LEFT JOIN departments d ON d.id = e.department_id
WHERE e.organization_id = $1
  AND e.starts_at >= $2
  AND (
      (e.scope = 'global' AND LOWER(e.created_by_role) = LOWER($3))
      OR
      (e.scope = 'department' AND e.department_id = $4)
  )
ORDER BY e.starts_at ASC
`
		args = []any{actor.Organization, now, RoleSuperAdmin, actor.DepartmentID}
	}

	return r.listEvents(ctx, query, args...)
}

func (r *Repository) ListMyEvents(ctx context.Context, actor ActorContext, now time.Time) ([]Event, error) {
	const query = `
SELECT
    e.id, e.title, e.description, e.starts_at, e.ends_at, e.scope, e.department_id,
    d.name, e.created_by, e.created_by_role, e.organization_id, e.created_at, e.updated_at
FROM events e
LEFT JOIN departments d ON d.id = e.department_id
WHERE e.organization_id = $1
  AND e.created_by = $2
  AND e.starts_at >= $3
ORDER BY e.starts_at ASC
`

	return r.listEvents(ctx, query, actor.Organization, actor.UserID, now)
}

func (r *Repository) getDepartmentName(ctx context.Context, departmentID uuid.UUID) (string, error) {
	var name string
	if err := r.db.QueryRowContext(ctx, `SELECT name FROM departments WHERE id = $1`, departmentID).Scan(&name); err != nil {
		return "", err
	}
	return name, nil
}

func (r *Repository) listEvents(ctx context.Context, query string, args ...any) ([]Event, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []Event
	for rows.Next() {
		event, scanErr := scanEvent(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return events, nil
}

func scanEvent(scanner interface{ Scan(dest ...any) error }) (Event, error) {
	var event Event
	var departmentID sql.NullString
	var departmentName sql.NullString

	err := scanner.Scan(
		&event.ID,
		&event.Title,
		&event.Description,
		&event.StartsAt,
		&event.EndsAt,
		&event.Scope,
		&departmentID,
		&departmentName,
		&event.CreatedBy,
		&event.CreatedByRole,
		&event.OrganizationID,
		&event.CreatedAt,
		&event.UpdatedAt,
	)
	if err != nil {
		return Event{}, err
	}

	if departmentID.Valid && strings.TrimSpace(departmentID.String) != "" {
		parsed, err := uuid.Parse(departmentID.String)
		if err != nil {
			return Event{}, fmt.Errorf("parse event department id: %w", err)
		}
		event.DepartmentID = &parsed
	}

	if departmentName.Valid {
		name := departmentName.String
		event.DepartmentName = &name
	}

	return event, nil
}

func normalizeRole(role string) string {
	switch strings.ToLower(strings.TrimSpace(role)) {
	case "sysadmin", "superadmin":
		return strings.ToLower(RoleSuperAdmin)
	case "admin":
		return strings.ToLower(RoleAdmin)
	case "employee":
		return strings.ToLower(RoleEmployee)
	default:
		return strings.ToLower(strings.TrimSpace(role))
	}
}
