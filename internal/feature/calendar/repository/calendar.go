package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	calendarPostgres "hrms/internal/feature/calendar/repository/postgres"

	"github.com/google/uuid"
)

type calendarRepository struct {
	queries *calendarPostgres.Queries
}

func NewRepository(conn *sql.DB) CalendarRepository {
	return &calendarRepository{queries: calendarPostgres.New(conn)}
}

func (r *calendarRepository) AddDay(ctx context.Context, callerUserID uuid.UUID, params AddCalendarDayParams) (uuid.UUID, error) {
	orgID, err := r.GetOrgIDByUserID(ctx, callerUserID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("resolve org: %w", err)
	}
	id := uuid.New()
	if err := r.queries.InsertCalendarDay(ctx, calendarPostgres.InsertCalendarDayParams{
		ID:    id,
		OrgID: orgID,
		Date:  params.Date,
		Type:  params.Type,
		Name:  params.Name,
	}); err != nil {
		return uuid.Nil, err
	}
	return id, nil
}

func (r *calendarRepository) DeleteDay(ctx context.Context, callerUserID uuid.UUID, id uuid.UUID) error {
	orgID, err := r.GetOrgIDByUserID(ctx, callerUserID)
	if err != nil {
		return fmt.Errorf("resolve org: %w", err)
	}
	return r.queries.DeleteCalendarDay(ctx, id, orgID)
}

func (r *calendarRepository) GetByMonth(ctx context.Context, callerUserID uuid.UUID, year int, month int) ([]CalendarDay, error) {
	orgID, err := r.GetOrgIDByUserID(ctx, callerUserID)
	if err != nil {
		return nil, fmt.Errorf("resolve org: %w", err)
	}

	from := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	to := from.AddDate(0, 1, -1)

	rows, err := r.queries.GetCalendarByOrgAndMonth(ctx, orgID, from, to)
	if err != nil {
		return nil, err
	}

	out := make([]CalendarDay, len(rows))
	for i, row := range rows {
		out[i] = CalendarDay{
			ID:        row.ID,
			OrgID:     row.OrgID,
			Date:      row.Date,
			Type:      row.Type,
			CreatedAt: row.CreatedAt,
		}
		if row.Name.Valid {
			out[i].Name = row.Name.String
		}
	}
	return out, nil
}

func (r *calendarRepository) IsNonWorkingDay(ctx context.Context, orgID uuid.UUID, date time.Time) (bool, error) {
	// Check explicit calendar entry first
	entryType, err := r.queries.GetCalendarEntryForDate(ctx, orgID, date.Truncate(24*time.Hour))
	if err != nil {
		return false, err
	}

	switch entryType {
	case "HOLIDAY":
		// Explicitly marked as holiday
		return true, nil
	case "WORKDAY":
		// Explicitly marked as working day (weekend override)
		return false, nil
	default:
		// No explicit entry → fall back to weekend rule (Sat=6, Sun=0)
		weekday := date.Weekday()
		return weekday == time.Saturday || weekday == time.Sunday, nil
	}
}

func (r *calendarRepository) DayExists(ctx context.Context, id uuid.UUID) (bool, error) {
	count, err := r.queries.CheckCalendarDayExists(ctx, id)
	return count > 0, err
}

func (r *calendarRepository) GetOrgIDByUserID(ctx context.Context, userID uuid.UUID) (uuid.UUID, error) {
	return r.queries.GetOrgIDByUserID(ctx, userID)
}
