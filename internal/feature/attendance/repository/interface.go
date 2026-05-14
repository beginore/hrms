package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type AttendanceRepository interface {
	// work schedules
	SetWorkSchedule(ctx context.Context, callerUserID uuid.UUID, params SetWorkScheduleParams) error
	GetWorkSchedule(ctx context.Context, employeeID uuid.UUID) (WorkSchedule, error)

	// skud events
	CreateSkudEvent(ctx context.Context, callerUserID uuid.UUID, params CreateSkudEventParams) error
	ProcessSkudEvents(ctx context.Context, orgID uuid.UUID) error

	// leave requests
	CreateLeaveRequest(ctx context.Context, callerUserID uuid.UUID, params CreateLeaveRequestParams) (uuid.UUID, error)
	GetLeaveRequestByID(ctx context.Context, id uuid.UUID) (LeaveRequest, error)
	GetLeaveRequestsByOrg(ctx context.Context, callerUserID uuid.UUID) ([]LeaveRequest, error)
	GetLeaveRequestsByEmployee(ctx context.Context, employeeID uuid.UUID) ([]LeaveRequest, error)
	ReviewLeaveRequest(ctx context.Context, callerUserID uuid.UUID, params ReviewLeaveRequestParams) error

	// attendance records
	UpsertAttendance(ctx context.Context, callerUserID uuid.UUID, params UpsertAttendanceParams) error
	GetAttendanceByOrg(ctx context.Context, callerUserID uuid.UUID, period GetByPeriodParams) ([]AttendanceRecord, error)
	GetAttendanceByEmployee(ctx context.Context, employeeID uuid.UUID, period GetByPeriodParams) ([]AttendanceRecord, error)

	// helpers
	GetOrgIDByUserID(ctx context.Context, userID uuid.UUID) (uuid.UUID, error)
	GetEmployeeIDByUserID(ctx context.Context, userID uuid.UUID) (uuid.UUID, error)
	GetOrgIDByEmployeeID(ctx context.Context, employeeID uuid.UUID) (uuid.UUID, error)
	GetEmployeeNotificationInfo(ctx context.Context, employeeID uuid.UUID) (EmployeeNotificationInfo, error)
	EmployeeExists(ctx context.Context, employeeID uuid.UUID) (bool, error)
	LeaveRequestExists(ctx context.Context, id uuid.UUID) (bool, error)
	GetApprovedLeaveForDate(ctx context.Context, employeeID uuid.UUID, date time.Time) (LeaveRequest, bool, error)
	GetWorkScheduleOrDefault(ctx context.Context, employeeID uuid.UUID) WorkSchedule
}
