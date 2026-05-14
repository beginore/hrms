package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Querier interface {
	// work schedules
	UpsertWorkSchedule(ctx context.Context, arg UpsertWorkScheduleParams) error
	GetWorkScheduleByEmployeeID(ctx context.Context, employeeID uuid.UUID) (WorkSchedule, error)

	// skud events
	InsertSkudEvent(ctx context.Context, arg InsertSkudEventParams) error
	GetUnprocessedSkudEvents(ctx context.Context, orgID uuid.UUID) ([]SkudEvent, error)
	MarkSkudEventProcessed(ctx context.Context, id uuid.UUID) error

	// leave requests
	InsertLeaveRequest(ctx context.Context, arg InsertLeaveRequestParams) error
	GetLeaveRequestByID(ctx context.Context, id uuid.UUID) (LeaveRequest, error)
	GetLeaveRequestsByOrgID(ctx context.Context, orgID uuid.UUID) ([]LeaveRequest, error)
	GetLeaveRequestsByEmployeeID(ctx context.Context, employeeID uuid.UUID) ([]LeaveRequest, error)
	GetApprovedLeaveForDate(ctx context.Context, arg GetApprovedLeaveForDateParams) (LeaveRequest, error)
	UpdateLeaveRequestStatus(ctx context.Context, arg UpdateLeaveRequestStatusParams) error

	// attendance records
	UpsertAttendanceRecord(ctx context.Context, arg UpsertAttendanceRecordParams) error
	GetAttendanceByOrgAndPeriod(ctx context.Context, arg GetAttendanceByOrgAndPeriodParams) ([]AttendanceRecord, error)
	GetAttendanceByEmployeeAndPeriod(ctx context.Context, arg GetAttendanceByEmployeeAndPeriodParams) ([]AttendanceRecord, error)
	GetAttendanceByEmployeeAndDate(ctx context.Context, arg GetAttendanceByEmployeeAndDateParams) (AttendanceRecord, error)
	UpdateCheckOut(ctx context.Context, employeeID uuid.UUID, date time.Time, checkOut time.Time) error
	GetOrgIDByEmployeeID(ctx context.Context, employeeID uuid.UUID) (uuid.UUID, error)
	GetOrgIDByUserID(ctx context.Context, userID uuid.UUID) (uuid.UUID, error)
	GetEmployeeIDByUserID(ctx context.Context, userID uuid.UUID) (uuid.UUID, error)
	CheckEmployeeExists(ctx context.Context, employeeID uuid.UUID) (int64, error)
	CheckLeaveRequestExists(ctx context.Context, id uuid.UUID) (int64, error)
}

type UpsertWorkScheduleParams struct {
	ID                   uuid.UUID
	OrgID                uuid.UUID
	EmployeeID           uuid.UUID
	WorkStart            string
	WorkEnd              string
	LateThresholdMinutes int32
}

type InsertSkudEventParams struct {
	ID         uuid.UUID
	OrgID      uuid.UUID
	EmployeeID uuid.UUID
	EventType  string
	DeviceID   string
	OccurredAt time.Time
}

type InsertLeaveRequestParams struct {
	ID          uuid.UUID
	OrgID       uuid.UUID
	EmployeeID  uuid.UUID
	Type        string
	StartDate   time.Time
	EndDate     time.Time
	Reason      string
	DocumentURL string
}

type GetApprovedLeaveForDateParams struct {
	EmployeeID uuid.UUID
	Date       time.Time
}

type UpdateLeaveRequestStatusParams struct {
	ID         uuid.UUID
	Status     string
	ReviewedBy uuid.UUID
}

type UpsertAttendanceRecordParams struct {
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

type GetAttendanceByOrgAndPeriodParams struct {
	OrgID     uuid.UUID
	StartDate time.Time
	EndDate   time.Time
}

type GetAttendanceByEmployeeAndPeriodParams struct {
	EmployeeID uuid.UUID
	StartDate  time.Time
	EndDate    time.Time
}

type GetAttendanceByEmployeeAndDateParams struct {
	EmployeeID uuid.UUID
	Date       time.Time
}
