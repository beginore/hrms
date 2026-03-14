package repository

import (
	"context"
	"hrms/internal/feature/organization/repository/postgres"
)

type OrganizationRepository interface {
	CheckVATUnique(ctx context.Context, vatID string) (int64, error)
	GetUserByEmail(ctx context.Context, email string) (postgres.GetUserByEmailRow, error)
	CreateConsent(ctx context.Context, arg postgres.InsertConsentParams) error
	CreateOrganization(ctx context.Context, arg postgres.InsertOrganizationParams) error
	CreateUser(ctx context.Context, arg postgres.InsertUserParams) error
	UpdateUserVerificationStatus(ctx context.Context, email string) error
}
