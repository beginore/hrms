package repository

import (
	"context"
	"database/sql"
	"fmt"

	taskPostgres "hrms/internal/feature/task/repository/postgres"

	"github.com/google/uuid"
)

type taskRepository struct {
	db      *sql.DB
	queries *taskPostgres.Queries
}

func NewRepository(conn *sql.DB) TaskRepository {
	return &taskRepository{db: conn, queries: taskPostgres.New(conn)}
}

func (r *taskRepository) CreateTask(ctx context.Context, callerUserID uuid.UUID, params CreateTaskParams) (uuid.UUID, error) {
	orgID, err := r.GetOrgIDByUserID(ctx, callerUserID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("resolve org: %w", err)
	}

	var desc sql.NullString
	if params.Description != "" {
		desc = sql.NullString{String: params.Description, Valid: true}
	}

	id := uuid.New()
	if err := r.queries.InsertTask(ctx, taskPostgres.InsertTaskParams{
		ID:          id,
		OrgID:       orgID,
		CreatedBy:   callerUserID,
		AssignedTo:  params.AssignedTo,
		Title:       params.Title,
		Description: desc,
		DueDate:     params.DueDate,
	}); err != nil {
		return uuid.Nil, err
	}
	return id, nil
}

func (r *taskRepository) GetTaskByID(ctx context.Context, id uuid.UUID) (Task, error) {
	t, err := r.queries.GetTaskByID(ctx, id)
	if err != nil {
		return Task{}, err
	}
	return mapTask(t), nil
}

func (r *taskRepository) ListTasksByOrg(ctx context.Context, callerUserID uuid.UUID) ([]Task, error) {
	orgID, err := r.GetOrgIDByUserID(ctx, callerUserID)
	if err != nil {
		return nil, fmt.Errorf("resolve org: %w", err)
	}
	rows, err := r.queries.ListTasksByOrg(ctx, orgID)
	if err != nil {
		return nil, err
	}
	return mapTasks(rows), nil
}

func (r *taskRepository) ListTasksByEmployee(ctx context.Context, employeeID uuid.UUID) ([]Task, error) {
	rows, err := r.queries.ListTasksByEmployee(ctx, employeeID)
	if err != nil {
		return nil, err
	}
	return mapTasks(rows), nil
}

func (r *taskRepository) SubmitReport(ctx context.Context, params SubmitReportParams) error {
	var desc, url sql.NullString
	if params.ReportDescription != "" {
		desc = sql.NullString{String: params.ReportDescription, Valid: true}
	}
	if params.ReportURL != "" {
		url = sql.NullString{String: params.ReportURL, Valid: true}
	}
	return r.queries.SubmitReport(ctx, taskPostgres.SubmitReportParams{
		ID:                params.ID,
		ReportDescription: desc,
		ReportURL:         url,
	})
}

func (r *taskRepository) ReviewTask(ctx context.Context, callerUserID uuid.UUID, params ReviewTaskParams) error {
	return r.queries.ReviewTask(ctx, taskPostgres.ReviewTaskParams{
		ID:         params.ID,
		Status:     params.Status,
		ReviewedBy: callerUserID,
	})
}

func (r *taskRepository) GetOrgIDByUserID(ctx context.Context, userID uuid.UUID) (uuid.UUID, error) {
	return r.queries.GetOrgIDByUserID(ctx, userID)
}

func (r *taskRepository) GetEmployeeIDByUserID(ctx context.Context, userID uuid.UUID) (uuid.UUID, error) {
	return r.queries.GetEmployeeIDByUserID(ctx, userID)
}

func (r *taskRepository) EmployeeExistsInOrg(ctx context.Context, employeeID, orgID uuid.UUID) (bool, error) {
	return r.queries.EmployeeExistsInOrg(ctx, employeeID, orgID)
}

// ─── Mappers ──────────────────────────────────────────────────────────────────

func mapTask(t taskPostgres.Task) Task {
	out := Task{
		ID:         t.ID,
		OrgID:      t.OrgID,
		CreatedBy:  t.CreatedBy,
		AssignedTo: t.AssignedTo,
		Title:      t.Title,
		DueDate:    t.DueDate,
		Status:     t.Status,
		CreatedAt:  t.CreatedAt,
	}
	if t.Description.Valid {
		out.Description = t.Description.String
	}
	if t.ReportDescription.Valid {
		out.ReportDescription = t.ReportDescription.String
	}
	if t.ReportURL.Valid {
		out.ReportURL = t.ReportURL.String
	}
	if t.SubmittedAt.Valid {
		ts := t.SubmittedAt.Time
		out.SubmittedAt = &ts
	}
	if t.ReviewedBy.Valid {
		id := t.ReviewedBy.UUID
		out.ReviewedBy = &id
	}
	if t.ReviewedAt.Valid {
		ts := t.ReviewedAt.Time
		out.ReviewedAt = &ts
	}
	return out
}

func mapTasks(rows []taskPostgres.Task) []Task {
	out := make([]Task, len(rows))
	for i, r := range rows {
		out[i] = mapTask(r)
	}
	return out
}
