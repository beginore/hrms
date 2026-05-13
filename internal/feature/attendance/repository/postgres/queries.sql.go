package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// ─── Work Schedules ──────────────────────────────────────────────────────────

const upsertWorkSchedule = `
INSERT INTO work_schedules (id, org_id, employee_id, work_start, work_end, late_threshold_minutes, updated_at)
VALUES ($1, $2, $3, $4::time, $5::time, $6, now())
ON CONFLICT (employee_id) DO UPDATE
    SET work_start = EXCLUDED.work_start,
        work_end   = EXCLUDED.work_end,
        late_threshold_minutes = EXCLUDED.late_threshold_minutes,
        updated_at = now()
`

func (q *Queries) UpsertWorkSchedule(ctx context.Context, arg UpsertWorkScheduleParams) error {
	_, err := q.db.ExecContext(ctx, upsertWorkSchedule,
		arg.ID, arg.OrgID, arg.EmployeeID,
		arg.WorkStart, arg.WorkEnd, arg.LateThresholdMinutes,
	)
	return err
}

const getWorkScheduleByEmployeeID = `
SELECT id, org_id, employee_id, work_start::text, work_end::text, late_threshold_minutes, created_at, updated_at
FROM work_schedules WHERE employee_id = $1
`

func (q *Queries) GetWorkScheduleByEmployeeID(ctx context.Context, employeeID uuid.UUID) (WorkSchedule, error) {
	row := q.db.QueryRowContext(ctx, getWorkScheduleByEmployeeID, employeeID)
	var ws WorkSchedule
	err := row.Scan(&ws.ID, &ws.OrgID, &ws.EmployeeID,
		&ws.WorkStart, &ws.WorkEnd, &ws.LateThresholdMinutes,
		&ws.CreatedAt, &ws.UpdatedAt)
	return ws, err
}

// ─── SKUD Events ─────────────────────────────────────────────────────────────

const insertSkudEvent = `
INSERT INTO skud_events (id, org_id, employee_id, event_type, device_id, occurred_at)
VALUES ($1, $2, $3, $4, $5, $6)
`

func (q *Queries) InsertSkudEvent(ctx context.Context, arg InsertSkudEventParams) error {
	var deviceID sql.NullString
	if arg.DeviceID != "" {
		deviceID = sql.NullString{String: arg.DeviceID, Valid: true}
	}
	_, err := q.db.ExecContext(ctx, insertSkudEvent,
		arg.ID, arg.OrgID, arg.EmployeeID, arg.EventType, deviceID, arg.OccurredAt)
	return err
}

const getUnprocessedSkudEvents = `
SELECT id, org_id, employee_id, event_type, device_id, occurred_at, processed, created_at
FROM skud_events WHERE org_id = $1 AND processed = false ORDER BY occurred_at
`

func (q *Queries) GetUnprocessedSkudEvents(ctx context.Context, orgID uuid.UUID) ([]SkudEvent, error) {
	rows, err := q.db.QueryContext(ctx, getUnprocessedSkudEvents, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []SkudEvent
	for rows.Next() {
		var e SkudEvent
		if err := rows.Scan(&e.ID, &e.OrgID, &e.EmployeeID, &e.EventType,
			&e.DeviceID, &e.OccurredAt, &e.Processed, &e.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, e)
	}
	return items, rows.Err()
}

const markSkudEventProcessed = `UPDATE skud_events SET processed = true WHERE id = $1`

func (q *Queries) MarkSkudEventProcessed(ctx context.Context, id uuid.UUID) error {
	_, err := q.db.ExecContext(ctx, markSkudEventProcessed, id)
	return err
}

// ─── Leave Requests ──────────────────────────────────────────────────────────

const insertLeaveRequest = `
INSERT INTO leave_requests (id, org_id, employee_id, type, start_date, end_date, reason, document_url)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
`

func (q *Queries) InsertLeaveRequest(ctx context.Context, arg InsertLeaveRequestParams) error {
	var reason, docURL sql.NullString
	if arg.Reason != "" {
		reason = sql.NullString{String: arg.Reason, Valid: true}
	}
	if arg.DocumentURL != "" {
		docURL = sql.NullString{String: arg.DocumentURL, Valid: true}
	}
	_, err := q.db.ExecContext(ctx, insertLeaveRequest,
		arg.ID, arg.OrgID, arg.EmployeeID, arg.Type,
		arg.StartDate, arg.EndDate, reason, docURL)
	return err
}

const getLeaveRequestByID = `
SELECT id, org_id, employee_id, type, start_date, end_date, reason, document_url,
       status, reviewed_by, reviewed_at, created_at, updated_at
FROM leave_requests WHERE id = $1
`

func (q *Queries) GetLeaveRequestByID(ctx context.Context, id uuid.UUID) (LeaveRequest, error) {
	row := q.db.QueryRowContext(ctx, getLeaveRequestByID, id)
	return scanLeaveRequest(row)
}

const getLeaveRequestsByOrgID = `
SELECT id, org_id, employee_id, type, start_date, end_date, reason, document_url,
       status, reviewed_by, reviewed_at, created_at, updated_at
FROM leave_requests WHERE org_id = $1 ORDER BY created_at DESC
`

func (q *Queries) GetLeaveRequestsByOrgID(ctx context.Context, orgID uuid.UUID) ([]LeaveRequest, error) {
	return queryLeaveRequests(ctx, q, getLeaveRequestsByOrgID, orgID)
}

const getLeaveRequestsByEmployeeID = `
SELECT id, org_id, employee_id, type, start_date, end_date, reason, document_url,
       status, reviewed_by, reviewed_at, created_at, updated_at
FROM leave_requests WHERE employee_id = $1 ORDER BY created_at DESC
`

func (q *Queries) GetLeaveRequestsByEmployeeID(ctx context.Context, employeeID uuid.UUID) ([]LeaveRequest, error) {
	return queryLeaveRequests(ctx, q, getLeaveRequestsByEmployeeID, employeeID)
}

const getApprovedLeaveForDate = `
SELECT id, org_id, employee_id, type, start_date, end_date, reason, document_url,
       status, reviewed_by, reviewed_at, created_at, updated_at
FROM leave_requests
WHERE employee_id = $1 AND status = 'APPROVED'
  AND start_date <= $2 AND end_date >= $2
LIMIT 1
`

func (q *Queries) GetApprovedLeaveForDate(ctx context.Context, arg GetApprovedLeaveForDateParams) (LeaveRequest, error) {
	row := q.db.QueryRowContext(ctx, getApprovedLeaveForDate, arg.EmployeeID, arg.Date)
	return scanLeaveRequest(row)
}

const updateLeaveRequestStatus = `
UPDATE leave_requests
SET status = $2, reviewed_by = $3, reviewed_at = now(), updated_at = now()
WHERE id = $1
`

func (q *Queries) UpdateLeaveRequestStatus(ctx context.Context, arg UpdateLeaveRequestStatusParams) error {
	_, err := q.db.ExecContext(ctx, updateLeaveRequestStatus, arg.ID, arg.Status, arg.ReviewedBy)
	return err
}

const checkLeaveRequestExists = `SELECT COUNT(*) FROM leave_requests WHERE id = $1`

func (q *Queries) CheckLeaveRequestExists(ctx context.Context, id uuid.UUID) (int64, error) {
	row := q.db.QueryRowContext(ctx, checkLeaveRequestExists, id)
	var count int64
	return count, row.Scan(&count)
}

// ─── Attendance Records ───────────────────────────────────────────────────────

const upsertAttendanceRecord = `
INSERT INTO attendance_records (id, org_id, employee_id, date, type, source, status, check_in, check_out, note, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, now())
ON CONFLICT (employee_id, date) DO UPDATE
    SET type      = EXCLUDED.type,
        source    = EXCLUDED.source,
        status    = EXCLUDED.status,
        check_in  = COALESCE(EXCLUDED.check_in,  attendance_records.check_in),
        check_out = COALESCE(EXCLUDED.check_out, attendance_records.check_out),
        note      = COALESCE(EXCLUDED.note,      attendance_records.note),
        updated_at = now()
`

func (q *Queries) UpsertAttendanceRecord(ctx context.Context, arg UpsertAttendanceRecordParams) error {
	var checkIn, checkOut sql.NullTime
	if arg.CheckIn != nil {
		checkIn = sql.NullTime{Time: *arg.CheckIn, Valid: true}
	}
	if arg.CheckOut != nil {
		checkOut = sql.NullTime{Time: *arg.CheckOut, Valid: true}
	}
	var note sql.NullString
	if arg.Note != "" {
		note = sql.NullString{String: arg.Note, Valid: true}
	}
	_, err := q.db.ExecContext(ctx, upsertAttendanceRecord,
		arg.ID, arg.OrgID, arg.EmployeeID, arg.Date,
		arg.Type, arg.Source, arg.Status, checkIn, checkOut, note)
	return err
}

const getAttendanceByOrgAndPeriod = `
SELECT id, org_id, employee_id, date, type, source, status, check_in, check_out, note, created_at, updated_at
FROM attendance_records
WHERE org_id = $1 AND date >= $2 AND date <= $3
ORDER BY date DESC, employee_id
`

func (q *Queries) GetAttendanceByOrgAndPeriod(ctx context.Context, arg GetAttendanceByOrgAndPeriodParams) ([]AttendanceRecord, error) {
	rows, err := q.db.QueryContext(ctx, getAttendanceByOrgAndPeriod, arg.OrgID, arg.StartDate, arg.EndDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanAttendanceRows(rows)
}

const getAttendanceByEmployeeAndPeriod = `
SELECT id, org_id, employee_id, date, type, source, status, check_in, check_out, note, created_at, updated_at
FROM attendance_records
WHERE employee_id = $1 AND date >= $2 AND date <= $3
ORDER BY date DESC
`

func (q *Queries) GetAttendanceByEmployeeAndPeriod(ctx context.Context, arg GetAttendanceByEmployeeAndPeriodParams) ([]AttendanceRecord, error) {
	rows, err := q.db.QueryContext(ctx, getAttendanceByEmployeeAndPeriod, arg.EmployeeID, arg.StartDate, arg.EndDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanAttendanceRows(rows)
}

const getAttendanceByEmployeeAndDate = `
SELECT id, org_id, employee_id, date, type, source, status, check_in, check_out, note, created_at, updated_at
FROM attendance_records WHERE employee_id = $1 AND date = $2
`

func (q *Queries) GetAttendanceByEmployeeAndDate(ctx context.Context, arg GetAttendanceByEmployeeAndDateParams) (AttendanceRecord, error) {
	row := q.db.QueryRowContext(ctx, getAttendanceByEmployeeAndDate, arg.EmployeeID, arg.Date)
	return scanAttendanceRow(row)
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

const getOrgIDByEmployeeID = `SELECT org_id FROM employees WHERE id = $1`

func (q *Queries) GetOrgIDByEmployeeID(ctx context.Context, employeeID uuid.UUID) (uuid.UUID, error) {
	row := q.db.QueryRowContext(ctx, getOrgIDByEmployeeID, employeeID)
	var orgID uuid.UUID
	return orgID, row.Scan(&orgID)
}

const getOrgIDByUserID = `SELECT org_id FROM users WHERE id = $1`

func (q *Queries) GetOrgIDByUserID(ctx context.Context, userID uuid.UUID) (uuid.UUID, error) {
	row := q.db.QueryRowContext(ctx, getOrgIDByUserID, userID)
	var orgID uuid.UUID
	return orgID, row.Scan(&orgID)
}

const getEmployeeIDByUserID = `SELECT id FROM employees WHERE user_id = $1`

func (q *Queries) GetEmployeeIDByUserID(ctx context.Context, userID uuid.UUID) (uuid.UUID, error) {
	row := q.db.QueryRowContext(ctx, getEmployeeIDByUserID, userID)
	var empID uuid.UUID
	return empID, row.Scan(&empID)
}

const checkEmployeeExists = `SELECT COUNT(*) FROM employees WHERE id = $1`

func (q *Queries) CheckEmployeeExists(ctx context.Context, employeeID uuid.UUID) (int64, error) {
	row := q.db.QueryRowContext(ctx, checkEmployeeExists, employeeID)
	var count int64
	return count, row.Scan(&count)
}

// ─── scan helpers ─────────────────────────────────────────────────────────────

type rowScanner interface {
	Scan(dest ...any) error
}

func scanLeaveRequest(row rowScanner) (LeaveRequest, error) {
	var lr LeaveRequest
	err := row.Scan(
		&lr.ID, &lr.OrgID, &lr.EmployeeID, &lr.Type,
		&lr.StartDate, &lr.EndDate, &lr.Reason, &lr.DocumentURL,
		&lr.Status, &lr.ReviewedBy, &lr.ReviewedAt,
		&lr.CreatedAt, &lr.UpdatedAt,
	)
	return lr, err
}

func queryLeaveRequests(ctx context.Context, q *Queries, query string, arg uuid.UUID) ([]LeaveRequest, error) {
	rows, err := q.db.QueryContext(ctx, query, arg)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []LeaveRequest
	for rows.Next() {
		lr, err := scanLeaveRequest(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, lr)
	}
	return items, rows.Err()
}

func scanAttendanceRow(row rowScanner) (AttendanceRecord, error) {
	var ar AttendanceRecord
	err := row.Scan(
		&ar.ID, &ar.OrgID, &ar.EmployeeID, &ar.Date,
		&ar.Type, &ar.Source, &ar.Status,
		&ar.CheckIn, &ar.CheckOut, &ar.Note,
		&ar.CreatedAt, &ar.UpdatedAt,
	)
	return ar, err
}

func scanAttendanceRows(rows *sql.Rows) ([]AttendanceRecord, error) {
	var items []AttendanceRecord
	for rows.Next() {
		ar, err := scanAttendanceRow(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, ar)
	}
	return items, rows.Err()
}

// ensure Queries implements Querier at compile time
var _ Querier = (*Queries)(nil)

// time import used by InsertSkudEventParams
var _ = time.Now
