package repository

import (
	"context"

	"github.com/google/uuid"
)

type ConsentRepository interface {
	GetActiveDocuments(ctx context.Context) ([]ActiveDocument, error)
	GetConsentsByOrgID(ctx context.Context, orgID uuid.UUID) ([]OrgConsent, error)
	GetLatestDocumentVersions(ctx context.Context) ([]ActiveDocument, error)
	InsertConsentForOrg(ctx context.Context, arg InsertConsentForOrgParams) error
	GetAdminIDByOrgID(ctx context.Context, orgID uuid.UUID) (uuid.UUID, error)
}
