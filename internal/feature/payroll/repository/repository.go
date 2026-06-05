package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type payrollRepository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) PayrollRepository {
	return &payrollRepository{db: db}
}

func (r *payrollRepository) GetOrgIDByUserID(ctx context.Context, userID uuid.UUID) (uuid.UUID, error) {
	var orgID uuid.UUID
	err := r.db.QueryRowContext(ctx, `SELECT org_id FROM users WHERE id = $1`, userID).Scan(&orgID)
	return orgID, err
}

func (r *payrollRepository) GetEmployeeIDByUserID(ctx context.Context, userID uuid.UUID) (uuid.UUID, error) {
	var employeeID uuid.UUID
	err := r.db.QueryRowContext(ctx, `SELECT id FROM employees WHERE user_id = $1`, userID).Scan(&employeeID)
	return employeeID, err
}

func (r *payrollRepository) GetActorAccess(ctx context.Context, userID uuid.UUID) (*ActorAccess, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT u.id, u.org_id,
		       COALESCE(e.id, '00000000-0000-0000-0000-000000000000'::uuid),
		       COALESCE(e.role, u.role)
		FROM users u
		LEFT JOIN employees e ON e.user_id = u.id
		WHERE u.id = $1
	`, userID)
	var access ActorAccess
	if err := row.Scan(&access.UserID, &access.OrgID, &access.EmployeeID, &access.Role); err != nil {
		return nil, err
	}
	return &access, nil
}

func (r *payrollRepository) CreateCycle(ctx context.Context, params CreateCycleParams) (*PayrollCycle, error) {
	row := r.db.QueryRowContext(ctx, `
		INSERT INTO payroll_cycles (id, org_id, period_start, period_end, currency, created_by)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, org_id, period_start, period_end, currency, status, created_by, approved_by, approved_at, paid_at, created_at, updated_at
	`, params.ID, params.OrgID, params.PeriodStart, params.PeriodEnd, params.Currency, params.CreatedBy)
	return scanCycle(row)
}

func (r *payrollRepository) GetCycle(ctx context.Context, id uuid.UUID) (*PayrollCycle, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, org_id, period_start, period_end, currency, status, created_by, approved_by, approved_at, paid_at, created_at, updated_at
		FROM payroll_cycles
		WHERE id = $1
	`, id)
	return scanCycle(row)
}

func (r *payrollRepository) ListCycles(ctx context.Context, orgID uuid.UUID) ([]PayrollCycle, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, org_id, period_start, period_end, currency, status, created_by, approved_by, approved_at, paid_at, created_at, updated_at
		FROM payroll_cycles
		WHERE org_id = $1
		ORDER BY period_start DESC
	`, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cycles []PayrollCycle
	for rows.Next() {
		cycle, err := scanCycle(rows)
		if err != nil {
			return nil, err
		}
		cycles = append(cycles, *cycle)
	}
	return cycles, rows.Err()
}

func (r *payrollRepository) UpdateCycleStatus(ctx context.Context, id uuid.UUID, status string, actorUserID uuid.UUID) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `
		UPDATE payroll_cycles
		SET status = $2::varchar,
			approved_by = CASE
				WHEN $2::varchar = 'APPROVED' THEN $3
				WHEN $2::varchar = 'CALCULATED' THEN NULL
				ELSE approved_by
			END,
			approved_at = CASE
				WHEN $2::varchar = 'APPROVED' THEN now()
				WHEN $2::varchar = 'CALCULATED' THEN NULL
				ELSE approved_at
			END,
			updated_at = now()
		WHERE id = $1
	`, id, status, actorUserID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE payroll_items
		SET status = $2::varchar, updated_at = now()
		WHERE cycle_id = $1
	`, id, status); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *payrollRepository) MarkCyclePaid(ctx context.Context, id uuid.UUID) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `
		UPDATE payroll_cycles
		SET status = 'PAID', paid_at = now(), updated_at = now()
		WHERE id = $1
	`, id); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE payroll_items
		SET status = 'PAID', updated_at = now()
		WHERE cycle_id = $1
	`, id); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *payrollRepository) ListActiveEmployees(ctx context.Context, orgID uuid.UUID) ([]EmployeePayrollInput, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, org_id, user_id, salary_rate::text, status, hire_date, termination_date
		FROM employees
		WHERE org_id = $1 AND LOWER(status) = 'active'
		ORDER BY id
	`, orgID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var employees []EmployeePayrollInput
	for rows.Next() {
		employee, err := scanEmployeePayrollInput(rows)
		if err != nil {
			return nil, err
		}
		employees = append(employees, *employee)
	}
	return employees, rows.Err()
}

func (r *payrollRepository) GetEmployee(ctx context.Context, employeeID uuid.UUID) (*EmployeePayrollInput, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, org_id, user_id, salary_rate::text, status, hire_date, termination_date
		FROM employees
		WHERE id = $1
	`, employeeID)
	return scanEmployeePayrollInput(row)
}

func (r *payrollRepository) ListAttendance(ctx context.Context, employeeID uuid.UUID, start, end time.Time) ([]AttendanceRecord, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, employee_id, date, type, source, status, check_in, check_out, note
		FROM attendance_records
		WHERE employee_id = $1 AND date >= $2 AND date <= $3
		ORDER BY date
	`, employeeID, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []AttendanceRecord
	for rows.Next() {
		record, err := scanAttendanceRecord(rows)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, rows.Err()
}

func (r *payrollRepository) ListApprovedLeaves(ctx context.Context, employeeID uuid.UUID, start, end time.Time) ([]ApprovedLeave, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, type, start_date, end_date
		FROM leave_requests
		WHERE employee_id = $1
		  AND status = 'APPROVED'
		  AND start_date <= $3
		  AND end_date >= $2
		ORDER BY start_date
	`, employeeID, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var leaves []ApprovedLeave
	for rows.Next() {
		var leave ApprovedLeave
		if err := rows.Scan(&leave.ID, &leave.Type, &leave.StartDate, &leave.EndDate); err != nil {
			return nil, err
		}
		leaves = append(leaves, leave)
	}
	return leaves, rows.Err()
}

func (r *payrollRepository) ListSalaryHistory(ctx context.Context, employeeID uuid.UUID, start, end time.Time) ([]SalaryHistoryItem, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, employee_id, salary_rate::text, effective_from, effective_to
		FROM salary_history
		WHERE employee_id = $1
		  AND effective_from <= $3
		  AND (effective_to IS NULL OR effective_to >= $2)
		ORDER BY effective_from
	`, employeeID, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []SalaryHistoryItem
	for rows.Next() {
		item, err := scanSalaryHistoryItem(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *item)
	}
	return items, rows.Err()
}

func (r *payrollRepository) CreateSalaryHistory(ctx context.Context, params CreateSalaryHistoryParams) (*SalaryHistoryItem, error) {
	row := r.db.QueryRowContext(ctx, `
		INSERT INTO salary_history (id, org_id, employee_id, salary_rate, effective_from, effective_to, created_by)
		VALUES ($1, $2, $3, $4::numeric, $5, $6, $7)
		RETURNING id, employee_id, salary_rate::text, effective_from, effective_to
	`, params.ID, params.OrgID, params.EmployeeID, params.SalaryRate, params.EffectiveFrom, params.EffectiveTo, params.CreatedBy)
	return scanSalaryHistoryItem(row)
}

func (r *payrollRepository) ListAttendanceOverrides(ctx context.Context, employeeID uuid.UUID, start, end time.Time) ([]AttendanceOverride, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, employee_id, date, paid_day, attendance_type, attendance_status, overtime_minutes, note
		FROM payroll_attendance_overrides
		WHERE employee_id = $1 AND date >= $2 AND date <= $3
		ORDER BY date
	`, employeeID, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var overrides []AttendanceOverride
	for rows.Next() {
		override, err := scanAttendanceOverride(rows)
		if err != nil {
			return nil, err
		}
		overrides = append(overrides, *override)
	}
	return overrides, rows.Err()
}

func (r *payrollRepository) UpsertAttendanceOverride(ctx context.Context, params UpsertAttendanceOverrideParams) (*AttendanceOverride, error) {
	row := r.db.QueryRowContext(ctx, `
		INSERT INTO payroll_attendance_overrides (
			id, org_id, employee_id, date, paid_day, attendance_type,
			attendance_status, overtime_minutes, note, created_by
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (employee_id, date) DO UPDATE
			SET paid_day = EXCLUDED.paid_day,
				attendance_type = EXCLUDED.attendance_type,
				attendance_status = EXCLUDED.attendance_status,
				overtime_minutes = EXCLUDED.overtime_minutes,
				note = EXCLUDED.note,
				created_by = EXCLUDED.created_by,
				created_at = now()
		RETURNING id, employee_id, date, paid_day, attendance_type, attendance_status, overtime_minutes, note
	`, params.ID, params.OrgID, params.EmployeeID, params.Date, params.PaidDay, params.AttendanceType,
		params.AttendanceStatus, params.OvertimeMinutes, nullableString(params.Note), params.CreatedBy)
	return scanAttendanceOverride(row)
}

func (r *payrollRepository) UpdateEmploymentPeriod(ctx context.Context, params UpdateEmploymentPeriodParams) (*EmployeePayrollInput, error) {
	row := r.db.QueryRowContext(ctx, `
		UPDATE employees
		SET hire_date = $3,
			termination_date = $4
		WHERE id = $1 AND org_id = $2
		RETURNING id, org_id, user_id, salary_rate::text, status, hire_date, termination_date
	`, params.EmployeeID, params.OrgID, params.HireDate, params.TerminationDate)
	return scanEmployeePayrollInput(row)
}

func (r *payrollRepository) GetWorkSchedule(ctx context.Context, employeeID uuid.UUID) (*WorkSchedule, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT employee_id, work_start::text, work_end::text, late_threshold_minutes
		FROM work_schedules
		WHERE employee_id = $1
	`, employeeID)
	var schedule WorkSchedule
	if err := row.Scan(&schedule.EmployeeID, &schedule.WorkStart, &schedule.WorkEnd, &schedule.LateThresholdMinutes); err != nil {
		return nil, err
	}
	return &schedule, nil
}

func (r *payrollRepository) ListCalendarDays(ctx context.Context, orgID uuid.UUID, start, end time.Time) ([]CalendarDay, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT date, type, name
		FROM working_calendar
		WHERE org_id = $1 AND date >= $2 AND date <= $3
		ORDER BY date
	`, orgID, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var days []CalendarDay
	for rows.Next() {
		var day CalendarDay
		var name sql.NullString
		if err := rows.Scan(&day.Date, &day.Type, &name); err != nil {
			return nil, err
		}
		if name.Valid {
			day.Name = name.String
		}
		days = append(days, day)
	}
	return days, rows.Err()
}

func (r *payrollRepository) GetPolicy(ctx context.Context, orgID uuid.UUID) (*PayrollPolicy, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, org_id, missing_attendance_policy,
			regular_overtime_multiplier::text, holiday_overtime_multiplier::text,
			late_penalty_mode, late_penalty_amount::text, rounding_mode,
			vacation_pay_rate::text, sick_leave_pay_rate::text, remote_pay_rate::text,
			business_trip_pay_rate::text, unpaid_leave_pay_rate::text
		FROM payroll_policies
		WHERE org_id = $1
	`, orgID)
	return scanPolicy(row)
}

func (r *payrollRepository) UpsertPolicy(ctx context.Context, params UpsertPolicyParams) (*PayrollPolicy, error) {
	row := r.db.QueryRowContext(ctx, `
		INSERT INTO payroll_policies (
			id, org_id, missing_attendance_policy, regular_overtime_multiplier,
			holiday_overtime_multiplier, late_penalty_mode, late_penalty_amount,
			rounding_mode, vacation_pay_rate, sick_leave_pay_rate, remote_pay_rate,
			business_trip_pay_rate, unpaid_leave_pay_rate, updated_at
		)
		VALUES (
			$1, $2, $3, $4::numeric, $5::numeric, $6, $7::numeric, $8,
			$9::numeric, $10::numeric, $11::numeric, $12::numeric, $13::numeric, now()
		)
		ON CONFLICT (org_id) DO UPDATE
			SET missing_attendance_policy = EXCLUDED.missing_attendance_policy,
				regular_overtime_multiplier = EXCLUDED.regular_overtime_multiplier,
				holiday_overtime_multiplier = EXCLUDED.holiday_overtime_multiplier,
				late_penalty_mode = EXCLUDED.late_penalty_mode,
				late_penalty_amount = EXCLUDED.late_penalty_amount,
				rounding_mode = EXCLUDED.rounding_mode,
				vacation_pay_rate = EXCLUDED.vacation_pay_rate,
				sick_leave_pay_rate = EXCLUDED.sick_leave_pay_rate,
				remote_pay_rate = EXCLUDED.remote_pay_rate,
				business_trip_pay_rate = EXCLUDED.business_trip_pay_rate,
				unpaid_leave_pay_rate = EXCLUDED.unpaid_leave_pay_rate,
				updated_at = now()
		RETURNING id, org_id, missing_attendance_policy,
			regular_overtime_multiplier::text, holiday_overtime_multiplier::text,
			late_penalty_mode, late_penalty_amount::text, rounding_mode,
			vacation_pay_rate::text, sick_leave_pay_rate::text, remote_pay_rate::text,
			business_trip_pay_rate::text, unpaid_leave_pay_rate::text
	`, uuid.New(), params.OrgID, params.MissingAttendancePolicy, params.RegularOvertimeMultiplier,
		params.HolidayOvertimeMultiplier, params.LatePenaltyMode, params.LatePenaltyAmount,
		params.RoundingMode, params.VacationPayRate, params.SickLeavePayRate, params.RemotePayRate,
		params.BusinessTripPayRate, params.UnpaidLeavePayRate)
	return scanPolicy(row)
}

func (r *payrollRepository) CreateAdjustment(ctx context.Context, params CreateAdjustmentParams) (*PayrollAdjustment, error) {
	row := r.db.QueryRowContext(ctx, `
		INSERT INTO payroll_adjustments (id, org_id, employee_id, cycle_id, type, category, amount, currency, is_taxable, reason, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7::numeric, $8, $9, $10, $11)
		RETURNING id, org_id, employee_id, cycle_id, type, category, amount::text, currency, is_taxable, reason, created_by, created_at
	`, params.ID, params.OrgID, params.EmployeeID, params.CycleID, params.Type, params.Category, params.Amount, params.Currency, params.IsTaxable, nullableString(params.Reason), params.CreatedBy)
	return scanAdjustment(row)
}

func (r *payrollRepository) ListAdjustments(ctx context.Context, orgID uuid.UUID, cycleID *uuid.UUID, employeeID *uuid.UUID) ([]PayrollAdjustment, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, org_id, employee_id, cycle_id, type, category, amount::text, currency, is_taxable, reason, created_by, created_at
		FROM payroll_adjustments
		WHERE org_id = $1
		  AND ($2::uuid IS NULL OR cycle_id = $2)
		  AND ($3::uuid IS NULL OR employee_id = $3)
		ORDER BY created_at DESC
	`, orgID, cycleID, employeeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var adjustments []PayrollAdjustment
	for rows.Next() {
		adjustment, err := scanAdjustment(rows)
		if err != nil {
			return nil, err
		}
		adjustments = append(adjustments, *adjustment)
	}
	return adjustments, rows.Err()
}

func (r *payrollRepository) DeleteAdjustment(ctx context.Context, id uuid.UUID, orgID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM payroll_adjustments WHERE id = $1 AND org_id = $2`, id, orgID)
	return err
}

func (r *payrollRepository) CreateTaxRule(ctx context.Context, params CreateTaxRuleParams) (*PayrollTaxRule, error) {
	row := r.db.QueryRowContext(ctx, `
		INSERT INTO payroll_tax_rules (id, org_id, country, name, rate, applies_to, payer, threshold_min, threshold_max, effective_from, effective_to)
		VALUES ($1, $2, $3, $4, $5::numeric, $6, $7, $8::numeric, $9::numeric, $10, $11)
		RETURNING id, org_id, country, name, rate::text, applies_to, payer, threshold_min::text, threshold_max::text, is_active, effective_from, effective_to, created_at, updated_at
	`, params.ID, params.OrgID, params.Country, params.Name, params.Rate, params.AppliesTo, params.Payer, nullableNumeric(params.ThresholdMin), nullableNumeric(params.ThresholdMax), params.EffectiveFrom, params.EffectiveTo)
	return scanTaxRule(row)
}

func (r *payrollRepository) ListActiveTaxRules(ctx context.Context, orgID uuid.UUID, asOf time.Time) ([]PayrollTaxRule, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, org_id, country, name, rate::text, applies_to, payer, threshold_min::text, threshold_max::text, is_active, effective_from, effective_to, created_at, updated_at
		FROM payroll_tax_rules
		WHERE is_active = true
		  AND (org_id = $1 OR org_id IS NULL)
		  AND effective_from <= $2
		  AND (effective_to IS NULL OR effective_to >= $2)
		ORDER BY org_id NULLS LAST, effective_from DESC
	`, orgID, asOf)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []PayrollTaxRule
	for rows.Next() {
		rule, err := scanTaxRule(rows)
		if err != nil {
			return nil, err
		}
		rules = append(rules, *rule)
	}
	return rules, rows.Err()
}

func (r *payrollRepository) UpsertPayrollItem(ctx context.Context, params UpsertPayrollItemParams) (*PayrollItem, error) {
	row := r.db.QueryRowContext(ctx, `
		INSERT INTO payroll_items (
			id, cycle_id, org_id, employee_id, base_salary, attendance_adjustment,
			overtime_amount, bonuses_total, deductions_total, taxes_total, gross_salary,
			net_salary, currency, working_days, paid_days, unpaid_days, late_days, absent_days,
			missing_days, overtime_minutes, review_required, review_reasons,
			employer_taxes_total, total_employer_cost, status, calculation_snapshot, updated_at
		)
		VALUES (
			$1, $2, $3, $4, $5::numeric, $6::numeric,
			$7::numeric, $8::numeric, $9::numeric, $10::numeric, $11::numeric,
			$12::numeric, $13, $14, $15::numeric, $16::numeric, $17, $18,
			$19, $20, $21, $22::jsonb, $23::numeric, $24::numeric, $25, $26::jsonb, now()
		)
		ON CONFLICT (cycle_id, employee_id) DO UPDATE
			SET base_salary = EXCLUDED.base_salary,
				attendance_adjustment = EXCLUDED.attendance_adjustment,
				overtime_amount = EXCLUDED.overtime_amount,
				bonuses_total = EXCLUDED.bonuses_total,
				deductions_total = EXCLUDED.deductions_total,
				taxes_total = EXCLUDED.taxes_total,
				gross_salary = EXCLUDED.gross_salary,
				net_salary = EXCLUDED.net_salary,
				currency = EXCLUDED.currency,
				working_days = EXCLUDED.working_days,
				paid_days = EXCLUDED.paid_days,
				unpaid_days = EXCLUDED.unpaid_days,
				late_days = EXCLUDED.late_days,
				absent_days = EXCLUDED.absent_days,
				missing_days = EXCLUDED.missing_days,
				overtime_minutes = EXCLUDED.overtime_minutes,
				review_required = EXCLUDED.review_required,
				review_reasons = EXCLUDED.review_reasons,
				employer_taxes_total = EXCLUDED.employer_taxes_total,
				total_employer_cost = EXCLUDED.total_employer_cost,
				status = EXCLUDED.status,
				calculation_snapshot = EXCLUDED.calculation_snapshot,
				updated_at = now()
		RETURNING id, cycle_id, org_id, employee_id, base_salary::text, attendance_adjustment::text,
			overtime_amount::text, bonuses_total::text, deductions_total::text, taxes_total::text,
			gross_salary::text, net_salary::text, currency, working_days, paid_days::text, unpaid_days::text,
			late_days, absent_days, missing_days, overtime_minutes, review_required, review_reasons,
			employer_taxes_total::text, total_employer_cost::text, status, reviewed_by, reviewed_at,
			review_comment, calculation_snapshot, created_at, updated_at
	`, params.ID, params.CycleID, params.OrgID, params.EmployeeID, params.BaseSalary, params.AttendanceAdjustment,
		params.OvertimeAmount, params.BonusesTotal, params.DeductionsTotal, params.TaxesTotal, params.GrossSalary,
		params.NetSalary, params.Currency, params.WorkingDays, params.PaidDays, params.UnpaidDays, params.LateDays, params.AbsentDays,
		params.MissingDays, params.OvertimeMinutes, params.ReviewRequired, params.ReviewReasons,
		params.EmployerTaxesTotal, params.TotalEmployerCost, params.Status, params.CalculationSnapshot)
	return scanPayrollItem(row)
}

func (r *payrollRepository) ListPayrollItems(ctx context.Context, cycleID uuid.UUID) ([]PayrollItem, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, cycle_id, org_id, employee_id, base_salary::text, attendance_adjustment::text,
			overtime_amount::text, bonuses_total::text, deductions_total::text, taxes_total::text,
			gross_salary::text, net_salary::text, currency, working_days, paid_days::text, unpaid_days::text,
			late_days, absent_days, missing_days, overtime_minutes, review_required, review_reasons,
			employer_taxes_total::text, total_employer_cost::text, status, reviewed_by, reviewed_at,
			review_comment, calculation_snapshot, created_at, updated_at
		FROM payroll_items
		WHERE cycle_id = $1
		ORDER BY employee_id
	`, cycleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []PayrollItem
	for rows.Next() {
		item, err := scanPayrollItem(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, *item)
	}
	return items, rows.Err()
}

func (r *payrollRepository) GetPayrollItem(ctx context.Context, id uuid.UUID) (*PayrollItem, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, cycle_id, org_id, employee_id, base_salary::text, attendance_adjustment::text,
			overtime_amount::text, bonuses_total::text, deductions_total::text, taxes_total::text,
			gross_salary::text, net_salary::text, currency, working_days, paid_days::text, unpaid_days::text,
			late_days, absent_days, missing_days, overtime_minutes, review_required, review_reasons,
			employer_taxes_total::text, total_employer_cost::text, status, reviewed_by, reviewed_at,
			review_comment, calculation_snapshot, created_at, updated_at
		FROM payroll_items
		WHERE id = $1
	`, id)
	return scanPayrollItem(row)
}

func (r *payrollRepository) GetPayrollItemByCycleEmployee(ctx context.Context, cycleID, employeeID uuid.UUID) (*PayrollItem, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, cycle_id, org_id, employee_id, base_salary::text, attendance_adjustment::text,
			overtime_amount::text, bonuses_total::text, deductions_total::text, taxes_total::text,
			gross_salary::text, net_salary::text, currency, working_days, paid_days::text, unpaid_days::text,
			late_days, absent_days, missing_days, overtime_minutes, review_required, review_reasons,
			employer_taxes_total::text, total_employer_cost::text, status, reviewed_by, reviewed_at,
			review_comment, calculation_snapshot, created_at, updated_at
		FROM payroll_items
		WHERE cycle_id = $1 AND employee_id = $2
	`, cycleID, employeeID)
	return scanPayrollItem(row)
}

func (r *payrollRepository) ReviewPayrollItem(ctx context.Context, id uuid.UUID, reviewedBy uuid.UUID, comment string) (*PayrollItem, error) {
	row := r.db.QueryRowContext(ctx, `
		UPDATE payroll_items
		SET review_required = false,
			reviewed_by = $2,
			reviewed_at = now(),
			review_comment = $3,
			updated_at = now()
		WHERE id = $1
		RETURNING id, cycle_id, org_id, employee_id, base_salary::text, attendance_adjustment::text,
			overtime_amount::text, bonuses_total::text, deductions_total::text, taxes_total::text,
			gross_salary::text, net_salary::text, currency, working_days, paid_days::text, unpaid_days::text,
			late_days, absent_days, missing_days, overtime_minutes, review_required, review_reasons,
			employer_taxes_total::text, total_employer_cost::text, status, reviewed_by, reviewed_at,
			review_comment, calculation_snapshot, created_at, updated_at
	`, id, reviewedBy, nullableString(comment))
	return scanPayrollItem(row)
}

func (r *payrollRepository) CreateCorrection(ctx context.Context, params CreateCorrectionParams) (*PayrollCorrection, error) {
	row := r.db.QueryRowContext(ctx, `
		INSERT INTO payroll_corrections (
			id, org_id, employee_id, source_item_id, target_cycle_id,
			adjustment_id, amount, type, reason, created_by
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7::numeric, $8, $9, $10)
		RETURNING id, org_id, employee_id, source_item_id, target_cycle_id,
			adjustment_id, amount::text, type, reason, created_by, created_at
	`, params.ID, params.OrgID, params.EmployeeID, params.SourceItemID, params.TargetCycleID,
		params.AdjustmentID, params.Amount, params.Type, params.Reason, params.CreatedBy)
	return scanCorrection(row)
}

func (r *payrollRepository) InsertAuditLog(ctx context.Context, orgID uuid.UUID, cycleID *uuid.UUID, itemID *uuid.UUID, actorUserID uuid.UUID, action string, before, after []byte) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO payroll_audit_logs (id, org_id, cycle_id, payroll_item_id, actor_user_id, action, before_state, after_state)
		VALUES ($1, $2, $3, $4, $5, $6, $7::jsonb, $8::jsonb)
	`, uuid.New(), orgID, cycleID, itemID, actorUserID, action, nullableJSON(before), nullableJSON(after))
	return err
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanCycle(row rowScanner) (*PayrollCycle, error) {
	var cycle PayrollCycle
	var approvedBy sql.NullString
	var approvedAt, paidAt sql.NullTime
	if err := row.Scan(&cycle.ID, &cycle.OrgID, &cycle.PeriodStart, &cycle.PeriodEnd, &cycle.Currency, &cycle.Status, &cycle.CreatedBy, &approvedBy, &approvedAt, &paidAt, &cycle.CreatedAt, &cycle.UpdatedAt); err != nil {
		return nil, err
	}
	if approvedBy.Valid {
		id, err := uuid.Parse(approvedBy.String)
		if err != nil {
			return nil, err
		}
		cycle.ApprovedBy = &id
	}
	if approvedAt.Valid {
		cycle.ApprovedAt = &approvedAt.Time
	}
	if paidAt.Valid {
		cycle.PaidAt = &paidAt.Time
	}
	return &cycle, nil
}

func scanAdjustment(row rowScanner) (*PayrollAdjustment, error) {
	var adjustment PayrollAdjustment
	var cycleID sql.NullString
	var reason sql.NullString
	if err := row.Scan(&adjustment.ID, &adjustment.OrgID, &adjustment.EmployeeID, &cycleID, &adjustment.Type, &adjustment.Category, &adjustment.Amount, &adjustment.Currency, &adjustment.IsTaxable, &reason, &adjustment.CreatedBy, &adjustment.CreatedAt); err != nil {
		return nil, err
	}
	if cycleID.Valid {
		id, err := uuid.Parse(cycleID.String)
		if err != nil {
			return nil, err
		}
		adjustment.CycleID = &id
	}
	if reason.Valid {
		adjustment.Reason = reason.String
	}
	return &adjustment, nil
}

func scanTaxRule(row rowScanner) (*PayrollTaxRule, error) {
	var rule PayrollTaxRule
	var orgID sql.NullString
	var thresholdMin, thresholdMax sql.NullString
	var effectiveTo sql.NullTime
	if err := row.Scan(&rule.ID, &orgID, &rule.Country, &rule.Name, &rule.Rate, &rule.AppliesTo, &rule.Payer, &thresholdMin, &thresholdMax, &rule.IsActive, &rule.EffectiveFrom, &effectiveTo, &rule.CreatedAt, &rule.UpdatedAt); err != nil {
		return nil, err
	}
	if orgID.Valid {
		id, err := uuid.Parse(orgID.String)
		if err != nil {
			return nil, err
		}
		rule.OrgID = &id
	}
	if thresholdMin.Valid {
		rule.ThresholdMin = thresholdMin.String
	}
	if thresholdMax.Valid {
		rule.ThresholdMax = thresholdMax.String
	}
	if effectiveTo.Valid {
		rule.EffectiveTo = &effectiveTo.Time
	}
	return &rule, nil
}

func scanPolicy(row rowScanner) (*PayrollPolicy, error) {
	var policy PayrollPolicy
	if err := row.Scan(
		&policy.ID, &policy.OrgID, &policy.MissingAttendancePolicy,
		&policy.RegularOvertimeMultiplier, &policy.HolidayOvertimeMultiplier,
		&policy.LatePenaltyMode, &policy.LatePenaltyAmount, &policy.RoundingMode,
		&policy.VacationPayRate, &policy.SickLeavePayRate, &policy.RemotePayRate,
		&policy.BusinessTripPayRate, &policy.UnpaidLeavePayRate,
	); err != nil {
		return nil, err
	}
	return &policy, nil
}

func scanCorrection(row rowScanner) (*PayrollCorrection, error) {
	var correction PayrollCorrection
	var sourceItemID sql.NullString
	if err := row.Scan(
		&correction.ID, &correction.OrgID, &correction.EmployeeID, &sourceItemID,
		&correction.TargetCycleID, &correction.AdjustmentID, &correction.Amount,
		&correction.Type, &correction.Reason, &correction.CreatedBy, &correction.CreatedAt,
	); err != nil {
		return nil, err
	}
	if sourceItemID.Valid {
		id, err := uuid.Parse(sourceItemID.String)
		if err != nil {
			return nil, err
		}
		correction.SourceItemID = &id
	}
	return &correction, nil
}

func scanPayrollItem(row rowScanner) (*PayrollItem, error) {
	var item PayrollItem
	var reviewedBy sql.NullString
	var reviewedAt sql.NullTime
	var reviewComment sql.NullString
	if err := row.Scan(
		&item.ID, &item.CycleID, &item.OrgID, &item.EmployeeID,
		&item.BaseSalary, &item.AttendanceAdjustment, &item.OvertimeAmount,
		&item.BonusesTotal, &item.DeductionsTotal, &item.TaxesTotal,
		&item.GrossSalary, &item.NetSalary, &item.Currency, &item.WorkingDays,
		&item.PaidDays, &item.UnpaidDays, &item.LateDays, &item.AbsentDays,
		&item.MissingDays, &item.OvertimeMinutes, &item.ReviewRequired,
		&item.ReviewReasons, &item.EmployerTaxesTotal, &item.TotalEmployerCost,
		&item.Status, &reviewedBy, &reviewedAt, &reviewComment,
		&item.CalculationSnapshot,
		&item.CreatedAt, &item.UpdatedAt,
	); err != nil {
		return nil, err
	}
	if reviewedBy.Valid {
		id, err := uuid.Parse(reviewedBy.String)
		if err != nil {
			return nil, err
		}
		item.ReviewedBy = &id
	}
	if reviewedAt.Valid {
		item.ReviewedAt = &reviewedAt.Time
	}
	if reviewComment.Valid {
		item.ReviewComment = reviewComment.String
	}
	return &item, nil
}

func scanEmployeePayrollInput(row rowScanner) (*EmployeePayrollInput, error) {
	var employee EmployeePayrollInput
	var hireDate, terminationDate sql.NullTime
	if err := row.Scan(&employee.ID, &employee.OrgID, &employee.UserID, &employee.SalaryRate, &employee.Status, &hireDate, &terminationDate); err != nil {
		return nil, err
	}
	if hireDate.Valid {
		employee.HireDate = &hireDate.Time
	}
	if terminationDate.Valid {
		employee.TerminationDate = &terminationDate.Time
	}
	return &employee, nil
}

func scanSalaryHistoryItem(row rowScanner) (*SalaryHistoryItem, error) {
	var item SalaryHistoryItem
	var effectiveTo sql.NullTime
	if err := row.Scan(&item.ID, &item.EmployeeID, &item.SalaryRate, &item.EffectiveFrom, &effectiveTo); err != nil {
		return nil, err
	}
	if effectiveTo.Valid {
		item.EffectiveTo = &effectiveTo.Time
	}
	return &item, nil
}

func scanAttendanceOverride(row rowScanner) (*AttendanceOverride, error) {
	var override AttendanceOverride
	var note sql.NullString
	if err := row.Scan(&override.ID, &override.EmployeeID, &override.Date, &override.PaidDay, &override.AttendanceType, &override.AttendanceStatus, &override.OvertimeMinutes, &note); err != nil {
		return nil, err
	}
	if note.Valid {
		override.Note = note.String
	}
	return &override, nil
}

func scanAttendanceRecord(row rowScanner) (AttendanceRecord, error) {
	var record AttendanceRecord
	var checkIn, checkOut sql.NullTime
	var note sql.NullString
	err := row.Scan(&record.ID, &record.EmployeeID, &record.Date, &record.Type, &record.Source, &record.Status, &checkIn, &checkOut, &note)
	if checkIn.Valid {
		record.CheckIn = &checkIn.Time
	}
	if checkOut.Valid {
		record.CheckOut = &checkOut.Time
	}
	if note.Valid {
		record.Note = note.String
	}
	return record, err
}

func nullableString(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func nullableNumeric(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func nullableJSON(value []byte) any {
	if len(value) == 0 {
		return nil
	}
	return string(value)
}

func ptrString(id *uuid.UUID) string {
	if id == nil {
		return ""
	}
	return id.String()
}

