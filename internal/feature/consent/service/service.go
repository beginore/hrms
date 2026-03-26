package service

import (
	"context"
	"fmt"
	consentRepository "hrms/internal/feature/consent/repository"
	"log"

	"github.com/google/uuid"
)

type ConsentService interface {
	GetActiveDocuments(ctx context.Context) ([]ConsentItemResponse, error)
	SubmitConsents(ctx context.Context, req RenewConsentRequest) error
	ValidateConsents(ctx context.Context) (*ConsentValidationResponse, error)
}

type consentService struct {
	repo consentRepository.ConsentRepository
}

func NewConsentService(repo consentRepository.ConsentRepository) ConsentService {
	return &consentService{repo: repo}
}

func (s *consentService) GetActiveDocuments(ctx context.Context) ([]ConsentItemResponse, error) {
	docs, err := s.repo.GetActiveDocuments(ctx)
	if err != nil {
		return nil, err
	}
	if len(docs) == 0 {
		return nil, ErrNoActiveDocuments
	}

	result := make([]ConsentItemResponse, 0, len(docs))
	for _, d := range docs {
		result = append(result, ConsentItemResponse{
			DocumentType:  d.Type,
			LatestVersion: d.Version,
			URL:           d.Url,
		})
	}
	return result, nil
}

func (s *consentService) SubmitConsents(ctx context.Context, req RenewConsentRequest) error {
	// TODO: Parse orgID from JWT
	orgID := uuid.MustParse("00000000-0000-0000-0000-000000000000")

	adminID, err := s.repo.GetAdminIDByOrgID(ctx, orgID)
	if err != nil {
		log.Printf("[Consent] GetAdminIDByOrgID failed for org %s: %v", orgID, err)
		return fmt.Errorf("organization not found: %w", err)
	}

	activeDocs, err := s.repo.GetActiveDocuments(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch active documents: %w", err)
	}

	for _, doc := range activeDocs {
		params := consentRepository.InsertConsentForOrgParams{
			ID:           uuid.New(),
			UserID:       adminID,
			OrgID:        orgID,
			DocumentType: doc.Type,
			Version:      doc.Version,
		}
		if err := s.repo.InsertConsentForOrg(ctx, params); err != nil {
			log.Printf("[Consent] Insert failed: %v", err)
			return fmt.Errorf("failed to insert consent: %w", err)
		}
	}
	return nil
}

func (s *consentService) ValidateConsents(ctx context.Context) (*ConsentValidationResponse, error) {
	// TODO: Parse orgID from JWT
	parsedOrgID := uuid.MustParse("00000000-0000-0000-0000-000000000000")

	activeDocs, err := s.repo.GetActiveDocuments(ctx)
	if err != nil {
		return nil, err
	}

	orgConsents, err := s.repo.GetConsentsByOrgID(ctx, parsedOrgID)
	if err != nil {
		return nil, err
	}

	accepted := make(map[string]string)
	for _, c := range orgConsents {
		accepted[c.DocumentType] = c.Version
	}

	var errs []ConsentValidationError
	for _, doc := range activeDocs {
		current, ok := accepted[doc.Type]
		if !ok || current != doc.Version {
			var cur *string
			if ok {
				cur = &current
			}
			errs = append(errs, ConsentValidationError{
				DocumentType:   doc.Type,
				CurrentVersion: cur,
				LatestVersion:  doc.Version,
			})
		}
	}

	return &ConsentValidationResponse{
		Valid:  len(errs) == 0,
		Errors: errs,
	}, nil
}
