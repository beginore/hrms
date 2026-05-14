package repository

import (
	"context"

	"github.com/google/uuid"
)

type PayslipRepository interface {
	GetActorAccess(ctx context.Context, userID uuid.UUID) (*ActorAccess, error)
	GetEmployeeIDByUserID(ctx context.Context, userID uuid.UUID) (uuid.UUID, error)

	GetPayrollItem(ctx context.Context, id uuid.UUID) (*PayrollItem, error)
	ListPayrollItemsByCycle(ctx context.Context, cycleID uuid.UUID) ([]PayrollItem, error)
	GetPayrollCycle(ctx context.Context, id uuid.UUID) (*PayrollCycle, error)
	GetEmployeeProfile(ctx context.Context, employeeID uuid.UUID) (*EmployeeProfile, error)
	GetOrganizationProfile(ctx context.Context, orgID uuid.UUID) (*OrganizationProfile, error)
	ListPayrollAdjustments(ctx context.Context, cycleID, employeeID uuid.UUID) ([]PayrollAdjustment, error)

	CreatePayslip(ctx context.Context, params CreatePayslipParams) (*Payslip, error)
	GetPayslip(ctx context.Context, id uuid.UUID) (*Payslip, error)
	GetPayslipByPayrollItem(ctx context.Context, payrollItemID uuid.UUID) (*Payslip, error)
	ListPayslips(ctx context.Context, filter ListPayslipsFilter) ([]Payslip, error)
	StorePayslipPDF(ctx context.Context, id uuid.UUID, content []byte, filename, sha256 string) (*Payslip, error)
	MarkPayslipSent(ctx context.Context, id uuid.UUID, sentToEmail string) (*Payslip, error)
	VoidPayslip(ctx context.Context, id uuid.UUID, actorUserID uuid.UUID, reason string) (*Payslip, error)
}
