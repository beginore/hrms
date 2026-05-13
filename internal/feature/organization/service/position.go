package service

import (
	"context"
	"fmt"

	"hrms/internal/feature/organization/repository/postgres"

	"github.com/google/uuid"
)

func (s *SignUpService) CreatePosition(ctx context.Context, req CreatePositionRequest) (*CreatePositionResponse, error) {
	posID := uuid.New()
	if err := s.repo.InsertPosition(ctx, postgres.InsertPositionParams{
		ID:    posID,
		OrgID: req.OrgID,
		Name:  req.PositionName,
	}); err != nil {
		return nil, fmt.Errorf("%w: position insert: %v", ErrDatabaseInsertFailed, err)
	}
	return &CreatePositionResponse{ID: posID.String(), Name: req.PositionName}, nil
}

func (s *SignUpService) ListPositions(ctx context.Context, orgID uuid.UUID) ([]PositionItem, error) {
	rows, err := s.repo.GetPositionsByOrgID(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("%w: list positions: %v", ErrDatabaseQueryFailed, err)
	}
	items := make([]PositionItem, len(rows))
	for i, r := range rows {
		items[i] = PositionItem{ID: r.ID.String(), Name: r.Name}
	}
	return items, nil
}

func (s *SignUpService) DeletePosition(ctx context.Context, id uuid.UUID, orgID uuid.UUID) error {
	count, err := s.repo.CountEmployeesByPosition(ctx, id)
	if err != nil {
		return fmt.Errorf("%w: count employees by position: %v", ErrDatabaseQueryFailed, err)
	}
	if count > 0 {
		return ErrPositionHasEmployees
	}
	if err := s.repo.DeletePosition(ctx, id, orgID); err != nil {
		return fmt.Errorf("%w: delete position: %v", ErrDatabaseInsertFailed, err)
	}
	return nil
}
