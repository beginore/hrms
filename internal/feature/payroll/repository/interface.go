package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type PayrollRepository interface {
	GetOrgIDByUserID(ctx context.Context, userID uuid.UUID) (uuid.UUID, error)
	GetEmployeeIDByUserID(ctx context.Context, userID uuid.UUID) (uuid.UUID, error)
	GetActorAccess(ctx context.Context, userID uuid.UUID) (*ActorAccess, error)

	CreateCycle(ctx context.Context, params CreateCycleParams) (*PayrollCycle, error)
	GetCycle(ctx context.Context, id uuid.UUID) (*PayrollCycle, error)
	ListCycles(ctx context.Context, orgID uuid.UUID) ([]PayrollCycle, error)
	UpdateCycleStatus(ctx context.Context, id uuid.UUID, status string, actorUserID uuid.UUID) error
	MarkCyclePaid(ctx context.Context, id uuid.UUID) error

	ListActiveEmployees(ctx context.Context, orgID uuid.UUID) ([]EmployeePayrollInput, error)
	GetEmployee(ctx context.Context, employeeID uuid.UUID) (*EmployeePayrollInput, error)
	ListAttendance(ctx context.Context, employeeID uuid.UUID, start, end time.Time) ([]AttendanceRecord, error)
	ListApprovedLeaves(ctx context.Context, employeeID uuid.UUID, start, end time.Time) ([]ApprovedLeave, error)
	ListSalaryHistory(ctx context.Context, employeeID uuid.UUID, start, end time.Time) ([]SalaryHistoryItem, error)
	CreateSalaryHistory(ctx context.Context, params CreateSalaryHistoryParams) (*SalaryHistoryItem, error)
	ListAttendanceOverrides(ctx context.Context, employeeID uuid.UUID, start, end time.Time) ([]AttendanceOverride, error)
	UpsertAttendanceOverride(ctx context.Context, params UpsertAttendanceOverrideParams) (*AttendanceOverride, error)
	UpdateEmploymentPeriod(ctx context.Context, params UpdateEmploymentPeriodParams) (*EmployeePayrollInput, error)
	GetWorkSchedule(ctx context.Context, employeeID uuid.UUID) (*WorkSchedule, error)
	ListCalendarDays(ctx context.Context, orgID uuid.UUID, start, end time.Time) ([]CalendarDay, error)
	GetPolicy(ctx context.Context, orgID uuid.UUID) (*PayrollPolicy, error)
	UpsertPolicy(ctx context.Context, params UpsertPolicyParams) (*PayrollPolicy, error)

	CreateAdjustment(ctx context.Context, params CreateAdjustmentParams) (*PayrollAdjustment, error)
	ListAdjustments(ctx context.Context, orgID uuid.UUID, cycleID *uuid.UUID, employeeID *uuid.UUID) ([]PayrollAdjustment, error)
	DeleteAdjustment(ctx context.Context, id uuid.UUID, orgID uuid.UUID) error

	CreateTaxRule(ctx context.Context, params CreateTaxRuleParams) (*PayrollTaxRule, error)
	ListActiveTaxRules(ctx context.Context, orgID uuid.UUID, asOf time.Time) ([]PayrollTaxRule, error)
	CreateCorrection(ctx context.Context, params CreateCorrectionParams) (*PayrollCorrection, error)

	UpsertPayrollItem(ctx context.Context, params UpsertPayrollItemParams) (*PayrollItem, error)
	ListPayrollItems(ctx context.Context, cycleID uuid.UUID) ([]PayrollItem, error)
	GetPayrollItem(ctx context.Context, id uuid.UUID) (*PayrollItem, error)
	GetPayrollItemByCycleEmployee(ctx context.Context, cycleID, employeeID uuid.UUID) (*PayrollItem, error)
	ReviewPayrollItem(ctx context.Context, id uuid.UUID, reviewedBy uuid.UUID, comment string) (*PayrollItem, error)

	InsertAuditLog(ctx context.Context, orgID uuid.UUID, cycleID *uuid.UUID, itemID *uuid.UUID, actorUserID uuid.UUID, action string, before, after []byte) error
}
