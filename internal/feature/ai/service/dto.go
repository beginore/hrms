package service

// ─── Task Review ──────────────────────────────────────────────────────────────

type TaskReviewResponse struct {
	Summary        string `json:"summary"`
	Quality        string `json:"quality"`        // complete | incomplete | needs_revision
	Recommendation string `json:"recommendation"` // approve | reject
	Reasoning      string `json:"reasoning"`
}

// ─── Analytics Summary ────────────────────────────────────────────────────────

type AnalyticsSummaryResponse struct {
	Summary     string                  `json:"summary"`
	Highlights  []string                `json:"highlights"`
	Concerns    []string                `json:"concerns"`
	GeneratedAt string                  `json:"generated_at"`
	Period      *AnalyticsSummaryPeriod `json:"period,omitempty"`
}

type AnalyticsSummaryRequest struct {
	Date      string
	StartDate string
	EndDate   string
}

type AnalyticsSummaryPeriod struct {
	AttendanceDate string `json:"attendance_date"`
	WeekStart      string `json:"week_start"`
	StartDate      string `json:"start_date"`
	EndDate        string `json:"end_date"`
	Source         string `json:"source"`
}
