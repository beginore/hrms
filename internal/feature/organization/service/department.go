package service

import (
	"context"
	"fmt"

	"hrms/internal/feature/organization/repository/postgres"

	"github.com/google/uuid"
)

func (s *SignUpService) CreateDepartment(ctx context.Context, req CreateDepartmentRequest) (*CreateDepartmentResponse, error) {
	depID := uuid.New()
	if err := s.repo.InsertDepartment(ctx, postgres.InsertDepartmentParams{
		ID:    depID,
		OrgID: req.OrgID,
		Name:  req.DepartmentName,
	}); err != nil {
		return nil, fmt.Errorf("%w: department insert: %v", ErrDatabaseInsertFailed, err)
	}
	return &CreateDepartmentResponse{ID: depID.String(), Name: req.DepartmentName}, nil
}

func (s *SignUpService) ListDepartments(ctx context.Context, orgID uuid.UUID) ([]DepartmentItem, error) {
	rows, err := s.repo.GetDepartmentsByOrgID(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("%w: list departments: %v", ErrDatabaseQueryFailed, err)
	}
	items := make([]DepartmentItem, len(rows))
	for i, r := range rows {
		items[i] = DepartmentItem{ID: r.ID.String(), Name: r.Name}
	}
	return items, nil
}

func (s *SignUpService) DeleteDepartment(ctx context.Context, id uuid.UUID, orgID uuid.UUID) error {
	count, err := s.repo.CountEmployeesByDepartment(ctx, id)
	if err != nil {
		return fmt.Errorf("%w: count employees by department: %v", ErrDatabaseQueryFailed, err)
	}
	if count > 0 {
		return ErrDepartmentHasEmployees
	}
	if err := s.repo.DeleteDepartment(ctx, id, orgID); err != nil {
		return fmt.Errorf("%w: delete department: %v", ErrDatabaseInsertFailed, err)
	}
	return nil
}
