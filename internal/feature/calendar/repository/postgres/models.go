package postgres

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type CalendarDay struct {
	ID        uuid.UUID      `json:"id"`
	OrgID     uuid.UUID      `json:"org_id"`
	Date      time.Time      `json:"date"`
	Type      string         `json:"type"`
	Name      sql.NullString `json:"name"`
	CreatedAt time.Time      `json:"created_at"`
}
