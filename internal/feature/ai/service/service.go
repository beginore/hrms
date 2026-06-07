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
	GetAnalyticsSummary(ctx context.Context, callerUserID uuid.UUID, req AnalyticsSummaryRequest) (*AnalyticsSummaryResponse, error)
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

func (s *aiService) GetAnalyticsSummary(ctx context.Context, callerUserID uuid.UUID, req AnalyticsSummaryRequest) (*AnalyticsSummaryResponse, error) {
	period := resolveAnalyticsPeriod(req, time.Now())

	dashboard, err := s.reports.GetDashboard(ctx, callerUserID)
	if err != nil {
		return nil, fmt.Errorf("get dashboard: %w", err)
	}

	attendance, err := s.reports.GetAttendanceToday(ctx, callerUserID, period.AttendanceDate)
	if err != nil {
		attendance = nil
	}

	payroll, err := s.reports.GetPayrollSummary(ctx, callerUserID, period.StartDate, period.EndDate)
	if err != nil {
		payroll = nil
	}

	if payroll != nil && payroll.EmployeesPaid == 0 && !hasExplicitAnalyticsPeriod(req) {
		period = previousCompletedMonthPeriod(time.Now())
		attendance, _ = s.reports.GetAttendanceToday(ctx, callerUserID, period.AttendanceDate)
		payroll, _ = s.reports.GetPayrollSummary(ctx, callerUserID, period.StartDate, period.EndDate)
	}

	attendanceWeekly, _ := s.reports.GetAttendanceWeekly(ctx, callerUserID, period.WeekStart)
	employeeStats, _ := s.reports.GetEmployeeStatistics(ctx, callerUserID)
	departmentStats, _ := s.reports.GetDepartmentStatistics(ctx, callerUserID, period.StartDate, period.EndDate)
	departmentPayroll, _ := s.reports.GetDepartmentPayroll(ctx, callerUserID, period.StartDate, period.EndDate)
	payrollTrends, _ := s.reports.GetPayrollTrends(ctx, callerUserID, payrollTrendStart(period.StartDate), period.EndDate)

	system := `You are an HR analytics assistant. Analyze the provided HR data and respond with a JSON object:
{
  "summary": "2-3 sentence executive summary of the current state of the organization",
  "highlights": ["positive finding 1", "positive finding 2"],
  "concerns": ["concern 1", "concern 2"]
}
Be concise and factual. Use the provided analytics_period as the reporting context.
Do not infer a company-wide absence or payroll failure from the live dashboard alone when the reporting period contains attendance/payroll data.
If the selected date is outside the seeded/reporting period or has no attendance records, say the date has no records instead of calling it a workforce disappearance.
Respond with JSON only, no markdown, no extra text.`

	dataJSON, _ := json.MarshalIndent(map[string]any{
		"analytics_period":      period,
		"dashboard_live":        dashboard,
		"attendance_day":        attendance,
		"attendance_week":       attendanceWeekly,
		"employee_statistics":   employeeStats,
		"department_stats":      departmentStats,
		"payroll_summary":       payroll,
		"payroll_by_department": departmentPayroll,
		"payroll_trends":        payrollTrends,
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
	result.Period = &period
	return &result, nil
}

func resolveAnalyticsPeriod(req AnalyticsSummaryRequest, now time.Time) AnalyticsSummaryPeriod {
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	reportDate := truncateDate(now)
	start := monthStart
	end := reportDate
	source := "current_month_to_date"

	if parsed, ok := parseAnalyticsDate(req.StartDate); ok {
		start = parsed
		source = "explicit"
	}
	if parsed, ok := parseAnalyticsDate(req.EndDate); ok {
		end = parsed
		source = "explicit"
	}
	if parsed, ok := parseAnalyticsDate(req.Date); ok {
		reportDate = parsed
		source = "explicit"
	}
	if end.Before(start) {
		end = start
	}

	weekStart := weekStartDate(reportDate)
	return AnalyticsSummaryPeriod{
		AttendanceDate: reportDate.Format("2006-01-02"),
		WeekStart:      weekStart.Format("2006-01-02"),
		StartDate:      start.Format("2006-01-02"),
		EndDate:        end.Format("2006-01-02"),
		Source:         source,
	}
}

func previousCompletedMonthPeriod(now time.Time) AnalyticsSummaryPeriod {
	currentMonthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	start := currentMonthStart.AddDate(0, -1, 0)
	end := currentMonthStart.AddDate(0, 0, -1)
	reportDate := lastWeekday(end)
	return AnalyticsSummaryPeriod{
		AttendanceDate: reportDate.Format("2006-01-02"),
		WeekStart:      weekStartDate(reportDate).Format("2006-01-02"),
		StartDate:      start.Format("2006-01-02"),
		EndDate:        end.Format("2006-01-02"),
		Source:         "previous_completed_month_fallback",
	}
}

func hasExplicitAnalyticsPeriod(req AnalyticsSummaryRequest) bool {
	return strings.TrimSpace(req.Date) != "" ||
		strings.TrimSpace(req.StartDate) != "" ||
		strings.TrimSpace(req.EndDate) != ""
}

func parseAnalyticsDate(value string) (time.Time, bool) {
	parsed, err := time.Parse("2006-01-02", strings.TrimSpace(value))
	if err != nil {
		return time.Time{}, false
	}
	return parsed, true
}

func truncateDate(value time.Time) time.Time {
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, value.Location())
}

func weekStartDate(value time.Time) time.Time {
	date := truncateDate(value)
	weekday := int(date.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	return date.AddDate(0, 0, -(weekday - 1))
}

func lastWeekday(value time.Time) time.Time {
	date := truncateDate(value)
	for date.Weekday() == time.Saturday || date.Weekday() == time.Sunday {
		date = date.AddDate(0, 0, -1)
	}
	return date
}

func payrollTrendStart(startDate string) string {
	start, ok := parseAnalyticsDate(startDate)
	if !ok {
		return startDate
	}
	return start.AddDate(0, -5, 0).Format("2006-01-02")
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
