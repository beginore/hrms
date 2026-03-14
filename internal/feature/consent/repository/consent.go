package repository

import (
	"context"
	"database/sql"
	"hrms/internal/feature/consent/repository/postgres"

	"github.com/google/uuid"
)

type consentRepository struct {
	queries postgres.Querier
}

func NewRepository(conn *sql.DB) ConsentRepository {
	return &consentRepository{queries: postgres.New(conn)}
}

func (r *consentRepository) GetActiveDocuments(ctx context.Context) ([]ActiveDocument, error) {
	rows, err := r.queries.GetActiveDocuments(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]ActiveDocument, len(rows))
	for i, row := range rows {
		result[i] = ActiveDocument{
			Type:    row.Type,
			Version: row.Version,
			Url:     row.Url,
		}
	}
	return result, nil
}

func (r *consentRepository) GetConsentsByOrgID(ctx context.Context, orgID uuid.UUID) ([]OrgConsent, error) {
	rows, err := r.queries.GetConsentsByOrgID(ctx, orgID)
	if err != nil {
		return nil, err
	}
	result := make([]OrgConsent, len(rows))
	for i, row := range rows {
		result[i] = OrgConsent{
			DocumentType: row.DocumentType,
			Version:      row.Version,
		}
	}
	return result, nil
}

func (r *consentRepository) GetLatestDocumentVersions(ctx context.Context) ([]ActiveDocument, error) {
	rows, err := r.queries.GetLatestDocumentVersions(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]ActiveDocument, len(rows))
	for i, row := range rows {
		result[i] = ActiveDocument{
			Type:    row.Type,
			Version: row.Version,
		}
	}
	return result, nil
}

func (r *consentRepository) InsertConsentForOrg(ctx context.Context, arg InsertConsentForOrgParams) error {
	return r.queries.InsertConsentForOrg(ctx, postgres.InsertConsentForOrgParams{
		ID:           arg.ID,
		UserID:       arg.UserID,
		OrgID:        arg.OrgID,
		DocumentType: arg.DocumentType,
		Version:      arg.Version,
	})
}

func (r *consentRepository) GetAdminIDByOrgID(ctx context.Context, orgID uuid.UUID) (uuid.UUID, error) {
	return r.queries.GetAdminIDByOrgID(ctx, orgID)
}
