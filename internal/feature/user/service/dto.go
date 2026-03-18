package service

import "time"

type UserProfileResponse struct {
	ID                 string    `json:"id"`
	OrganizationID     string    `json:"organizationId"`
	OrganizationName   string    `json:"organizationName"`
	Email              string    `json:"email"`
	Firstname          string    `json:"firstname"`
	Lastname           string    `json:"lastname"`
	FullName           string    `json:"fullName"`
	Role               string    `json:"role"`
	Phone              string    `json:"phone"`
	PhoneNumber        string    `json:"phoneNumber"`
	VerificationStatus string    `json:"verificationStatus"`
	JoinedDate         time.Time `json:"joinedDate"`
	Department         string    `json:"department"`
	DepartmentID       string    `json:"departmentId"`
	Position           string    `json:"position"`
	Salary             string    `json:"salary"`
	Location           string    `json:"location"`
}
