package repository

import "github.com/google/uuid"

type ActiveDocument struct {
	Type    string
	Version string
	Url     string
}

type OrgConsent struct {
	DocumentType string
	Version      string
}

type InsertConsentForOrgParams struct {
	ID           uuid.UUID
	UserID       uuid.UUID
	OrgID        uuid.UUID
	DocumentType string
	Version      string
}
