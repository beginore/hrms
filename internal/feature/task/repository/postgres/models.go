package postgres

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type Task struct {
	ID                uuid.UUID      `json:"id"`
	OrgID             uuid.UUID      `json:"org_id"`
	CreatedBy         uuid.UUID      `json:"created_by"`
	AssignedTo        uuid.UUID      `json:"assigned_to"`
	Title             string         `json:"title"`
	Description       sql.NullString `json:"description"`
	DueDate           time.Time      `json:"due_date"`
	Status            string         `json:"status"`
	ReportDescription sql.NullString `json:"report_description"`
	ReportURL         sql.NullString `json:"report_url"`
	SubmittedAt       sql.NullTime   `json:"submitted_at"`
	ReviewedBy        uuid.NullUUID  `json:"reviewed_by"`
	ReviewedAt        sql.NullTime   `json:"reviewed_at"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
}
