package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	employeeRepository "hrms/internal/feature/employee/repository"
	"log"

	"github.com/google/uuid"
)

type EmployeeService interface {
	CreateEmployee(ctx context.Context, callerUserID uuid.UUID, req CreateEmployeeRequest) (*CreateEmployeeResponse, error)
	GetByID(ctx context.Context, id uuid.UUID) (*EmployeeResponse, error)
	GetByOrgID(ctx context.Context, callerUserID uuid.UUID) ([]EmployeeResponse, error)
	UpdateRole(ctx context.Context, id uuid.UUID, req UpdateEmployeeRoleRequest) error
	UpdateSalary(ctx context.Context, id uuid.UUID, req UpdateEmployeeSalaryRequest) error
	UpdateStatus(ctx context.Context, id uuid.UUID, req UpdateEmployeeStatusRequest) error
	UpdateDepartment(ctx context.Context, id uuid.UUID, req UpdateEmployeeDepartmentRequest) error
	UpdatePosition(ctx context.Context, id uuid.UUID, req UpdateEmployeePositionRequest) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type employeeService struct {
	repo employeeRepository.EmployeeRepository
}

func NewEmployeeService(repo employeeRepository.EmployeeRepository) EmployeeService {
	return &employeeService{repo: repo}
}

func (s *employeeService) CreateEmployee(ctx context.Context, callerUserID uuid.UUID, req CreateEmployeeRequest) (*CreateEmployeeResponse, error) {
	orgID, err := s.resolveOrgID(ctx, callerUserID)
	if err != nil {
		return nil, err
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, ErrInvalidUserID
	}

	departmentID, err := uuid.Parse(req.DepartmentID)
	if err != nil {
		return nil, ErrInvalidDepartmentID
	}

	positionID, err := uuid.Parse(req.PositionID)
	if err != nil {
		return nil, ErrInvalidPositionID
	}

	employeeID := uuid.New()

	log.Printf("[Employee] Creating employee userID=%s orgID=%s", userID, orgID)

	if err := s.repo.CreateEmployee(ctx, employeeRepository.CreateEmployeeParams{
		ID:           employeeID,
		OrgID:        orgID,
		UserID:       userID,
		DepartmentID: departmentID,
		PositionID:   positionID,
		Role:         req.Role,
		SalaryRate:   req.SalaryRate,
		Status:       req.Status,
	}); err != nil {
		log.Printf("[Employee] CreateEmployee failed: %v", err)
		return nil, fmt.Errorf("%w: %v", ErrDatabaseSaveFailed, err)
	}

	log.Printf("[Employee] Employee created: %s", employeeID)
	return &CreateEmployeeResponse{EmployeeID: employeeID.String()}, nil
}

func (s *employeeService) GetByID(ctx context.Context, id uuid.UUID) (*EmployeeResponse, error) {
	log.Printf("[Employee] GetByID: %s", id)

	emp, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrEmployeeNotFound
		}
		log.Printf("[Employee] GetByID failed: %v", err)
		return nil, fmt.Errorf("%w: %v", ErrDatabaseQueryFailed, err)
	}

	resp := mapToResponse(emp)
	return &resp, nil
}

func (s *employeeService) GetByOrgID(ctx context.Context, callerUserID uuid.UUID) ([]EmployeeResponse, error) {
	orgID, err := s.resolveOrgID(ctx, callerUserID)
	if err != nil {
		return nil, err
	}

	log.Printf("[Employee] GetByOrgID: %s", orgID)

	employees, err := s.repo.GetByOrgID(ctx, orgID)
	if err != nil {
		log.Printf("[Employee] GetByOrgID failed: %v", err)
		return nil, fmt.Errorf("%w: %v", ErrDatabaseQueryFailed, err)
	}

	result := make([]EmployeeResponse, len(employees))
	for i, emp := range employees {
		result[i] = mapToResponse(emp)
	}
	return result, nil
}

func (s *employeeService) UpdateRole(ctx context.Context, id uuid.UUID, req UpdateEmployeeRoleRequest) error {
	log.Printf("[Employee] UpdateRole: %s -> %s", id, req.Role)

	if err := s.ensureExists(ctx, id); err != nil {
		return err
	}
	if err := s.repo.UpdateRole(ctx, id, req.Role); err != nil {
		log.Printf("[Employee] UpdateRole failed: %v", err)
		return fmt.Errorf("%w: %v", ErrDatabaseSaveFailed, err)
	}
	return nil
}

func (s *employeeService) UpdateSalary(ctx context.Context, id uuid.UUID, req UpdateEmployeeSalaryRequest) error {
	log.Printf("[Employee] UpdateSalary: %s -> %s", id, req.SalaryRate)

	if err := s.ensureExists(ctx, id); err != nil {
		return err
	}
	if err := s.repo.UpdateSalary(ctx, id, req.SalaryRate); err != nil {
		log.Printf("[Employee] UpdateSalary failed: %v", err)
		return fmt.Errorf("%w: %v", ErrDatabaseSaveFailed, err)
	}
	return nil
}

func (s *employeeService) UpdateStatus(ctx context.Context, id uuid.UUID, req UpdateEmployeeStatusRequest) error {
	log.Printf("[Employee] UpdateStatus: %s -> %s", id, req.Status)

	if err := s.ensureExists(ctx, id); err != nil {
		return err
	}
	if err := s.repo.UpdateStatus(ctx, id, req.Status); err != nil {
		log.Printf("[Employee] UpdateStatus failed: %v", err)
		return fmt.Errorf("%w: %v", ErrDatabaseSaveFailed, err)
	}
	return nil
}

func (s *employeeService) UpdateDepartment(ctx context.Context, id uuid.UUID, req UpdateEmployeeDepartmentRequest) error {
	departmentID, err := uuid.Parse(req.DepartmentID)
	if err != nil {
		return ErrInvalidDepartmentID
	}

	log.Printf("[Employee] UpdateDepartment: %s -> %s", id, departmentID)

	if err := s.ensureExists(ctx, id); err != nil {
		return err
	}
	if err := s.repo.UpdateDepartment(ctx, id, departmentID); err != nil {
		log.Printf("[Employee] UpdateDepartment failed: %v", err)
		return fmt.Errorf("%w: %v", ErrDatabaseSaveFailed, err)
	}
	return nil
}

func (s *employeeService) UpdatePosition(ctx context.Context, id uuid.UUID, req UpdateEmployeePositionRequest) error {
	positionID, err := uuid.Parse(req.PositionID)
	if err != nil {
		return ErrInvalidPositionID
	}

	log.Printf("[Employee] UpdatePosition: %s -> %s", id, positionID)

	if err := s.ensureExists(ctx, id); err != nil {
		return err
	}
	if err := s.repo.UpdatePosition(ctx, id, positionID); err != nil {
		log.Printf("[Employee] UpdatePosition failed: %v", err)
		return fmt.Errorf("%w: %v", ErrDatabaseSaveFailed, err)
	}
	return nil
}

func (s *employeeService) Delete(ctx context.Context, id uuid.UUID) error {
	log.Printf("[Employee] Delete: %s", id)

	if err := s.ensureExists(ctx, id); err != nil {
		return err
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		log.Printf("[Employee] Delete failed: %v", err)
		return fmt.Errorf("%w: %v", ErrDatabaseSaveFailed, err)
	}
	return nil
}

func (s *employeeService) resolveOrgID(ctx context.Context, userID uuid.UUID) (uuid.UUID, error) {
	orgID, err := s.repo.GetOrgIDByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return uuid.Nil, ErrInvalidUserID
		}
		log.Printf("[Employee] resolveOrgID failed for userID=%s: %v", userID, err)
		return uuid.Nil, fmt.Errorf("%w: %v", ErrDatabaseQueryFailed, err)
	}
	return orgID, nil
}

func (s *employeeService) ensureExists(ctx context.Context, id uuid.UUID) error {
	exists, err := s.repo.Exists(ctx, id)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrDatabaseQueryFailed, err)
	}
	if !exists {
		return ErrEmployeeNotFound
	}
	return nil
}

func mapToResponse(emp employeeRepository.Employee) EmployeeResponse {
	return EmployeeResponse{
		ID:             emp.ID.String(),
		OrgID:          emp.OrgID.String(),
		UserID:         emp.UserID.String(),
		DepartmentID:   emp.DepartmentID.String(),
		PositionID:     emp.PositionID.String(),
		Role:           emp.Role,
		SalaryRate:     emp.SalaryRate,
		Status:         emp.Status,
		FirstName:      emp.FirstName,
		LastName:       emp.LastName,
		Email:          emp.Email,
		PhoneNumber:    emp.PhoneNumber,
		DepartmentName: emp.DepartmentName,
		PositionName:   emp.PositionName,
	}
}
