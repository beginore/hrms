package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type ReportsRepository interface {
	GetActorAccess(ctx context.Context, userID uuid.UUID) (*ActorAccess, error)
	GetDashboardMetrics(ctx context.Context, orgID uuid.UUID, today, monthStart, monthEnd, previousMonthStart, previousMonthEnd time.Time) (*DashboardMetrics, error)
	GetPayrollSummary(ctx context.Context, orgID uuid.UUID, start, end time.Time) (*PayrollSummary, error)
	GetPayrollTrends(ctx context.Context, orgID uuid.UUID, start, end time.Time) ([]TrendPoint, error)
	GetDepartmentPayroll(ctx context.Context, orgID uuid.UUID, start, end time.Time) ([]DepartmentPayrollRow, error)
	GetAttendanceToday(ctx context.Context, orgID uuid.UUID, today time.Time) (*AttendanceBreakdown, error)
	GetAttendanceWeekly(ctx context.Context, orgID uuid.UUID, start, end time.Time) ([]WeeklyAttendancePoint, error)
	GetEmployeeStatistics(ctx context.Context, orgID uuid.UUID) (*EmployeeStatistics, error)
	GetDepartmentStatistics(ctx context.Context, orgID uuid.UUID, start, end time.Time) ([]DepartmentStatisticsRow, error)
	ListPayrollExportRows(ctx context.Context, orgID uuid.UUID, start, end time.Time) ([]PayrollExportRow, error)
	ListAttendanceExportRows(ctx context.Context, orgID uuid.UUID, start, end time.Time) ([]AttendanceExportRow, error)
}
