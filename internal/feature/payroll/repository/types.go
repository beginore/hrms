package repository

import (
	"time"

	"github.com/google/uuid"
)

type PayrollCycle struct {
	ID          uuid.UUID
	OrgID       uuid.UUID
	PeriodStart time.Time
	PeriodEnd   time.Time
	Currency    string
	Status      string
	CreatedBy   uuid.UUID
	ApprovedBy  *uuid.UUID
	ApprovedAt  *time.Time
	PaidAt      *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type PayrollAdjustment struct {
	ID         uuid.UUID
	OrgID      uuid.UUID
	EmployeeID uuid.UUID
	CycleID    *uuid.UUID
	Type       string
	Category   string
	Amount     string
	Currency   string
	IsTaxable  bool
	Reason     string
	CreatedBy  uuid.UUID
	CreatedAt  time.Time
}

type PayrollTaxRule struct {
	ID            uuid.UUID
	OrgID         *uuid.UUID
	Country       string
	Name          string
	Rate          string
	AppliesTo     string
	Payer         string
	ThresholdMin  string
	ThresholdMax  string
	IsActive      bool
	EffectiveFrom time.Time
	EffectiveTo   *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type PayrollItem struct {
	ID                   uuid.UUID
	CycleID              uuid.UUID
	OrgID                uuid.UUID
	EmployeeID           uuid.UUID
	BaseSalary           string
	AttendanceAdjustment string
	OvertimeAmount       string
	BonusesTotal         string
	DeductionsTotal      string
	TaxesTotal           string
	GrossSalary          string
	NetSalary            string
	Currency             string
	WorkingDays          int32
	PaidDays             string
	UnpaidDays           string
	LateDays             int32
	AbsentDays           int32
	MissingDays          int32
	OvertimeMinutes      int32
	ReviewRequired       bool
	ReviewReasons        []byte
	EmployerTaxesTotal   string
	TotalEmployerCost    string
	Status               string
	ReviewedBy           *uuid.UUID
	ReviewedAt           *time.Time
	ReviewComment        string
	CalculationSnapshot  []byte
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

type PayrollPolicy struct {
	ID                        uuid.UUID
	OrgID                     uuid.UUID
	MissingAttendancePolicy   string
	RegularOvertimeMultiplier string
	HolidayOvertimeMultiplier string
	LatePenaltyMode           string
	LatePenaltyAmount         string
	RoundingMode              string
	VacationPayRate           string
	SickLeavePayRate          string
	RemotePayRate             string
	BusinessTripPayRate       string
	UnpaidLeavePayRate        string
}

type PayrollCorrection struct {
	ID            uuid.UUID
	OrgID         uuid.UUID
	EmployeeID    uuid.UUID
	SourceItemID  *uuid.UUID
	TargetCycleID uuid.UUID
	AdjustmentID  uuid.UUID
	Amount        string
	Type          string
	Reason        string
	CreatedBy     uuid.UUID
	CreatedAt     time.Time
}

type ActorAccess struct {
	UserID     uuid.UUID
	OrgID      uuid.UUID
	EmployeeID uuid.UUID
	Role       string
}

type EmployeePayrollInput struct {
	ID              uuid.UUID
	OrgID           uuid.UUID
	UserID          uuid.UUID
	SalaryRate      string
	Status          string
	HireDate        *time.Time
	TerminationDate *time.Time
}

type AttendanceRecord struct {
	ID         uuid.UUID
	EmployeeID uuid.UUID
	Date       time.Time
	Type       string
	Source     string
	Status     string
	CheckIn    *time.Time
	CheckOut   *time.Time
	Note       string
}

type WorkSchedule struct {
	EmployeeID           uuid.UUID
	WorkStart            string
	WorkEnd              string
	LateThresholdMinutes int32
}

type CalendarDay struct {
	Date time.Time
	Type string
	Name string
}

type ApprovedLeave struct {
	ID        uuid.UUID
	Type      string
	StartDate time.Time
	EndDate   time.Time
}

type SalaryHistoryItem struct {
	ID            uuid.UUID
	EmployeeID    uuid.UUID
	SalaryRate    string
	EffectiveFrom time.Time
	EffectiveTo   *time.Time
}

type AttendanceOverride struct {
	ID               uuid.UUID
	EmployeeID       uuid.UUID
	Date             time.Time
	PaidDay          bool
	AttendanceType   string
	AttendanceStatus string
	OvertimeMinutes  int32
	Note             string
}

type CreateCycleParams struct {
	ID          uuid.UUID
	OrgID       uuid.UUID
	PeriodStart time.Time
	PeriodEnd   time.Time
	Currency    string
	CreatedBy   uuid.UUID
}

type CreateAdjustmentParams struct {
	ID         uuid.UUID
	OrgID      uuid.UUID
	EmployeeID uuid.UUID
	CycleID    *uuid.UUID
	Type       string
	Category   string
	Amount     string
	Currency   string
	IsTaxable  bool
	Reason     string
	CreatedBy  uuid.UUID
}

type CreateTaxRuleParams struct {
	ID            uuid.UUID
	OrgID         *uuid.UUID
	Country       string
	Name          string
	Rate          string
	AppliesTo     string
	Payer         string
	ThresholdMin  string
	ThresholdMax  string
	EffectiveFrom time.Time
	EffectiveTo   *time.Time
}

type UpsertPayrollItemParams struct {
	ID                   uuid.UUID
	CycleID              uuid.UUID
	OrgID                uuid.UUID
	EmployeeID           uuid.UUID
	BaseSalary           string
	AttendanceAdjustment string
	OvertimeAmount       string
	BonusesTotal         string
	DeductionsTotal      string
	TaxesTotal           string
	GrossSalary          string
	NetSalary            string
	Currency             string
	WorkingDays          int32
	PaidDays             string
	UnpaidDays           string
	LateDays             int32
	AbsentDays           int32
	MissingDays          int32
	OvertimeMinutes      int32
	ReviewRequired       bool
	ReviewReasons        []byte
	EmployerTaxesTotal   string
	TotalEmployerCost    string
	Status               string
	CalculationSnapshot  []byte
}

type CreateSalaryHistoryParams struct {
	ID            uuid.UUID
	OrgID         uuid.UUID
	EmployeeID    uuid.UUID
	SalaryRate    string
	EffectiveFrom time.Time
	EffectiveTo   *time.Time
	CreatedBy     uuid.UUID
}

type UpsertAttendanceOverrideParams struct {
	ID               uuid.UUID
	OrgID            uuid.UUID
	EmployeeID       uuid.UUID
	Date             time.Time
	PaidDay          bool
	AttendanceType   string
	AttendanceStatus string
	OvertimeMinutes  int32
	Note             string
	CreatedBy        uuid.UUID
}

type UpdateEmploymentPeriodParams struct {
	EmployeeID      uuid.UUID
	OrgID           uuid.UUID
	HireDate        *time.Time
	TerminationDate *time.Time
}

type UpsertPolicyParams struct {
	OrgID                     uuid.UUID
	MissingAttendancePolicy   string
	RegularOvertimeMultiplier string
	HolidayOvertimeMultiplier string
	LatePenaltyMode           string
	LatePenaltyAmount         string
	RoundingMode              string
	VacationPayRate           string
	SickLeavePayRate          string
	RemotePayRate             string
	BusinessTripPayRate       string
	UnpaidLeavePayRate        string
}

type CreateCorrectionParams struct {
	ID            uuid.UUID
	OrgID         uuid.UUID
	EmployeeID    uuid.UUID
	SourceItemID  *uuid.UUID
	TargetCycleID uuid.UUID
	AdjustmentID  uuid.UUID
	Amount        string
	Type          string
	Reason        string
	CreatedBy     uuid.UUID
}
