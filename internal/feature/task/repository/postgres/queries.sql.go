package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// ─── Insert Task ──────────────────────────────────────────────────────────────

const insertTask = `
INSERT INTO tasks (id, org_id, created_by, assigned_to, title, description, due_date, status)
VALUES ($1, $2, $3, $4, $5, $6, $7, 'PENDING')
`

type InsertTaskParams struct {
	ID          uuid.UUID
	OrgID       uuid.UUID
	CreatedBy   uuid.UUID
	AssignedTo  uuid.UUID
	Title       string
	Description sql.NullString
	DueDate     time.Time
}

func (q *Queries) InsertTask(ctx context.Context, arg InsertTaskParams) error {
	_, err := q.db.ExecContext(ctx, insertTask,
		arg.ID, arg.OrgID, arg.CreatedBy, arg.AssignedTo,
		arg.Title, arg.Description, arg.DueDate,
	)
	return err
}

// ─── Get Task By ID ───────────────────────────────────────────────────────────

const getTaskByID = `
SELECT id, org_id, created_by, assigned_to, title, description, due_date,
       status, report_description, report_url, submitted_at,
       reviewed_by, reviewed_at, created_at, updated_at
FROM tasks WHERE id = $1
`

func (q *Queries) GetTaskByID(ctx context.Context, id uuid.UUID) (Task, error) {
	row := q.db.QueryRowContext(ctx, getTaskByID, id)
	return scanTask(row)
}

// ─── List Tasks By Org ────────────────────────────────────────────────────────

const listTasksByOrg = `
SELECT id, org_id, created_by, assigned_to, title, description, due_date,
       status, report_description, report_url, submitted_at,
       reviewed_by, reviewed_at, created_at, updated_at
FROM tasks WHERE org_id = $1 ORDER BY created_at DESC
`

func (q *Queries) ListTasksByOrg(ctx context.Context, orgID uuid.UUID) ([]Task, error) {
	rows, err := q.db.QueryContext(ctx, listTasksByOrg, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanTasks(rows)
}

// ─── List Tasks By Employee ───────────────────────────────────────────────────

const listTasksByEmployee = `
SELECT id, org_id, created_by, assigned_to, title, description, due_date,
       status, report_description, report_url, submitted_at,
       reviewed_by, reviewed_at, created_at, updated_at
FROM tasks WHERE assigned_to = $1 ORDER BY created_at DESC
`

func (q *Queries) ListTasksByEmployee(ctx context.Context, employeeID uuid.UUID) ([]Task, error) {
	rows, err := q.db.QueryContext(ctx, listTasksByEmployee, employeeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanTasks(rows)
}

// ─── Submit Report ────────────────────────────────────────────────────────────

const submitReport = `
UPDATE tasks
SET report_description = $2,
    report_url         = $3,
    submitted_at       = now(),
    status             = 'SUBMITTED',
    updated_at         = now()
WHERE id = $1
`

type SubmitReportParams struct {
	ID                uuid.UUID
	ReportDescription sql.NullString
	ReportURL         sql.NullString
}

func (q *Queries) SubmitReport(ctx context.Context, arg SubmitReportParams) error {
	_, err := q.db.ExecContext(ctx, submitReport, arg.ID, arg.ReportDescription, arg.ReportURL)
	return err
}

// ─── Review Task ──────────────────────────────────────────────────────────────

const reviewTask = `
UPDATE tasks SET status = $1, reviewed_by = $2, reviewed_at = now(), updated_at = now()
WHERE id = $3
`

type ReviewTaskParams struct {
	ID         uuid.UUID
	Status     string
	ReviewedBy uuid.UUID
}

func (q *Queries) ReviewTask(ctx context.Context, arg ReviewTaskParams) error {
	_, err := q.db.ExecContext(ctx, reviewTask, arg.Status, arg.ReviewedBy, arg.ID)
	return err
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

const getOrgIDByUserID = `SELECT org_id FROM users WHERE id = $1`

func (q *Queries) GetOrgIDByUserID(ctx context.Context, userID uuid.UUID) (uuid.UUID, error) {
	var orgID uuid.UUID
	err := q.db.QueryRowContext(ctx, getOrgIDByUserID, userID).Scan(&orgID)
	return orgID, err
}

const getEmployeeIDByUserID = `SELECT id FROM employees WHERE user_id = $1`

func (q *Queries) GetEmployeeIDByUserID(ctx context.Context, userID uuid.UUID) (uuid.UUID, error) {
	var id uuid.UUID
	err := q.db.QueryRowContext(ctx, getEmployeeIDByUserID, userID).Scan(&id)
	return id, err
}

const employeeExistsInOrg = `
SELECT EXISTS(SELECT 1 FROM employees WHERE id = $1 AND org_id = $2)
`

func (q *Queries) EmployeeExistsInOrg(ctx context.Context, employeeID, orgID uuid.UUID) (bool, error) {
	var exists bool
	err := q.db.QueryRowContext(ctx, employeeExistsInOrg, employeeID, orgID).Scan(&exists)
	return exists, err
}

// ─── Scan helpers ─────────────────────────────────────────────────────────────

type rowScanner interface {
	Scan(dest ...any) error
}

func scanTask(row rowScanner) (Task, error) {
	var t Task
	err := row.Scan(
		&t.ID, &t.OrgID, &t.CreatedBy, &t.AssignedTo,
		&t.Title, &t.Description, &t.DueDate,
		&t.Status, &t.ReportDescription, &t.ReportURL, &t.SubmittedAt,
		&t.ReviewedBy, &t.ReviewedAt,
		&t.CreatedAt, &t.UpdatedAt,
	)
	return t, err
}

func scanTasks(rows *sql.Rows) ([]Task, error) {
	var items []Task
	for rows.Next() {
		var t Task
		if err := rows.Scan(
			&t.ID, &t.OrgID, &t.CreatedBy, &t.AssignedTo,
			&t.Title, &t.Description, &t.DueDate,
			&t.Status, &t.ReportDescription, &t.ReportURL, &t.SubmittedAt,
			&t.ReviewedBy, &t.ReviewedAt,
			&t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, t)
	}
	return items, rows.Err()
}
