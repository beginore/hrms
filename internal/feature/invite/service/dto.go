package service

import "time"

type GenerateInviteRequest struct {
	OrganizationID string  `json:"organizationId"`
	FirstName      string  `json:"firstName"`
	LastName       string  `json:"lastName"`
	Email          string  `json:"email"`
	Role           *string `json:"role,omitempty"`
	Position       *string `json:"position,omitempty"`
}

type GenerateInviteResponse struct {
	InviteID         string    `json:"inviteId"`
	OrganizationID   string    `json:"organizationId"`
	OrganizationName string    `json:"organizationName"`
	Email            string    `json:"email"`
	Code             string    `json:"code"`
	ExpiresAt        time.Time `json:"expiresAt"`
}

type VerifyInviteRequest struct {
	Code string `json:"code"`
}

type VerifyInviteResponse struct {
	OrganizationID   string    `json:"organizationId"`
	OrganizationName string    `json:"organizationName"`
	FirstName        string    `json:"firstName"`
	LastName         string    `json:"lastName"`
	FullName         string    `json:"fullName"`
	Email            string    `json:"email"`
	Role             string    `json:"role"`
	Position         *string   `json:"position,omitempty"`
	ExpiresAt        time.Time `json:"expiresAt"`
	Message          string    `json:"message"`
}

type CompleteRegistrationRequest struct {
	Code        string `json:"code"`
	Password    string `json:"password"`
	PhoneNumber string `json:"phoneNumber"`
}

type CompleteRegistrationResponse struct {
	UserID         string `json:"userId"`
	OrganizationID string `json:"organizationId"`
	Role           string `json:"role"`
}
