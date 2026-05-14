package repository

import (
	"time"

	"github.com/google/uuid"
)

type ActorAccess struct {
	UserID     uuid.UUID
	OrgID      uuid.UUID
	EmployeeID uuid.UUID
	Role       string
}

type PayrollCycle struct {
	ID          uuid.UUID
	OrgID       uuid.UUID
	PeriodStart time.Time
	PeriodEnd   time.Time
	Status      string
	Currency    string
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
	CalculationSnapshot  []byte
}

type EmployeeProfile struct {
	ID             uuid.UUID
	OrgID          uuid.UUID
	UserID         uuid.UUID
	FirstName      string
	LastName       string
	Email          string
	Role           string
	DepartmentName string
	PositionName   string
}

type OrganizationProfile struct {
	ID      uuid.UUID
	Name    string
	VATID   string
	Address string
}

type PayrollAdjustment struct {
	ID        uuid.UUID
	Type      string
	Category  string
	Amount    string
	Currency  string
	IsTaxable bool
	Reason    string
}

type Payslip struct {
	ID                 uuid.UUID
	OrgID              uuid.UUID
	EmployeeID         uuid.UUID
	PayrollCycleID     uuid.UUID
	PayrollItemID      uuid.UUID
	PeriodStart        time.Time
	PeriodEnd          time.Time
	Status             string
	Currency           string
	BaseSalary         string
	OvertimeAmount     string
	BonusesTotal       string
	DeductionsTotal    string
	TaxesTotal         string
	GrossSalary        string
	NetSalary          string
	EmployerTaxesTotal string
	TotalEmployerCost  string
	PayloadSnapshot    []byte
	PDFContent         []byte
	PDFFilename        string
	PDFGeneratedAt     *time.Time
	PDFSHA256          string
	SentToEmail        string
	SentAt             *time.Time
	GeneratedBy        uuid.UUID
	GeneratedAt        time.Time
	VoidedBy           *uuid.UUID
	VoidedAt           *time.Time
	VoidReason         string
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type CreatePayslipParams struct {
	ID                 uuid.UUID
	OrgID              uuid.UUID
	EmployeeID         uuid.UUID
	PayrollCycleID     uuid.UUID
	PayrollItemID      uuid.UUID
	PeriodStart        time.Time
	PeriodEnd          time.Time
	Status             string
	Currency           string
	BaseSalary         string
	OvertimeAmount     string
	BonusesTotal       string
	DeductionsTotal    string
	TaxesTotal         string
	GrossSalary        string
	NetSalary          string
	EmployerTaxesTotal string
	TotalEmployerCost  string
	PayloadSnapshot    []byte
	PDFContent         []byte
	PDFFilename        string
	PDFGeneratedAt     *time.Time
	PDFSHA256          string
	GeneratedBy        uuid.UUID
}

type ListPayslipsFilter struct {
	OrgID       uuid.UUID
	EmployeeID  *uuid.UUID
	CycleID     *uuid.UUID
	Status      string
	PeriodStart *time.Time
	PeriodEnd   *time.Time
}
