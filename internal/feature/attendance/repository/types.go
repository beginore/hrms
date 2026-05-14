package repository

import (
	"time"

	"github.com/google/uuid"
)

// ─── Work Schedule ────────────────────────────────────────────────────────────

type WorkSchedule struct {
	ID                   uuid.UUID
	OrgID                uuid.UUID
	EmployeeID           uuid.UUID
	WorkStart            string
	WorkEnd              string
	LateThresholdMinutes int32
}

type SetWorkScheduleParams struct {
	EmployeeID           uuid.UUID
	WorkStart            string
	WorkEnd              string
	LateThresholdMinutes int32
}

// ─── SKUD Events ──────────────────────────────────────────────────────────────

type SkudEvent struct {
	ID         uuid.UUID
	OrgID      uuid.UUID
	EmployeeID uuid.UUID
	EventType  string
	DeviceID   string
	OccurredAt time.Time
	Processed  bool
}

type CreateSkudEventParams struct {
	EmployeeID uuid.UUID
	EventType  string
	DeviceID   string
	OccurredAt time.Time
}

// ─── Leave Requests ───────────────────────────────────────────────────────────

type LeaveRequest struct {
	ID          uuid.UUID
	OrgID       uuid.UUID
	EmployeeID  uuid.UUID
	Type        string
	StartDate   time.Time
	EndDate     time.Time
	Reason      string
	DocumentURL string
	Status      string
	ReviewedBy  *uuid.UUID
	ReviewedAt  *time.Time
	CreatedAt   time.Time
}

type CreateLeaveRequestParams struct {
	EmployeeID  uuid.UUID
	Type        string
	StartDate   time.Time
	EndDate     time.Time
	Reason      string
	DocumentURL string
}

type ReviewLeaveRequestParams struct {
	ID         uuid.UUID
	Status     string
	ReviewedBy uuid.UUID
}

// ─── Attendance Records ───────────────────────────────────────────────────────

type AttendanceRecord struct {
	ID         uuid.UUID
	OrgID      uuid.UUID
	EmployeeID uuid.UUID
	Date       time.Time
	Type       string
	Source     string
	Status     string
	CheckIn    *time.Time
	CheckOut   *time.Time
	Note       string
}

type ManualCheckInParams struct {
	EmployeeID uuid.UUID
	OrgID      uuid.UUID
	WorkType   string
	Status     string
	CheckIn    time.Time
	Note       string
}

type UpsertAttendanceParams struct {
	EmployeeID uuid.UUID
	Date       time.Time
	Type       string
	Source     string
	Status     string
	CheckIn    *time.Time
	CheckOut   *time.Time
	Note       string
}

type GetByPeriodParams struct {
	StartDate time.Time
	EndDate   time.Time
}
