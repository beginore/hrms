package postgres

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type WorkSchedule struct {
	ID                   uuid.UUID `json:"id"`
	OrgID                uuid.UUID `json:"org_id"`
	EmployeeID           uuid.UUID `json:"employee_id"`
	WorkStart            string    `json:"work_start"`
	WorkEnd              string    `json:"work_end"`
	LateThresholdMinutes int32     `json:"late_threshold_minutes"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

type SkudEvent struct {
	ID         uuid.UUID      `json:"id"`
	OrgID      uuid.UUID      `json:"org_id"`
	EmployeeID uuid.UUID      `json:"employee_id"`
	EventType  string         `json:"event_type"`
	DeviceID   sql.NullString `json:"device_id"`
	OccurredAt time.Time      `json:"occurred_at"`
	Processed  bool           `json:"processed"`
	CreatedAt  time.Time      `json:"created_at"`
}

type LeaveRequest struct {
	ID          uuid.UUID      `json:"id"`
	OrgID       uuid.UUID      `json:"org_id"`
	EmployeeID  uuid.UUID      `json:"employee_id"`
	Type        string         `json:"type"`
	StartDate   time.Time      `json:"start_date"`
	EndDate     time.Time      `json:"end_date"`
	Reason      sql.NullString `json:"reason"`
	DocumentURL sql.NullString `json:"document_url"`
	Status      string         `json:"status"`
	ReviewedBy  uuid.NullUUID  `json:"reviewed_by"`
	ReviewedAt  sql.NullTime   `json:"reviewed_at"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

type AttendanceRecord struct {
	ID         uuid.UUID      `json:"id"`
	OrgID      uuid.UUID      `json:"org_id"`
	EmployeeID uuid.UUID      `json:"employee_id"`
	Date       time.Time      `json:"date"`
	Type       string         `json:"type"`
	Source     string         `json:"source"`
	Status     string         `json:"status"`
	CheckIn    sql.NullTime   `json:"check_in"`
	CheckOut   sql.NullTime   `json:"check_out"`
	Note       sql.NullString `json:"note"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
}
