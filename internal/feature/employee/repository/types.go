package repository

import "github.com/google/uuid"

type Employee struct {
	ID             uuid.UUID
	OrgID          uuid.UUID
	UserID         uuid.UUID
	DepartmentID   uuid.UUID
	PositionID     uuid.UUID
	Role           string
	SalaryRate     string
	Status         string
	FirstName      string
	LastName       string
	Email          string
	PhoneNumber    string
	DepartmentName string
	PositionName   string
}

type CreateEmployeeParams struct {
	ID           uuid.UUID
	OrgID        uuid.UUID
	UserID       uuid.UUID
	DepartmentID uuid.UUID
	PositionID   uuid.UUID
	Role         string
	SalaryRate   string
	Status       string
}
