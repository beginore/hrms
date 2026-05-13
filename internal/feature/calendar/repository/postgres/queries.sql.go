package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

const insertCalendarDay = `
INSERT INTO working_calendar (id, org_id, date, type, name)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (org_id, date) DO UPDATE
    SET type = EXCLUDED.type,
        name = EXCLUDED.name
`

type InsertCalendarDayParams struct {
	ID    uuid.UUID
	OrgID uuid.UUID
	Date  time.Time
	Type  string
	Name  string
}

func (q *Queries) InsertCalendarDay(ctx context.Context, arg InsertCalendarDayParams) error {
	var name sql.NullString
	if arg.Name != "" {
		name = sql.NullString{String: arg.Name, Valid: true}
	}
	_, err := q.db.ExecContext(ctx, insertCalendarDay,
		arg.ID, arg.OrgID, arg.Date, arg.Type, name)
	return err
}

const deleteCalendarDay = `DELETE FROM working_calendar WHERE id = $1 AND org_id = $2`

func (q *Queries) DeleteCalendarDay(ctx context.Context, id, orgID uuid.UUID) error {
	_, err := q.db.ExecContext(ctx, deleteCalendarDay, id, orgID)
	return err
}

const getCalendarDayByID = `
SELECT id, org_id, date, type, name, created_at FROM working_calendar WHERE id = $1
`

func (q *Queries) GetCalendarDayByID(ctx context.Context, id uuid.UUID) (CalendarDay, error) {
	row := q.db.QueryRowContext(ctx, getCalendarDayByID, id)
	return scanCalendarDay(row)
}

const getCalendarByOrgAndMonth = `
SELECT id, org_id, date, type, name, created_at
FROM working_calendar
WHERE org_id = $1 AND date >= $2 AND date <= $3
ORDER BY date
`

func (q *Queries) GetCalendarByOrgAndMonth(ctx context.Context, orgID uuid.UUID, from, to time.Time) ([]CalendarDay, error) {
	rows, err := q.db.QueryContext(ctx, getCalendarByOrgAndMonth, orgID, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []CalendarDay
	for rows.Next() {
		day, err := scanCalendarDay(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, day)
	}
	return items, rows.Err()
}

const checkCalendarDayExists = `SELECT COUNT(*) FROM working_calendar WHERE id = $1`

func (q *Queries) CheckCalendarDayExists(ctx context.Context, id uuid.UUID) (int64, error) {
	row := q.db.QueryRowContext(ctx, checkCalendarDayExists, id)
	var count int64
	return count, row.Scan(&count)
}

// GetNonWorkingDayType returns the type (HOLIDAY/WORKDAY) for a given org+date, or empty string if not found.
const getNonWorkingDayType = `
SELECT type FROM working_calendar WHERE org_id = $1 AND date = $2
`

func (q *Queries) GetCalendarEntryForDate(ctx context.Context, orgID uuid.UUID, date time.Time) (string, error) {
	row := q.db.QueryRowContext(ctx, getNonWorkingDayType, orgID, date)
	var t string
	err := row.Scan(&t)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return t, err
}

const getOrgIDByUserID = `SELECT org_id FROM users WHERE id = $1`

func (q *Queries) GetOrgIDByUserID(ctx context.Context, userID uuid.UUID) (uuid.UUID, error) {
	row := q.db.QueryRowContext(ctx, getOrgIDByUserID, userID)
	var orgID uuid.UUID
	return orgID, row.Scan(&orgID)
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanCalendarDay(row rowScanner) (CalendarDay, error) {
	var d CalendarDay
	err := row.Scan(&d.ID, &d.OrgID, &d.Date, &d.Type, &d.Name, &d.CreatedAt)
	return d, err
}
