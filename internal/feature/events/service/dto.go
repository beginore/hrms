package service

import "time"

type CreateEventRequest struct {
	Title        string  `json:"title"`
	Description  string  `json:"description"`
	StartsAt     string  `json:"startsAt"`
	EndsAt       string  `json:"endsAt"`
	Scope        string  `json:"scope"`
	DepartmentID *string `json:"departmentId"`
}

type UpdateEventRequest struct {
	Title        string  `json:"title"`
	Description  string  `json:"description"`
	StartsAt     string  `json:"startsAt"`
	EndsAt       string  `json:"endsAt"`
	DepartmentID *string `json:"departmentId"`
}

type EventResponse struct {
	ID             string    `json:"id"`
	Title          string    `json:"title"`
	Description    string    `json:"description"`
	StartsAt       time.Time `json:"startsAt"`
	EndsAt         time.Time `json:"endsAt"`
	Scope          string    `json:"scope"`
	DepartmentID   *string   `json:"departmentId"`
	DepartmentName *string   `json:"departmentName"`
	CreatedBy      string    `json:"createdBy"`
	CreatedByRole  string    `json:"createdByRole"`
	OrganizationID string    `json:"organizationId"`
	CreatedAt      time.Time `json:"createdAt"`
	CanEdit        bool      `json:"canEdit"`
	CanDelete      bool      `json:"canDelete"`
}
