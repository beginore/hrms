package repository

import (
	"context"

	"github.com/google/uuid"
)

type EmployeeRepository interface {
	CreateEmployee(ctx context.Context, params CreateEmployeeParams) error
	GetByID(ctx context.Context, id uuid.UUID) (Employee, error)
	GetByOrgID(ctx context.Context, orgID uuid.UUID) ([]Employee, error)
	UpdateRole(ctx context.Context, id uuid.UUID, role string) error
	UpdateSalary(ctx context.Context, id uuid.UUID, salaryRate string) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
	UpdateDepartment(ctx context.Context, id uuid.UUID, departmentID uuid.UUID) error
	UpdatePosition(ctx context.Context, id uuid.UUID, positionID uuid.UUID) error
	Delete(ctx context.Context, id uuid.UUID) error
	Exists(ctx context.Context, id uuid.UUID) (bool, error)
	GetOrgIDByUserID(ctx context.Context, userID uuid.UUID) (uuid.UUID, error)
}
