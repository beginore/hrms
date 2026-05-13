package repository

import (
	"context"
	"database/sql"
	"hrms/internal/feature/organization/repository/postgres"

	"github.com/google/uuid"
)

type organizationRepository struct {
	queries *postgres.Queries
}

func NewOrganizationRepository(conn *sql.DB) OrganizationRepository {
	return &organizationRepository{
		queries: postgres.New(conn),
	}
}

func (r *organizationRepository) CheckVATUnique(ctx context.Context, vatID string) (int64, error) {
	return r.queries.CheckVATUnique(ctx, vatID)
}

func (r *organizationRepository) CreateOrganization(ctx context.Context, params postgres.InsertOrganizationParams) error {
	return r.queries.InsertOrganization(ctx, params)
}

func (r *organizationRepository) CreateUser(ctx context.Context, params postgres.InsertUserParams) error {
	return r.queries.InsertUser(ctx, params)
}

func (r *organizationRepository) CreateConsent(ctx context.Context, params postgres.InsertConsentParams) error {
	return r.queries.InsertConsent(ctx, params)
}

func (r *organizationRepository) GetUserByEmail(ctx context.Context, email string) (postgres.GetUserByEmailRow, error) {
	return r.queries.GetUserByEmail(ctx, email)
}

func (r *organizationRepository) UpdateUserVerificationStatus(ctx context.Context, email string) error {
	return r.queries.UpdateUserVerificationStatus(ctx, email)
}

func (r *organizationRepository) CountEmployeesByDepartment(ctx context.Context, departmentID uuid.UUID) (int64, error) {
	return r.queries.CountEmployeesByDepartment(ctx, departmentID)
}

func (r *organizationRepository) CountEmployeesByPosition(ctx context.Context, positionID uuid.UUID) (int64, error) {
	return r.queries.CountEmployeesByPosition(ctx, positionID)
}

func (r *organizationRepository) GetDepartmentsByOrgID(ctx context.Context, orgID uuid.UUID) ([]postgres.Department, error) {
	return r.queries.GetDepartmentsByOrgID(ctx, orgID)
}

func (r *organizationRepository) GetPositionsByOrgID(ctx context.Context, orgID uuid.UUID) ([]postgres.Position, error) {
	return r.queries.GetPositionsByOrgID(ctx, orgID)
}

func (r *organizationRepository) DeleteDepartment(ctx context.Context, id uuid.UUID, orgID uuid.UUID) error {
	return r.queries.DeleteDepartment(ctx, postgres.DeleteDepartmentParams{ID: id, OrgID: orgID})
}

func (r *organizationRepository) DeletePosition(ctx context.Context, id uuid.UUID, orgID uuid.UUID) error {
	return r.queries.DeletePosition(ctx, postgres.DeletePositionParams{ID: id, OrgID: orgID})
}

func (r *organizationRepository) InsertPosition(ctx context.Context, arg postgres.InsertPositionParams) error {
	return r.queries.InsertPosition(ctx, arg)
}

func (r *organizationRepository) InsertDepartment(ctx context.Context, arg postgres.InsertDepartmentParams) error {
	return r.queries.InsertDepartment(ctx, arg)
}
