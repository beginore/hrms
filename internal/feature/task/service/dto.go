package service

// ─── Assign Task ──────────────────────────────────────────────────────────────

type AssignTaskRequest struct {
	AssignedTo  string `json:"assigned_to"  binding:"required"`
	Title       string `json:"title"        binding:"required"`
	Description string `json:"description"`
	DueDate     string `json:"due_date"     binding:"required"` // YYYY-MM-DD
}

type AssignTaskResponse struct {
	TaskID string `json:"task_id"`
}

// ─── Submit Report ────────────────────────────────────────────────────────────

type SubmitReportRequest struct {
	Description string `json:"description" binding:"required"`
	ReportURL   string `json:"report_url"`
}

// ─── Task Response ────────────────────────────────────────────────────────────

type TaskResponse struct {
	ID                string  `json:"id"`
	CreatedBy         string  `json:"created_by"`
	AssignedTo        string  `json:"assigned_to"`
	Title             string  `json:"title"`
	Description       string  `json:"description,omitempty"`
	DueDate           string  `json:"due_date"`
	Status            string  `json:"status"`
	ReportDescription string  `json:"report_description,omitempty"`
	ReportURL         string  `json:"report_url,omitempty"`
	SubmittedAt       *string `json:"submitted_at,omitempty"`
	ReviewedBy        *string `json:"reviewed_by,omitempty"`
	ReviewedAt        *string `json:"reviewed_at,omitempty"`
	CreatedAt         string  `json:"created_at"`
}

// ─── Review Task ──────────────────────────────────────────────────────────────

type ReviewTaskRequest struct {
	Action string `json:"action" binding:"required"` // approve | reject
}
