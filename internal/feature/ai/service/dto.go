package service

// ─── Task Review ──────────────────────────────────────────────────────────────

type TaskReviewResponse struct {
	Summary        string `json:"summary"`
	Quality        string `json:"quality"`         // complete | incomplete | needs_revision
	Recommendation string `json:"recommendation"`  // approve | reject
	Reasoning      string `json:"reasoning"`
}

// ─── Analytics Summary ────────────────────────────────────────────────────────

type AnalyticsSummaryResponse struct {
	Summary     string   `json:"summary"`
	Highlights  []string `json:"highlights"`
	Concerns    []string `json:"concerns"`
	GeneratedAt string   `json:"generated_at"`
}
