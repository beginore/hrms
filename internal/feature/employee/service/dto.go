package service

type CreateEmployeeRequest struct {
	UserID       string `json:"userId"`
	DepartmentID string `json:"departmentId"`
	PositionID   string `json:"positionId"`
	Role         string `json:"role"`
	SalaryRate   string `json:"salaryRate"`
	Status       string `json:"status"`
}

type CreateEmployeeResponse struct {
	EmployeeID string `json:"employeeId"`
}

type UpdateEmployeeRoleRequest struct {
	Role string `json:"role"`
}

type UpdateEmployeeSalaryRequest struct {
	SalaryRate string `json:"salaryRate"`
}

type UpdateEmployeeStatusRequest struct {
	Status string `json:"status"`
}

type UpdateEmployeeDepartmentRequest struct {
	DepartmentID string `json:"departmentId"`
}

type UpdateEmployeePositionRequest struct {
	PositionID string `json:"positionId"`
}

type EmployeeResponse struct {
	ID             string `json:"id"`
	OrgID          string `json:"orgId"`
	UserID         string `json:"userId"`
	DepartmentID   string `json:"departmentId"`
	PositionID     string `json:"positionId"`
	Role           string `json:"role"`
	SalaryRate     string `json:"salaryRate"`
	Status         string `json:"status"`
	FirstName      string `json:"firstName"`
	LastName       string `json:"lastName"`
	Email          string `json:"email"`
	PhoneNumber    string `json:"phoneNumber"`
	DepartmentName string `json:"departmentName"`
	PositionName   string `json:"positionName"`
}
