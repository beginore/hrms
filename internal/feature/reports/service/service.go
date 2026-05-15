package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"strconv"
	"strings"
	"time"

	"hrms/internal/feature/reports/repository"

	"github.com/google/uuid"
)

const dateLayout = "2006-01-02"

type ReportsService interface {
	GetDashboard(ctx context.Context, callerUserID uuid.UUID) (*DashboardResponse, error)
	GetPayrollSummary(ctx context.Context, callerUserID uuid.UUID, startDate, endDate string) (*PayrollSummaryResponse, error)
	GetPayrollTrends(ctx context.Context, callerUserID uuid.UUID, startDate, endDate string) (*LineChartResponse, error)
	GetDepartmentPayroll(ctx context.Context, callerUserID uuid.UUID, startDate, endDate string) (*DepartmentPayrollResponse, error)
	GetAttendanceToday(ctx context.Context, callerUserID uuid.UUID, date string) (*AttendanceTodayResponse, error)
	GetAttendanceWeekly(ctx context.Context, callerUserID uuid.UUID, weekStart string) (*LineChartResponse, error)
	GetEmployeeStatistics(ctx context.Context, callerUserID uuid.UUID) (*EmployeeStatisticsResponse, error)
	GetDepartmentStatistics(ctx context.Context, callerUserID uuid.UUID, startDate, endDate string) (*DepartmentStatisticsResponse, error)
	ExportReport(ctx context.Context, callerUserID uuid.UUID, reportType, format, startDate, endDate string) (*ExportReportResponse, error)
}

type reportsService struct {
	repo repository.ReportsRepository
}

func NewReportsService(repo repository.ReportsRepository) ReportsService {
	return &reportsService{repo: repo}
}

func (s *reportsService) GetDashboard(ctx context.Context, callerUserID uuid.UUID) (*DashboardResponse, error) {
	actor, err := s.requireReportsAccess(ctx, callerUserID)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	today := truncateDate(now)
	monthStart := time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, today.Location())
	monthEnd := monthStart.AddDate(0, 1, -1)
	previousMonthStart := monthStart.AddDate(0, -1, 0)
	previousMonthEnd := monthStart.AddDate(0, 0, -1)
	metrics, err := s.repo.GetDashboardMetrics(ctx, actor.OrgID, today, monthStart, monthEnd, previousMonthStart, previousMonthEnd)
	if err != nil {
		return nil, fmt.Errorf("%w: dashboard: %v", ErrDatabaseQueryFailed, err)
	}
	return &DashboardResponse{
		TotalEmployees:       metrics.TotalEmployees,
		ActiveEmployees:      metrics.ActiveEmployees,
		PresentToday:         metrics.PresentToday,
		LateToday:            metrics.LateToday,
		OnLeaveToday:         metrics.OnLeaveToday,
		AbsentToday:          metrics.AbsentToday,
		MonthlyPayroll:       metrics.MonthlyPayroll,
		PayrollCurrency:      metrics.PayrollCurrency,
		EmployeeGrowth:       metrics.EmployeeGrowth,
		PayrollGrowthPercent: metrics.PayrollGrowthPercent,
		AttendanceRate:       metrics.AttendanceRate,
	}, nil
}

func (s *reportsService) GetPayrollSummary(ctx context.Context, callerUserID uuid.UUID, startDate, endDate string) (*PayrollSummaryResponse, error) {
	actor, err := s.requireReportsAccess(ctx, callerUserID)
	if err != nil {
		return nil, err
	}
	start, end, err := parsePeriod(startDate, endDate)
	if err != nil {
		return nil, err
	}
	summary, err := s.repo.GetPayrollSummary(ctx, actor.OrgID, start, end)
	if err != nil {
		return nil, fmt.Errorf("%w: payroll summary: %v", ErrDatabaseQueryFailed, err)
	}
	return mapPayrollSummary(summary), nil
}

func (s *reportsService) GetPayrollTrends(ctx context.Context, callerUserID uuid.UUID, startDate, endDate string) (*LineChartResponse, error) {
	actor, err := s.requireReportsAccess(ctx, callerUserID)
	if err != nil {
		return nil, err
	}
	start, end, err := parsePeriod(startDate, endDate)
	if err != nil {
		return nil, err
	}
	points, err := s.repo.GetPayrollTrends(ctx, actor.OrgID, start, end)
	if err != nil {
		return nil, fmt.Errorf("%w: payroll trends: %v", ErrDatabaseQueryFailed, err)
	}
	labels := make([]string, 0, len(points))
	data := make([]float64, 0, len(points))
	for _, point := range points {
		labels = append(labels, point.Label)
		data = append(data, point.Value)
	}
	return &LineChartResponse{
		Labels: labels,
		Series: []ChartSeries{{Name: "Net payroll", Data: data}},
	}, nil
}

func (s *reportsService) GetDepartmentPayroll(ctx context.Context, callerUserID uuid.UUID, startDate, endDate string) (*DepartmentPayrollResponse, error) {
	actor, err := s.requireReportsAccess(ctx, callerUserID)
	if err != nil {
		return nil, err
	}
	start, end, err := parsePeriod(startDate, endDate)
	if err != nil {
		return nil, err
	}
	rows, err := s.repo.GetDepartmentPayroll(ctx, actor.OrgID, start, end)
	if err != nil {
		return nil, fmt.Errorf("%w: department payroll: %v", ErrDatabaseQueryFailed, err)
	}
	responseRows := make([]DepartmentPayrollItem, 0, len(rows))
	labels := make([]string, 0, len(rows))
	netSalary := make([]float64, 0, len(rows))
	for _, row := range rows {
		responseRows = append(responseRows, mapDepartmentPayroll(row))
		labels = append(labels, row.DepartmentName)
		netSalary = append(netSalary, parseFloat(row.NetSalaryTotal))
	}
	return &DepartmentPayrollResponse{
		Rows: responseRows,
		Chart: BarChartResponse{
			Labels: labels,
			Series: []ChartSeries{{Name: "Net payroll", Data: netSalary}},
		},
	}, nil
}

func (s *reportsService) GetAttendanceToday(ctx context.Context, callerUserID uuid.UUID, date string) (*AttendanceTodayResponse, error) {
	actor, err := s.requireReportsAccess(ctx, callerUserID)
	if err != nil {
		return nil, err
	}
	targetDate := truncateDate(time.Now())
	if strings.TrimSpace(date) != "" {
		targetDate, err = parseDate(date)
		if err != nil {
			return nil, err
		}
	}
	breakdown, err := s.repo.GetAttendanceToday(ctx, actor.OrgID, targetDate)
	if err != nil {
		return nil, fmt.Errorf("%w: attendance today: %v", ErrDatabaseQueryFailed, err)
	}
	return &AttendanceTodayResponse{
		Date:           breakdown.Date.Format(dateLayout),
		TotalActive:    breakdown.TotalActive,
		Present:        breakdown.Present,
		Late:           breakdown.Late,
		OnLeave:        breakdown.OnLeave,
		Absent:         breakdown.Absent,
		Remote:         breakdown.Remote,
		AttendanceRate: breakdown.AttendanceRate,
		Chart: PieChartResponse{Items: []PieItem{
			{Label: "Present", Value: float64(breakdown.Present)},
			{Label: "Late", Value: float64(breakdown.Late)},
			{Label: "On Leave", Value: float64(breakdown.OnLeave)},
			{Label: "Absent", Value: float64(breakdown.Absent)},
		}},
	}, nil
}

func (s *reportsService) GetAttendanceWeekly(ctx context.Context, callerUserID uuid.UUID, weekStart string) (*LineChartResponse, error) {
	actor, err := s.requireReportsAccess(ctx, callerUserID)
	if err != nil {
		return nil, err
	}
	start := truncateDate(time.Now())
	weekday := int(start.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	start = start.AddDate(0, 0, -(weekday - 1))
	if strings.TrimSpace(weekStart) != "" {
		start, err = parseDate(weekStart)
		if err != nil {
			return nil, err
		}
	}
	end := start.AddDate(0, 0, 6)
	points, err := s.repo.GetAttendanceWeekly(ctx, actor.OrgID, start, end)
	if err != nil {
		return nil, fmt.Errorf("%w: attendance weekly: %v", ErrDatabaseQueryFailed, err)
	}
	labels := make([]string, 0, len(points))
	present := make([]float64, 0, len(points))
	late := make([]float64, 0, len(points))
	onLeave := make([]float64, 0, len(points))
	absent := make([]float64, 0, len(points))
	for _, point := range points {
		labels = append(labels, point.Date.Format("Mon 02"))
		present = append(present, float64(point.Present))
		late = append(late, float64(point.Late))
		onLeave = append(onLeave, float64(point.OnLeave))
		absent = append(absent, float64(point.Absent))
	}
	return &LineChartResponse{
		Labels: labels,
		Series: []ChartSeries{
			{Name: "Present", Data: present},
			{Name: "Late", Data: late},
			{Name: "On Leave", Data: onLeave},
			{Name: "Absent", Data: absent},
		},
	}, nil
}

func (s *reportsService) GetEmployeeStatistics(ctx context.Context, callerUserID uuid.UUID) (*EmployeeStatisticsResponse, error) {
	actor, err := s.requireReportsAccess(ctx, callerUserID)
	if err != nil {
		return nil, err
	}
	stats, err := s.repo.GetEmployeeStatistics(ctx, actor.OrgID)
	if err != nil {
		return nil, fmt.Errorf("%w: employee statistics: %v", ErrDatabaseQueryFailed, err)
	}
	return &EmployeeStatisticsResponse{
		TotalEmployees:    stats.TotalEmployees,
		ActiveEmployees:   stats.ActiveEmployees,
		InactiveEmployees: stats.InactiveEmployees,
		AverageSalary:     stats.AverageSalary,
		ByDepartment:      countBarChart("Employees", stats.ByDepartment),
		ByRole:            countPieChart(stats.ByRole),
		ByStatus:          countPieChart(stats.ByStatus),
	}, nil
}

func (s *reportsService) GetDepartmentStatistics(ctx context.Context, callerUserID uuid.UUID, startDate, endDate string) (*DepartmentStatisticsResponse, error) {
	actor, err := s.requireReportsAccess(ctx, callerUserID)
	if err != nil {
		return nil, err
	}
	start, end, err := parsePeriod(startDate, endDate)
	if err != nil {
		return nil, err
	}
	rows, err := s.repo.GetDepartmentStatistics(ctx, actor.OrgID, start, end)
	if err != nil {
		return nil, fmt.Errorf("%w: department statistics: %v", ErrDatabaseQueryFailed, err)
	}
	responseRows := make([]DepartmentStatisticsItem, 0, len(rows))
	labels := make([]string, 0, len(rows))
	employees := make([]float64, 0, len(rows))
	attendanceRates := make([]float64, 0, len(rows))
	payrollCosts := make([]float64, 0, len(rows))
	for _, row := range rows {
		responseRows = append(responseRows, mapDepartmentStatistics(row))
		labels = append(labels, row.DepartmentName)
		employees = append(employees, float64(row.EmployeesCount))
		attendanceRates = append(attendanceRates, parseFloat(row.AttendanceRate))
		payrollCosts = append(payrollCosts, parseFloat(row.TotalEmployerCost))
	}
	return &DepartmentStatisticsResponse{
		Rows: responseRows,
		EmployeeChart: BarChartResponse{
			Labels: labels,
			Series: []ChartSeries{{Name: "Employees", Data: employees}},
		},
		AttendanceRateChart: BarChartResponse{
			Labels: labels,
			Series: []ChartSeries{{Name: "Attendance rate", Data: attendanceRates}},
		},
		PayrollCostChart: BarChartResponse{
			Labels: labels,
			Series: []ChartSeries{{Name: "Total employer cost", Data: payrollCosts}},
		},
	}, nil
}

func (s *reportsService) ExportReport(ctx context.Context, callerUserID uuid.UUID, reportType, format, startDate, endDate string) (*ExportReportResponse, error) {
	actor, err := s.requireReportsAccess(ctx, callerUserID)
	if err != nil {
		return nil, err
	}
	reportType = strings.ToLower(strings.TrimSpace(reportType))
	format = strings.ToLower(strings.TrimSpace(format))
	if format == "" {
		format = "csv"
	}
	if format != "csv" && format != "pdf" {
		return nil, ErrInvalidExportFormat
	}
	start, end, err := parsePeriod(startDate, endDate)
	if err != nil {
		return nil, err
	}
	switch reportType {
	case "payroll":
		if format == "pdf" {
			return s.exportPayrollPDF(ctx, actor.OrgID, start, end)
		}
		return s.exportPayroll(ctx, actor.OrgID, start, end)
	case "attendance":
		if format == "pdf" {
			return s.exportAttendancePDF(ctx, actor.OrgID, start, end)
		}
		return s.exportAttendance(ctx, actor.OrgID, start, end)
	default:
		return nil, ErrInvalidReportType
	}
}

func (s *reportsService) exportPayroll(ctx context.Context, orgID uuid.UUID, start, end time.Time) (*ExportReportResponse, error) {
	rows, err := s.repo.ListPayrollExportRows(ctx, orgID, start, end)
	if err != nil {
		return nil, fmt.Errorf("%w: payroll export: %v", ErrDatabaseQueryFailed, err)
	}
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)
	_ = writer.Write([]string{"cycle_id", "period_start", "period_end", "employee_id", "employee_name", "department", "gross_salary", "deductions", "taxes", "net_salary", "currency", "status", "review_required"})
	for _, row := range rows {
		_ = writer.Write([]string{
			row.CycleID.String(),
			row.PeriodStart.Format(dateLayout),
			row.PeriodEnd.Format(dateLayout),
			row.EmployeeID.String(),
			repository.SanitizeCSVCell(row.EmployeeName),
			repository.SanitizeCSVCell(row.DepartmentName),
			row.GrossSalary,
			row.Deductions,
			row.Taxes,
			row.NetSalary,
			row.Currency,
			row.Status,
			strconv.FormatBool(row.ReviewRequired),
		})
	}
	writer.Flush()
	return &ExportReportResponse{
		Filename:    fmt.Sprintf("payroll-report-%s-%s.csv", start.Format(dateLayout), end.Format(dateLayout)),
		ContentType: "text/csv; charset=utf-8",
		Content:     buf.Bytes(),
	}, writer.Error()
}

func (s *reportsService) exportAttendance(ctx context.Context, orgID uuid.UUID, start, end time.Time) (*ExportReportResponse, error) {
	rows, err := s.repo.ListAttendanceExportRows(ctx, orgID, start, end)
	if err != nil {
		return nil, fmt.Errorf("%w: attendance export: %v", ErrDatabaseQueryFailed, err)
	}
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)
	_ = writer.Write([]string{"date", "employee_id", "employee_name", "department", "type", "source", "status", "check_in", "check_out"})
	for _, row := range rows {
		checkIn := ""
		if row.CheckIn != nil {
			checkIn = row.CheckIn.Format(time.RFC3339)
		}
		checkOut := ""
		if row.CheckOut != nil {
			checkOut = row.CheckOut.Format(time.RFC3339)
		}
		_ = writer.Write([]string{
			row.Date.Format(dateLayout),
			row.EmployeeID.String(),
			repository.SanitizeCSVCell(row.EmployeeName),
			repository.SanitizeCSVCell(row.DepartmentName),
			row.Type,
			row.Source,
			row.Status,
			checkIn,
			checkOut,
		})
	}
	writer.Flush()
	return &ExportReportResponse{
		Filename:    fmt.Sprintf("attendance-report-%s-%s.csv", start.Format(dateLayout), end.Format(dateLayout)),
		ContentType: "text/csv; charset=utf-8",
		Content:     buf.Bytes(),
	}, writer.Error()
}

func (s *reportsService) exportPayrollPDF(ctx context.Context, orgID uuid.UUID, start, end time.Time) (*ExportReportResponse, error) {
	summary, err := s.repo.GetPayrollSummary(ctx, orgID, start, end)
	if err != nil {
		return nil, fmt.Errorf("%w: payroll pdf summary: %v", ErrDatabaseQueryFailed, err)
	}
	departments, err := s.repo.GetDepartmentPayroll(ctx, orgID, start, end)
	if err != nil {
		return nil, fmt.Errorf("%w: payroll pdf departments: %v", ErrDatabaseQueryFailed, err)
	}
	lines := []string{
		"Smart EMP Payroll Report",
		"Period: " + start.Format(dateLayout) + " - " + end.Format(dateLayout),
		"Generated: " + time.Now().Format(time.RFC3339),
		"",
		"Summary",
		"Cycles: " + strconv.Itoa(summary.CyclesCount),
		"Employees paid: " + strconv.Itoa(summary.EmployeesPaid),
		"Gross salary: " + summary.GrossSalaryTotal + " " + summary.Currency,
		"Net salary: " + summary.NetSalaryTotal + " " + summary.Currency,
		"Bonuses: " + summary.BonusesTotal + " " + summary.Currency,
		"Deductions: " + summary.DeductionsTotal + " " + summary.Currency,
		"Taxes: " + summary.TaxesTotal + " " + summary.Currency,
		"Employer taxes: " + summary.EmployerTaxesTotal + " " + summary.Currency,
		"Total employer cost: " + summary.TotalEmployerCost + " " + summary.Currency,
		"",
		"Department Payroll",
		"Department | Employees | Avg salary | Net salary | Employer cost",
	}
	for _, row := range departments {
		lines = append(lines, fmt.Sprintf(
			"%s | %d | %s %s | %s %s | %s %s",
			row.DepartmentName,
			row.EmployeesCount,
			row.AverageSalary,
			row.Currency,
			row.NetSalaryTotal,
			row.Currency,
			row.TotalEmployerCost,
			row.Currency,
		))
	}
	return &ExportReportResponse{
		Filename:    fmt.Sprintf("payroll-report-%s-%s.pdf", start.Format(dateLayout), end.Format(dateLayout)),
		ContentType: "application/pdf",
		Content:     renderReportPDF(lines),
	}, nil
}

func (s *reportsService) exportAttendancePDF(ctx context.Context, orgID uuid.UUID, start, end time.Time) (*ExportReportResponse, error) {
	rows, err := s.repo.ListAttendanceExportRows(ctx, orgID, start, end)
	if err != nil {
		return nil, fmt.Errorf("%w: attendance pdf rows: %v", ErrDatabaseQueryFailed, err)
	}
	labels := map[string]int{
		"PRESENT": 0,
		"LATE":    0,
		"ABSENT":  0,
	}
	for _, row := range rows {
		labels[row.Status]++
	}
	lines := []string{
		"Smart EMP Attendance Report",
		"Period: " + start.Format(dateLayout) + " - " + end.Format(dateLayout),
		"Generated: " + time.Now().Format(time.RFC3339),
		"",
		"Summary",
		"Records: " + strconv.Itoa(len(rows)),
		"Present records: " + strconv.Itoa(labels["PRESENT"]),
		"Late records: " + strconv.Itoa(labels["LATE"]),
		"Absent records: " + strconv.Itoa(labels["ABSENT"]),
		"",
		"Attendance Records",
		"Date | Employee | Department | Type | Source | Status",
	}
	for _, row := range rows {
		lines = append(lines, fmt.Sprintf(
			"%s | %s | %s | %s | %s | %s",
			row.Date.Format(dateLayout),
			row.EmployeeName,
			row.DepartmentName,
			row.Type,
			row.Source,
			row.Status,
		))
	}
	return &ExportReportResponse{
		Filename:    fmt.Sprintf("attendance-report-%s-%s.pdf", start.Format(dateLayout), end.Format(dateLayout)),
		ContentType: "application/pdf",
		Content:     renderReportPDF(lines),
	}, nil
}

func (s *reportsService) requireReportsAccess(ctx context.Context, callerUserID uuid.UUID) (*repository.ActorAccess, error) {
	actor, err := s.repo.GetActorAccess(ctx, callerUserID)
	if err != nil {
		return nil, fmt.Errorf("%w: actor: %v", ErrDatabaseQueryFailed, err)
	}
	if !canViewReports(actor.Role) {
		return nil, ErrForbidden
	}
	return actor, nil
}

func canViewReports(role string) bool {
	switch role {
	case "SysAdmin", "Admin", "HR":
		return true
	default:
		return false
	}
}

func parsePeriod(startDate, endDate string) (time.Time, time.Time, error) {
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	end := start.AddDate(0, 1, -1)
	var err error
	if strings.TrimSpace(startDate) != "" {
		start, err = parseDate(startDate)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
	}
	if strings.TrimSpace(endDate) != "" {
		end, err = parseDate(endDate)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
	}
	if end.Before(start) {
		return time.Time{}, time.Time{}, ErrInvalidDateRange
	}
	return start, end, nil
}

func parseDate(value string) (time.Time, error) {
	parsed, err := time.Parse(dateLayout, strings.TrimSpace(value))
	if err != nil {
		return time.Time{}, ErrInvalidDateFormat
	}
	return parsed, nil
}

func truncateDate(value time.Time) time.Time {
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, value.Location())
}

func mapPayrollSummary(summary *repository.PayrollSummary) *PayrollSummaryResponse {
	return &PayrollSummaryResponse{
		PeriodStart:        summary.PeriodStart.Format(dateLayout),
		PeriodEnd:          summary.PeriodEnd.Format(dateLayout),
		CyclesCount:        summary.CyclesCount,
		EmployeesPaid:      summary.EmployeesPaid,
		GrossSalaryTotal:   summary.GrossSalaryTotal,
		NetSalaryTotal:     summary.NetSalaryTotal,
		BonusesTotal:       summary.BonusesTotal,
		DeductionsTotal:    summary.DeductionsTotal,
		TaxesTotal:         summary.TaxesTotal,
		EmployerTaxesTotal: summary.EmployerTaxesTotal,
		TotalEmployerCost:  summary.TotalEmployerCost,
		AverageNetSalary:   summary.AverageNetSalary,
		Currency:           summary.Currency,
	}
}

func mapDepartmentPayroll(row repository.DepartmentPayrollRow) DepartmentPayrollItem {
	var id *string
	if row.DepartmentID != nil {
		value := row.DepartmentID.String()
		id = &value
	}
	return DepartmentPayrollItem{
		DepartmentID:       id,
		DepartmentName:     row.DepartmentName,
		EmployeesCount:     row.EmployeesCount,
		AverageSalary:      row.AverageSalary,
		GrossSalaryTotal:   row.GrossSalaryTotal,
		NetSalaryTotal:     row.NetSalaryTotal,
		BonusesTotal:       row.BonusesTotal,
		DeductionsTotal:    row.DeductionsTotal,
		TaxesTotal:         row.TaxesTotal,
		EmployerTaxesTotal: row.EmployerTaxesTotal,
		TotalEmployerCost:  row.TotalEmployerCost,
		Currency:           row.Currency,
	}
}

func mapDepartmentStatistics(row repository.DepartmentStatisticsRow) DepartmentStatisticsItem {
	var id *string
	if row.DepartmentID != nil {
		value := row.DepartmentID.String()
		id = &value
	}
	return DepartmentStatisticsItem{
		DepartmentID:      id,
		DepartmentName:    row.DepartmentName,
		EmployeesCount:    row.EmployeesCount,
		AverageSalary:     row.AverageSalary,
		AttendanceRate:    row.AttendanceRate,
		AbsentDays:        row.AbsentDays,
		LateDays:          row.LateDays,
		OvertimeMinutes:   row.OvertimeMinutes,
		NetSalaryTotal:    row.NetSalaryTotal,
		TotalEmployerCost: row.TotalEmployerCost,
		Currency:          row.Currency,
	}
}

func countBarChart(seriesName string, values []repository.CountByLabel) BarChartResponse {
	labels := make([]string, 0, len(values))
	data := make([]float64, 0, len(values))
	for _, value := range values {
		labels = append(labels, value.Label)
		data = append(data, float64(value.Count))
	}
	return BarChartResponse{
		Labels: labels,
		Series: []ChartSeries{{Name: seriesName, Data: data}},
	}
}

func countPieChart(values []repository.CountByLabel) PieChartResponse {
	items := make([]PieItem, 0, len(values))
	for _, value := range values {
		items = append(items, PieItem{Label: value.Label, Value: float64(value.Count)})
	}
	return PieChartResponse{Items: items}
}

func parseFloat(value string) float64 {
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0
	}
	return parsed
}
