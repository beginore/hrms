package repository

import (
	"time"

	"github.com/google/uuid"
)

type ActorAccess struct {
	UserID uuid.UUID
	OrgID  uuid.UUID
	Role   string
}

type DashboardMetrics struct {
	TotalEmployees       int
	ActiveEmployees      int
	PresentToday         int
	LateToday            int
	OnLeaveToday         int
	AbsentToday          int
	MonthlyPayroll       string
	PayrollCurrency      string
	EmployeeGrowth       int
	PayrollGrowthPercent string
	AttendanceRate       string
}

type PayrollSummary struct {
	PeriodStart        time.Time
	PeriodEnd          time.Time
	CyclesCount        int
	EmployeesPaid      int
	GrossSalaryTotal   string
	NetSalaryTotal     string
	BonusesTotal       string
	DeductionsTotal    string
	TaxesTotal         string
	EmployerTaxesTotal string
	TotalEmployerCost  string
	AverageNetSalary   string
	Currency           string
}

type TrendPoint struct {
	Label string
	Value float64
}

type DepartmentPayrollRow struct {
	DepartmentID       *uuid.UUID
	DepartmentName     string
	EmployeesCount     int
	AverageSalary      string
	GrossSalaryTotal   string
	NetSalaryTotal     string
	BonusesTotal       string
	DeductionsTotal    string
	TaxesTotal         string
	EmployerTaxesTotal string
	TotalEmployerCost  string
	Currency           string
}

type AttendanceBreakdown struct {
	Date           time.Time
	TotalActive    int
	Present        int
	Late           int
	OnLeave        int
	Absent         int
	Remote         int
	AttendanceRate string
}

type WeeklyAttendancePoint struct {
	Date    time.Time
	Present int
	Late    int
	OnLeave int
	Absent  int
}

type CountByLabel struct {
	Label string
	Count int
}

type EmployeeStatistics struct {
	TotalEmployees    int
	ActiveEmployees   int
	InactiveEmployees int
	AverageSalary     string
	ByDepartment      []CountByLabel
	ByRole            []CountByLabel
	ByStatus          []CountByLabel
}

type DepartmentStatisticsRow struct {
	DepartmentID      *uuid.UUID
	DepartmentName    string
	EmployeesCount    int
	AverageSalary     string
	AttendanceRate    string
	AbsentDays        int
	LateDays          int
	OvertimeMinutes   int
	NetSalaryTotal    string
	TotalEmployerCost string
	Currency          string
}

type PayrollExportRow struct {
	CycleID        uuid.UUID
	PeriodStart    time.Time
	PeriodEnd      time.Time
	EmployeeID     uuid.UUID
	EmployeeName   string
	DepartmentName string
	GrossSalary    string
	Deductions     string
	Taxes          string
	NetSalary      string
	Currency       string
	Status         string
	ReviewRequired bool
}

type AttendanceExportRow struct {
	Date           time.Time
	EmployeeID     uuid.UUID
	EmployeeName   string
	DepartmentName string
	Type           string
	Source         string
	Status         string
	CheckIn        *time.Time
	CheckOut       *time.Time
}
