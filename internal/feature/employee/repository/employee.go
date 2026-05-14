package repository

import (
	"context"
	"database/sql"
	"hrms/internal/feature/employee/repository/postgres"

	"github.com/google/uuid"
)

type employeeRepository struct {
	queries postgres.Querier
	conn    *sql.DB
}

func NewRepository(conn *sql.DB) EmployeeRepository {
	return &employeeRepository{
		queries: postgres.New(conn),
		conn:    conn,
	}
}

func (r *employeeRepository) CreateEmployee(ctx context.Context, params CreateEmployeeParams) error {
	return r.queries.InsertEmployee(ctx, postgres.InsertEmployeeParams{
		ID:           params.ID,
		OrgID:        params.OrgID,
		UserID:       params.UserID,
		DepartmentID: params.DepartmentID,
		PositionID:   params.PositionID,
		Role:         params.Role,
		SalaryRate:   params.SalaryRate,
		Status:       params.Status,
	})
}

func (r *employeeRepository) GetByID(ctx context.Context, id uuid.UUID) (Employee, error) {
	row, err := r.queries.GetEmployeeByID(ctx, id)
	if err != nil {
		return Employee{}, err
	}
	return mapGetByIDRow(row), nil
}

func (r *employeeRepository) GetByOrgID(ctx context.Context, orgID uuid.UUID) ([]Employee, error) {
	rows, err := r.queries.GetEmployeesByOrgID(ctx, orgID)
	if err != nil {
		return nil, err
	}
	result := make([]Employee, len(rows))
	for i, row := range rows {
		result[i] = mapGetByOrgIDRow(row)
	}
	return result, nil
}

func (r *employeeRepository) ListDepartmentsByOrgID(ctx context.Context, orgID uuid.UUID) ([]Department, error) {
	rows, err := r.conn.QueryContext(ctx, `
		SELECT id, org_id, name
		FROM departments
		WHERE org_id = $1
		ORDER BY name
	`, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var departments []Department
	for rows.Next() {
		var department Department
		if err := rows.Scan(&department.ID, &department.OrgID, &department.Name); err != nil {
			return nil, err
		}
		departments = append(departments, department)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return departments, nil
}

func (r *employeeRepository) ListPositionsByOrgID(ctx context.Context, orgID uuid.UUID) ([]Position, error) {
	rows, err := r.conn.QueryContext(ctx, `
		SELECT id, org_id, name
		FROM positions
		WHERE org_id = $1
		ORDER BY name
	`, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var positions []Position
	for rows.Next() {
		var position Position
		if err := rows.Scan(&position.ID, &position.OrgID, &position.Name); err != nil {
			return nil, err
		}
		positions = append(positions, position)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return positions, nil
}

func (r *employeeRepository) UpdateRole(ctx context.Context, id uuid.UUID, role string) error {
	return r.queries.UpdateEmployeeRole(ctx, postgres.UpdateEmployeeRoleParams{
		ID:   id,
		Role: role,
	})
}

func (r *employeeRepository) UpdateSalary(ctx context.Context, id uuid.UUID, salaryRate string) error {
	return r.queries.UpdateEmployeeSalary(ctx, postgres.UpdateEmployeeSalaryParams{
		ID:         id,
		SalaryRate: salaryRate,
	})
}

func (r *employeeRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	return r.queries.UpdateEmployeeStatus(ctx, postgres.UpdateEmployeeStatusParams{
		ID:     id,
		Status: status,
	})
}

func (r *employeeRepository) UpdateDepartment(ctx context.Context, id uuid.UUID, departmentID uuid.UUID) error {
	return r.queries.UpdateEmployeeDepartment(ctx, postgres.UpdateEmployeeDepartmentParams{
		ID:           id,
		DepartmentID: departmentID,
	})
}

func (r *employeeRepository) UpdatePosition(ctx context.Context, id uuid.UUID, positionID uuid.UUID) error {
	return r.queries.UpdateEmployeePosition(ctx, postgres.UpdateEmployeePositionParams{
		ID:         id,
		PositionID: positionID,
	})
}

func (r *employeeRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.queries.DeleteEmployee(ctx, id)
}

func (r *employeeRepository) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	count, err := r.queries.CheckEmployeeExists(ctx, id)
	return count > 0, err
}

func (r *employeeRepository) GetOrgIDByUserID(ctx context.Context, userID uuid.UUID) (uuid.UUID, error) {
	return r.queries.GetOrgIDByUserID(ctx, userID)
}

func (r *employeeRepository) GetRoleAndOrgIDByUserID(ctx context.Context, userID uuid.UUID) (RoleAndOrgID, error) {
	row, err := r.queries.GetRoleAndOrgIDByUserID(ctx, userID)
	if err != nil {
		return RoleAndOrgID{}, err
	}
	return RoleAndOrgID{OrgID: row.OrgID, Role: row.Role}, nil
}

func (r *employeeRepository) GetUserRoleByUserID(ctx context.Context, userID uuid.UUID) (string, error) {
	return r.queries.GetUserRoleByUserID(ctx, userID)
}

func mapGetByIDRow(row postgres.GetEmployeeByIDRow) Employee {
	return Employee{
		ID:             row.ID,
		OrgID:          row.OrgID,
		UserID:         row.UserID,
		DepartmentID:   row.DepartmentID,
		PositionID:     row.PositionID,
		Role:           row.Role,
		SalaryRate:     row.SalaryRate,
		Status:         row.Status,
		FirstName:      row.FirstName,
		LastName:       row.LastName,
		Email:          row.Email,
		PhoneNumber:    row.PhoneNumber,
		DepartmentName: row.DepartmentName,
		PositionName:   row.PositionName,
	}
}

func mapGetByOrgIDRow(row postgres.GetEmployeesByOrgIDRow) Employee {
	return Employee{
		ID:             row.ID,
		OrgID:          row.OrgID,
		UserID:         row.UserID,
		DepartmentID:   row.DepartmentID,
		PositionID:     row.PositionID,
		Role:           row.Role,
		SalaryRate:     row.SalaryRate,
		Status:         row.Status,
		FirstName:      row.FirstName,
		LastName:       row.LastName,
		Email:          row.Email,
		PhoneNumber:    row.PhoneNumber,
		DepartmentName: row.DepartmentName,
		PositionName:   row.PositionName,
	}
}
