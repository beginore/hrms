package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"hrms/internal/infrastructure/app/anthropic"

	reportsService "hrms/internal/feature/reports/service"
	taskRepository "hrms/internal/feature/task/repository"

	"github.com/google/uuid"
)

type AIService interface {
	ReviewTaskReport(ctx context.Context, callerUserID uuid.UUID, taskID uuid.UUID) (*TaskReviewResponse, error)
	GetAnalyticsSummary(ctx context.Context, callerUserID uuid.UUID) (*AnalyticsSummaryResponse, error)
}

type aiService struct {
	claude   *anthropic.Client
	taskRepo taskRepository.TaskRepository
	reports  reportsService.ReportsService
}

func NewAIService(
	claude *anthropic.Client,
	taskRepo taskRepository.TaskRepository,
	reports reportsService.ReportsService,
) AIService {
	return &aiService{
		claude:   claude,
		taskRepo: taskRepo,
		reports:  reports,
	}
}

func (s *aiService) ReviewTaskReport(ctx context.Context, callerUserID uuid.UUID, taskID uuid.UUID) (*TaskReviewResponse, error) {
	task, err := s.taskRepo.GetTaskByID(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("get task: %w", err)
	}
	if task.Status != "SUBMITTED" {
		return nil, fmt.Errorf("task is not in SUBMITTED status")
	}

	system := `You are an HR assistant that evaluates employee task reports.
Analyze the task and the submitted report, then respond with a JSON object using this exact structure:
{
  "summary": "brief summary of what the employee did",
  "quality": "complete|incomplete|needs_revision",
  "recommendation": "approve|reject",
  "reasoning": "explanation of your decision"
}
Respond with JSON only, no markdown, no extra text.`

	userPrompt := fmt.Sprintf(`Task Title: %s
Task Description: %s
Due Date: %s

Employee Report:
%s`,
		task.Title,
		task.Description,
		task.DueDate.Format("2006-01-02"),
		task.ReportDescription,
	)

	raw, err := s.claude.Complete(ctx, system, userPrompt)
	if err != nil {
		return nil, fmt.Errorf("claude: %w", err)
	}

	var result TaskReviewResponse
	if err := json.Unmarshal([]byte(stripMarkdown(raw)), &result); err != nil {
		return nil, fmt.Errorf("parse claude response: %w", err)
	}
	return &result, nil
}

func (s *aiService) GetAnalyticsSummary(ctx context.Context, callerUserID uuid.UUID) (*AnalyticsSummaryResponse, error) {
	dashboard, err := s.reports.GetDashboard(ctx, callerUserID)
	if err != nil {
		return nil, fmt.Errorf("get dashboard: %w", err)
	}

	today := time.Now().Format("2006-01-02")
	attendance, err := s.reports.GetAttendanceToday(ctx, callerUserID, today)
	if err != nil {
		attendance = nil
	}

	monthStart := time.Now().Format("2006-01-01")
	payroll, err := s.reports.GetPayrollSummary(ctx, callerUserID, monthStart, today)
	if err != nil {
		payroll = nil
	}

	system := `You are an HR analytics assistant. Analyze the provided HR data and respond with a JSON object:
{
  "summary": "2-3 sentence executive summary of the current state of the organization",
  "highlights": ["positive finding 1", "positive finding 2"],
  "concerns": ["concern 1", "concern 2"]
}
Be concise and factual. Respond with JSON only, no markdown, no extra text.`

	dataJSON, _ := json.MarshalIndent(map[string]any{
		"dashboard":  dashboard,
		"attendance": attendance,
		"payroll":    payroll,
	}, "", "  ")

	raw, err := s.claude.Complete(ctx, system, string(dataJSON))
	if err != nil {
		return nil, fmt.Errorf("claude: %w", err)
	}

	var result AnalyticsSummaryResponse
	if err := json.Unmarshal([]byte(stripMarkdown(raw)), &result); err != nil {
		return nil, fmt.Errorf("parse claude response: %w", err)
	}
	result.GeneratedAt = time.Now().Format(time.RFC3339)
	return &result, nil
}

func stripMarkdown(s string) string {
	s = strings.TrimSpace(s)
	if idx := strings.Index(s, "{"); idx > 0 {
		s = s[idx:]
	}
	if idx := strings.LastIndex(s, "}"); idx >= 0 && idx < len(s)-1 {
		s = s[:idx+1]
	}
	return strings.TrimSpace(s)
}
