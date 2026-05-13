package repository

import (
	"context"
	"hrms/internal/feature/organization/repository/postgres"

	"github.com/google/uuid"
)

type OrganizationRepository interface {
	CheckVATUnique(ctx context.Context, vatID string) (int64, error)
	GetUserByEmail(ctx context.Context, email string) (postgres.GetUserByEmailRow, error)
	CreateConsent(ctx context.Context, arg postgres.InsertConsentParams) error
	CreateOrganization(ctx context.Context, arg postgres.InsertOrganizationParams) error
	CreateUser(ctx context.Context, arg postgres.InsertUserParams) error
	UpdateUserVerificationStatus(ctx context.Context, email string) error
	InsertPosition(ctx context.Context, arg postgres.InsertPositionParams) error
	InsertDepartment(ctx context.Context, arg postgres.InsertDepartmentParams) error
	GetDepartmentsByOrgID(ctx context.Context, orgID uuid.UUID) ([]postgres.Department, error)
	GetPositionsByOrgID(ctx context.Context, orgID uuid.UUID) ([]postgres.Position, error)
	CountEmployeesByDepartment(ctx context.Context, departmentID uuid.UUID) (int64, error)
	CountEmployeesByPosition(ctx context.Context, positionID uuid.UUID) (int64, error)
	DeleteDepartment(ctx context.Context, id uuid.UUID, orgID uuid.UUID) error
	DeletePosition(ctx context.Context, id uuid.UUID, orgID uuid.UUID) error
}
