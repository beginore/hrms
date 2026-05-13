package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type CalendarRepository interface {
	AddDay(ctx context.Context, callerUserID uuid.UUID, params AddCalendarDayParams) (uuid.UUID, error)
	DeleteDay(ctx context.Context, callerUserID uuid.UUID, id uuid.UUID) error
	GetByMonth(ctx context.Context, callerUserID uuid.UUID, year int, month int) ([]CalendarDay, error)

	// IsNonWorkingDay returns true when the given date should NOT be treated as a regular workday.
	// It respects: explicit HOLIDAY entries, weekend days (Sat/Sun), and WORKDAY overrides.
	IsNonWorkingDay(ctx context.Context, orgID uuid.UUID, date time.Time) (bool, error)

	DayExists(ctx context.Context, id uuid.UUID) (bool, error)
	GetOrgIDByUserID(ctx context.Context, userID uuid.UUID) (uuid.UUID, error)
}
