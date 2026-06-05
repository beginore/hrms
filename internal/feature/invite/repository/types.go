package repository

import "time"

type Invite struct {
	ID               string
	OrgID            string
	OrganizationName string
	FirstName        string
	LastName         string
	Email            string
	Code             string
	Role             string
	Position         *string
	DepartmentID     string
	PositionID       string
	SalaryRate       string
	EmployeeStatus   string
	ExpiresAt        time.Time
	IsUsed           bool
	UsedAt           *time.Time
	CreatedAt        time.Time
}
