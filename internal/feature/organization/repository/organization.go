package repository

import (
	"context"
	"database/sql"
	"hrms/internal/feature/organization/repository/postgres"
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
