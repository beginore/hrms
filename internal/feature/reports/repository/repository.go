package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type DBTX interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

type reportsRepository struct {
	db DBTX
}

func NewRepository(db DBTX) ReportsRepository {
	return &reportsRepository{db: db}
}

func (r *reportsRepository) GetActorAccess(ctx context.Context, userID uuid.UUID) (*ActorAccess, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, org_id, role
		FROM users
		WHERE id = $1
	`, userID)
	var actor ActorAccess
	if err := row.Scan(&actor.UserID, &actor.OrgID, &actor.Role); err != nil {
		return nil, err
	}
	return &actor, nil
}

func (r *reportsRepository) GetDashboardMetrics(ctx context.Context, orgID uuid.UUID, today, monthStart, monthEnd, previousMonthStart, previousMonthEnd time.Time) (*DashboardMetrics, error) {
	totalEmployees, err := r.countEmployees(ctx, orgID, "")
	if err != nil {
		return nil, err
	}
	activeEmployees, err := r.countEmployees(ctx, orgID, "Active")
	if err != nil {
		return nil, err
	}
	present, late, _, err := r.attendanceCountsForDate(ctx, orgID, today)
	if err != nil {
		return nil, err
	}
	onLeave, err := r.approvedLeavesCountForDate(ctx, orgID, today)
	if err != nil {
		return nil, err
	}
	absent := activeEmployees - present - late - onLeave
	if absent < 0 {
		absent = 0
	}
	currentPayroll, currency, err := r.payrollTotalForPeriod(ctx, orgID, monthStart, monthEnd)
	if err != nil {
		return nil, err
	}
	previousPayroll, _, err := r.payrollTotalForPeriod(ctx, orgID, previousMonthStart, previousMonthEnd)
	if err != nil {
		return nil, err
	}
	employeeGrowth, err := r.employeeGrowthForPeriod(ctx, orgID, monthStart, monthEnd)
	if err != nil {
		return nil, err
	}

	return &DashboardMetrics{
		TotalEmployees:       totalEmployees,
		ActiveEmployees:      activeEmployees,
		PresentToday:         present,
		LateToday:            late,
		OnLeaveToday:         onLeave,
		AbsentToday:          absent,
		MonthlyPayroll:       money(currentPayroll),
		PayrollCurrency:      currency,
		EmployeeGrowth:       employeeGrowth,
		PayrollGrowthPercent: percentChange(currentPayroll, previousPayroll),
		AttendanceRate:       percent(float64(present+late+onLeave), float64(activeEmployees)),
	}, nil
}

func (r *reportsRepository) GetPayrollSummary(ctx context.Context, orgID uuid.UUID, start, end time.Time) (*PayrollSummary, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT
			COUNT(DISTINCT pc.id),
			COUNT(DISTINCT pi.employee_id),
			COALESCE(SUM(pi.gross_salary), 0)::text,
			COALESCE(SUM(pi.net_salary), 0)::text,
			COALESCE(SUM(pi.bonuses_total), 0)::text,
			COALESCE(SUM(pi.deductions_total), 0)::text,
			COALESCE(SUM(pi.taxes_total), 0)::text,
			COALESCE(SUM(pi.employer_taxes_total), 0)::text,
			COALESCE(SUM(pi.total_employer_cost), 0)::text,
			ROUND(COALESCE(AVG(pi.net_salary), 0), 2)::text,
			COALESCE(MAX(pi.currency), 'KZT')
		FROM payroll_items pi
		JOIN payroll_cycles pc ON pc.id = pi.cycle_id
		WHERE pi.org_id = $1
		  AND pc.period_start <= $3
		  AND pc.period_end >= $2
		  AND pc.status IN ('APPROVED', 'READY_FOR_PAYMENT', 'PAID')
	`, orgID, start, end)
	summary := PayrollSummary{PeriodStart: start, PeriodEnd: end}
	if err := row.Scan(
		&summary.CyclesCount,
		&summary.EmployeesPaid,
		&summary.GrossSalaryTotal,
		&summary.NetSalaryTotal,
		&summary.BonusesTotal,
		&summary.DeductionsTotal,
		&summary.TaxesTotal,
		&summary.EmployerTaxesTotal,
		&summary.TotalEmployerCost,
		&summary.AverageNetSalary,
		&summary.Currency,
	); err != nil {
		return nil, err
	}
	return &summary, nil
}

func (r *reportsRepository) GetPayrollTrends(ctx context.Context, orgID uuid.UUID, start, end time.Time) ([]TrendPoint, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT to_char(date_trunc('month', pc.period_start), 'YYYY-MM') AS label,
		       COALESCE(SUM(pi.net_salary), 0)::float8 AS total
		FROM payroll_items pi
		JOIN payroll_cycles pc ON pc.id = pi.cycle_id
		WHERE pi.org_id = $1
		  AND pc.period_start <= $3
		  AND pc.period_end >= $2
		  AND pc.status IN ('APPROVED', 'READY_FOR_PAYMENT', 'PAID')
		GROUP BY date_trunc('month', pc.period_start)
		ORDER BY date_trunc('month', pc.period_start)
	`, orgID, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var points []TrendPoint
	for rows.Next() {
		var point TrendPoint
		if err := rows.Scan(&point.Label, &point.Value); err != nil {
			return nil, err
		}
		points = append(points, point)
	}
	return points, rows.Err()
}

func (r *reportsRepository) GetDepartmentPayroll(ctx context.Context, orgID uuid.UUID, start, end time.Time) ([]DepartmentPayrollRow, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			d.id,
			COALESCE(d.name, 'Unassigned') AS department_name,
			COUNT(DISTINCT e.id) AS employees_count,
			ROUND(COALESCE(AVG(e.salary_rate), 0), 2)::text AS average_salary,
			COALESCE(SUM(pay.gross_salary), 0)::text,
			COALESCE(SUM(pay.net_salary), 0)::text,
			COALESCE(SUM(pay.bonuses_total), 0)::text,
			COALESCE(SUM(pay.deductions_total), 0)::text,
			COALESCE(SUM(pay.taxes_total), 0)::text,
			COALESCE(SUM(pay.employer_taxes_total), 0)::text,
			COALESCE(SUM(pay.total_employer_cost), 0)::text,
			COALESCE(MAX(pay.currency), 'KZT')
		FROM employees e
		LEFT JOIN departments d ON d.id = e.department_id
		LEFT JOIN (
			SELECT
				pi.employee_id,
				SUM(pi.gross_salary) AS gross_salary,
				SUM(pi.net_salary) AS net_salary,
				SUM(pi.bonuses_total) AS bonuses_total,
				SUM(pi.deductions_total) AS deductions_total,
				SUM(pi.taxes_total) AS taxes_total,
				SUM(pi.employer_taxes_total) AS employer_taxes_total,
				SUM(pi.total_employer_cost) AS total_employer_cost,
				MAX(pi.currency) AS currency
			FROM payroll_items pi
			JOIN payroll_cycles pc ON pc.id = pi.cycle_id
			WHERE pi.org_id = $1
			  AND pc.period_start <= $3
			  AND pc.period_end >= $2
			  AND pc.status IN ('APPROVED', 'READY_FOR_PAYMENT', 'PAID')
			GROUP BY pi.employee_id
		) pay ON pay.employee_id = e.id
		WHERE e.org_id = $1
		GROUP BY d.id, COALESCE(d.name, 'Unassigned')
		ORDER BY COALESCE(d.name, 'Unassigned')
	`, orgID, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanDepartmentPayrollRows(rows)
}

func (r *reportsRepository) GetAttendanceToday(ctx context.Context, orgID uuid.UUID, today time.Time) (*AttendanceBreakdown, error) {
	activeEmployees, err := r.countEmployees(ctx, orgID, "Active")
	if err != nil {
		return nil, err
	}
	present, late, remote, err := r.attendanceCountsForDate(ctx, orgID, today)
	if err != nil {
		return nil, err
	}
	onLeave, err := r.approvedLeavesCountForDate(ctx, orgID, today)
	if err != nil {
		return nil, err
	}
	absent := activeEmployees - present - late - onLeave
	if absent < 0 {
		absent = 0
	}
	return &AttendanceBreakdown{
		Date:           today,
		TotalActive:    activeEmployees,
		Present:        present,
		Late:           late,
		OnLeave:        onLeave,
		Absent:         absent,
		Remote:         remote,
		AttendanceRate: percent(float64(present+late+onLeave), float64(activeEmployees)),
	}, nil
}

func (r *reportsRepository) GetAttendanceWeekly(ctx context.Context, orgID uuid.UUID, start, end time.Time) ([]WeeklyAttendancePoint, error) {
	activeEmployees, err := r.countEmployees(ctx, orgID, "Active")
	if err != nil {
		return nil, err
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			day::date,
			COUNT(DISTINCT CASE WHEN ar.status = 'PRESENT' THEN ar.employee_id END),
			COUNT(DISTINCT CASE WHEN ar.status = 'LATE' THEN ar.employee_id END),
			COUNT(DISTINCT CASE WHEN lr.id IS NOT NULL THEN lr.employee_id END)
		FROM generate_series($2::date, $3::date, interval '1 day') AS day
		LEFT JOIN attendance_records ar
			ON ar.org_id = $1 AND ar.date = day::date
		LEFT JOIN leave_requests lr
			ON lr.org_id = $1
			AND lr.status = 'APPROVED'
			AND lr.start_date <= day::date
			AND lr.end_date >= day::date
		GROUP BY day
		ORDER BY day
	`, orgID, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var points []WeeklyAttendancePoint
	for rows.Next() {
		var point WeeklyAttendancePoint
		if err := rows.Scan(&point.Date, &point.Present, &point.Late, &point.OnLeave); err != nil {
			return nil, err
		}
		point.Absent = activeEmployees - point.Present - point.Late - point.OnLeave
		if point.Absent < 0 {
			point.Absent = 0
		}
		points = append(points, point)
	}
	return points, rows.Err()
}

func (r *reportsRepository) GetEmployeeStatistics(ctx context.Context, orgID uuid.UUID) (*EmployeeStatistics, error) {
	stats := &EmployeeStatistics{}
	row := r.db.QueryRowContext(ctx, `
		SELECT
			COUNT(*),
			COUNT(*) FILTER (WHERE status = 'Active'),
			COUNT(*) FILTER (WHERE status <> 'Active'),
			ROUND(COALESCE(AVG(salary_rate), 0), 2)::text
		FROM employees
		WHERE org_id = $1
	`, orgID)
	if err := row.Scan(&stats.TotalEmployees, &stats.ActiveEmployees, &stats.InactiveEmployees, &stats.AverageSalary); err != nil {
		return nil, err
	}
	var err error
	stats.ByDepartment, err = r.countByLabel(ctx, `
		SELECT COALESCE(d.name, 'Unassigned'), COUNT(e.id)
		FROM employees e
		LEFT JOIN departments d ON d.id = e.department_id
		WHERE e.org_id = $1
		GROUP BY COALESCE(d.name, 'Unassigned')
		ORDER BY COALESCE(d.name, 'Unassigned')
	`, orgID)
	if err != nil {
		return nil, err
	}
	stats.ByRole, err = r.countByLabel(ctx, `
		SELECT role, COUNT(*)
		FROM employees
		WHERE org_id = $1
		GROUP BY role
		ORDER BY role
	`, orgID)
	if err != nil {
		return nil, err
	}
	stats.ByStatus, err = r.countByLabel(ctx, `
		SELECT status, COUNT(*)
		FROM employees
		WHERE org_id = $1
		GROUP BY status
		ORDER BY status
	`, orgID)
	if err != nil {
		return nil, err
	}
	return stats, nil
}

func (r *reportsRepository) GetDepartmentStatistics(ctx context.Context, orgID uuid.UUID, start, end time.Time) ([]DepartmentStatisticsRow, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			d.id,
			COALESCE(d.name, 'Unassigned') AS department_name,
			COUNT(DISTINCT e.id) AS employees_count,
			ROUND(COALESCE(AVG(e.salary_rate), 0), 2)::text AS average_salary,
			COALESCE(SUM(att.absent_days), 0)::int AS absent_days,
			COALESCE(SUM(att.late_days), 0)::int AS late_days,
			COALESCE(SUM(pay.overtime_minutes), 0)::int AS overtime_minutes,
			COALESCE(SUM(pay.net_salary), 0)::text,
			COALESCE(SUM(pay.total_employer_cost), 0)::text,
			COALESCE(MAX(pay.currency), 'KZT'),
			COALESCE(SUM(att.attended_days), 0)::float8,
			COALESCE(SUM(att.total_days), 0)::float8
		FROM employees e
		LEFT JOIN departments d ON d.id = e.department_id
		LEFT JOIN (
			SELECT
				employee_id,
				COUNT(*) FILTER (WHERE status = 'ABSENT') AS absent_days,
				COUNT(*) FILTER (WHERE status = 'LATE') AS late_days,
				COUNT(*) FILTER (WHERE status IN ('PRESENT', 'LATE', 'ON_LEAVE')) AS attended_days,
				COUNT(*) AS total_days
			FROM attendance_records
			WHERE org_id = $1
			  AND date >= $2
			  AND date <= $3
			GROUP BY employee_id
		) att ON att.employee_id = e.id
		LEFT JOIN (
			SELECT
				pi.employee_id,
				SUM(pi.overtime_minutes) AS overtime_minutes,
				SUM(pi.net_salary) AS net_salary,
				SUM(pi.total_employer_cost) AS total_employer_cost,
				MAX(pi.currency) AS currency
			FROM payroll_items pi
			JOIN payroll_cycles pc ON pc.id = pi.cycle_id
			WHERE pi.org_id = $1
			  AND pc.period_start <= $3
			  AND pc.period_end >= $2
			  AND pc.status IN ('APPROVED', 'READY_FOR_PAYMENT', 'PAID')
			GROUP BY pi.employee_id
		) pay ON pay.employee_id = e.id
		WHERE e.org_id = $1
		GROUP BY d.id, COALESCE(d.name, 'Unassigned')
		ORDER BY COALESCE(d.name, 'Unassigned')
	`, orgID, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rowsOut []DepartmentStatisticsRow
	for rows.Next() {
		var row DepartmentStatisticsRow
		var departmentID uuid.NullUUID
		var attended, attendanceTotal float64
		if err := rows.Scan(
			&departmentID,
			&row.DepartmentName,
			&row.EmployeesCount,
			&row.AverageSalary,
			&row.AbsentDays,
			&row.LateDays,
			&row.OvertimeMinutes,
			&row.NetSalaryTotal,
			&row.TotalEmployerCost,
			&row.Currency,
			&attended,
			&attendanceTotal,
		); err != nil {
			return nil, err
		}
		if departmentID.Valid {
			id := departmentID.UUID
			row.DepartmentID = &id
		}
		row.AttendanceRate = percent(attended, attendanceTotal)
		rowsOut = append(rowsOut, row)
	}
	return rowsOut, rows.Err()
}

func (r *reportsRepository) ListPayrollExportRows(ctx context.Context, orgID uuid.UUID, start, end time.Time) ([]PayrollExportRow, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			pc.id,
			pc.period_start,
			pc.period_end,
			e.id,
			trim(u.first_name || ' ' || u.last_name) AS employee_name,
			COALESCE(d.name, 'Unassigned') AS department_name,
			pi.gross_salary::text,
			pi.deductions_total::text,
			pi.taxes_total::text,
			pi.net_salary::text,
			pi.currency,
			pi.status,
			pi.review_required
		FROM payroll_items pi
		JOIN payroll_cycles pc ON pc.id = pi.cycle_id
		JOIN employees e ON e.id = pi.employee_id
		JOIN users u ON u.id = e.user_id
		LEFT JOIN departments d ON d.id = e.department_id
		WHERE pi.org_id = $1
		  AND pc.period_start <= $3
		  AND pc.period_end >= $2
		  AND pc.status IN ('APPROVED', 'READY_FOR_PAYMENT', 'PAID')
		ORDER BY pc.period_start DESC, department_name, employee_name
	`, orgID, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []PayrollExportRow
	for rows.Next() {
		var row PayrollExportRow
		if err := rows.Scan(
			&row.CycleID,
			&row.PeriodStart,
			&row.PeriodEnd,
			&row.EmployeeID,
			&row.EmployeeName,
			&row.DepartmentName,
			&row.GrossSalary,
			&row.Deductions,
			&row.Taxes,
			&row.NetSalary,
			&row.Currency,
			&row.Status,
			&row.ReviewRequired,
		); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

func (r *reportsRepository) ListAttendanceExportRows(ctx context.Context, orgID uuid.UUID, start, end time.Time) ([]AttendanceExportRow, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			ar.date,
			e.id,
			trim(u.first_name || ' ' || u.last_name) AS employee_name,
			COALESCE(d.name, 'Unassigned') AS department_name,
			ar.type,
			ar.source,
			ar.status,
			ar.check_in,
			ar.check_out
		FROM attendance_records ar
		JOIN employees e ON e.id = ar.employee_id
		JOIN users u ON u.id = e.user_id
		LEFT JOIN departments d ON d.id = e.department_id
		WHERE ar.org_id = $1
		  AND ar.date >= $2
		  AND ar.date <= $3
		ORDER BY ar.date DESC, department_name, employee_name
	`, orgID, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []AttendanceExportRow
	for rows.Next() {
		var row AttendanceExportRow
		var checkIn, checkOut sql.NullTime
		if err := rows.Scan(
			&row.Date,
			&row.EmployeeID,
			&row.EmployeeName,
			&row.DepartmentName,
			&row.Type,
			&row.Source,
			&row.Status,
			&checkIn,
			&checkOut,
		); err != nil {
			return nil, err
		}
		if checkIn.Valid {
			row.CheckIn = &checkIn.Time
		}
		if checkOut.Valid {
			row.CheckOut = &checkOut.Time
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

func (r *reportsRepository) countEmployees(ctx context.Context, orgID uuid.UUID, status string) (int, error) {
	query := `SELECT COUNT(*) FROM employees WHERE org_id = $1`
	args := []any{orgID}
	if status != "" {
		query += ` AND status = $2`
		args = append(args, status)
	}
	row := r.db.QueryRowContext(ctx, query, args...)
	var count int
	return count, row.Scan(&count)
}

func (r *reportsRepository) attendanceCountsForDate(ctx context.Context, orgID uuid.UUID, date time.Time) (present, late, remote int, err error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT
			COUNT(DISTINCT CASE WHEN status = 'PRESENT' THEN employee_id END),
			COUNT(DISTINCT CASE WHEN status = 'LATE' THEN employee_id END),
			COUNT(DISTINCT CASE WHEN type = 'REMOTE' THEN employee_id END)
		FROM attendance_records
		WHERE org_id = $1 AND date = $2
	`, orgID, date)
	err = row.Scan(&present, &late, &remote)
	return present, late, remote, err
}

func (r *reportsRepository) approvedLeavesCountForDate(ctx context.Context, orgID uuid.UUID, date time.Time) (int, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT COUNT(DISTINCT employee_id)
		FROM leave_requests
		WHERE org_id = $1
		  AND status = 'APPROVED'
		  AND start_date <= $2
		  AND end_date >= $2
	`, orgID, date)
	var count int
	return count, row.Scan(&count)
}

func (r *reportsRepository) payrollTotalForPeriod(ctx context.Context, orgID uuid.UUID, start, end time.Time) (float64, string, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(pi.net_salary), 0)::float8, COALESCE(MAX(pi.currency), 'KZT')
		FROM payroll_items pi
		JOIN payroll_cycles pc ON pc.id = pi.cycle_id
		WHERE pi.org_id = $1
		  AND pc.period_start <= $3
		  AND pc.period_end >= $2
		  AND pc.status IN ('APPROVED', 'READY_FOR_PAYMENT', 'PAID')
	`, orgID, start, end)
	var total float64
	var currency string
	return total, currency, row.Scan(&total, &currency)
}

func (r *reportsRepository) employeeGrowthForPeriod(ctx context.Context, orgID uuid.UUID, start, end time.Time) (int, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM users u
		JOIN employees e ON e.user_id = u.id
		WHERE e.org_id = $1
		  AND u.created_at >= $2::date
		  AND u.created_at < ($3::date + interval '1 day')
	`, orgID, start, end)
	var count int
	return count, row.Scan(&count)
}

func (r *reportsRepository) countByLabel(ctx context.Context, query string, orgID uuid.UUID) ([]CountByLabel, error) {
	rows, err := r.db.QueryContext(ctx, query, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []CountByLabel
	for rows.Next() {
		var item CountByLabel
		if err := rows.Scan(&item.Label, &item.Count); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func scanDepartmentPayrollRows(rows *sql.Rows) ([]DepartmentPayrollRow, error) {
	var out []DepartmentPayrollRow
	for rows.Next() {
		var row DepartmentPayrollRow
		var departmentID uuid.NullUUID
		if err := rows.Scan(
			&departmentID,
			&row.DepartmentName,
			&row.EmployeesCount,
			&row.AverageSalary,
			&row.GrossSalaryTotal,
			&row.NetSalaryTotal,
			&row.BonusesTotal,
			&row.DeductionsTotal,
			&row.TaxesTotal,
			&row.EmployerTaxesTotal,
			&row.TotalEmployerCost,
			&row.Currency,
		); err != nil {
			return nil, err
		}
		if departmentID.Valid {
			id := departmentID.UUID
			row.DepartmentID = &id
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

func money(value float64) string {
	return fmt.Sprintf("%.2f", value)
}

func percent(numerator, denominator float64) string {
	if denominator <= 0 {
		return "0.00"
	}
	return fmt.Sprintf("%.2f", numerator/denominator*100)
}

func percentChange(current, previous float64) string {
	if previous == 0 {
		if current == 0 {
			return "0.00"
		}
		return "100.00"
	}
	return fmt.Sprintf("%.2f", (current-previous)/previous*100)
}

func SanitizeCSVCell(value string) string {
	if value == "" {
		return value
	}
	if strings.ContainsAny(value[:1], "=+-@") {
		return "'" + value
	}
	return value
}
