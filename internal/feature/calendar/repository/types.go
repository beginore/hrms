package repository

import (
	"time"

	"github.com/google/uuid"
)

type CalendarDay struct {
	ID        uuid.UUID
	OrgID     uuid.UUID
	Date      time.Time
	Type      string
	Name      string
	CreatedAt time.Time
}

type AddCalendarDayParams struct {
	Date time.Time
	Type string
	Name string
}
