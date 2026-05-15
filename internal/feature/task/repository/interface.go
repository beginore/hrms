package repository

import (
	"context"

	"github.com/google/uuid"
)

type TaskRepository interface {
	CreateTask(ctx context.Context, callerUserID uuid.UUID, params CreateTaskParams) (uuid.UUID, error)
	GetTaskByID(ctx context.Context, id uuid.UUID) (Task, error)
	ListTasksByOrg(ctx context.Context, callerUserID uuid.UUID) ([]Task, error)
	ListTasksByEmployee(ctx context.Context, employeeID uuid.UUID) ([]Task, error)
	SubmitReport(ctx context.Context, params SubmitReportParams) error
	ReviewTask(ctx context.Context, callerUserID uuid.UUID, params ReviewTaskParams) error

	GetOrgIDByUserID(ctx context.Context, userID uuid.UUID) (uuid.UUID, error)
	GetEmployeeIDByUserID(ctx context.Context, userID uuid.UUID) (uuid.UUID, error)
	EmployeeExistsInOrg(ctx context.Context, employeeID, orgID uuid.UUID) (bool, error)
}
