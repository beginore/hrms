package repository

import (
	"time"

	"github.com/google/uuid"
)

type Task struct {
	ID                uuid.UUID
	OrgID             uuid.UUID
	CreatedBy         uuid.UUID
	AssignedTo        uuid.UUID
	Title             string
	Description       string
	DueDate           time.Time
	Status            string
	ReportDescription string
	ReportURL         string
	SubmittedAt       *time.Time
	ReviewedBy        *uuid.UUID
	ReviewedAt        *time.Time
	CreatedAt         time.Time
}

type CreateTaskParams struct {
	AssignedTo  uuid.UUID
	Title       string
	Description string
	DueDate     time.Time
}

type SubmitReportParams struct {
	ID                uuid.UUID
	ReportDescription string
	ReportURL         string
}

type ReviewTaskParams struct {
	ID         uuid.UUID
	Status     string
	ReviewedBy uuid.UUID
}
