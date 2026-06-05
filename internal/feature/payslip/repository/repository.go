package repository

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
)

type payslipRepository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) PayslipRepository {
	return &payslipRepository{db: db}
}

func (r *payslipRepository) GetActorAccess(ctx context.Context, userID uuid.UUID) (*ActorAccess, error) {
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

func (r *payslipRepository) GetEmployeeIDByUserID(ctx context.Context, userID uuid.UUID) (uuid.UUID, error) {
	var employeeID uuid.UUID
	err := r.db.QueryRowContext(ctx, `SELECT id FROM employees WHERE user_id = $1`, userID).Scan(&employeeID)
	return employeeID, err
}

func (r *payslipRepository) GetPayrollItem(ctx context.Context, id uuid.UUID) (*PayrollItem, error) {
	row := r.db.QueryRowContext(ctx, payrollItemSelect()+` WHERE id = $1`, id)
	return scanPayrollItem(row)
}

func (r *payslipRepository) ListPayrollItemsByCycle(ctx context.Context, cycleID uuid.UUID) ([]PayrollItem, error) {
	rows, err := r.db.QueryContext(ctx, payrollItemSelect()+` WHERE cycle_id = $1 ORDER BY employee_id`, cycleID)
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

func (r *payslipRepository) GetPayrollCycle(ctx context.Context, id uuid.UUID) (*PayrollCycle, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, org_id, period_start, period_end, status, currency
		FROM payroll_cycles
		WHERE id = $1
	`, id)
	var cycle PayrollCycle
	if err := row.Scan(&cycle.ID, &cycle.OrgID, &cycle.PeriodStart, &cycle.PeriodEnd, &cycle.Status, &cycle.Currency); err != nil {
		return nil, err
	}
	return &cycle, nil
}

func (r *payslipRepository) GetEmployeeProfile(ctx context.Context, employeeID uuid.UUID) (*EmployeeProfile, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT e.id, e.org_id, e.user_id, u.first_name, u.last_name, u.email, e.role, d.name, p.name
		FROM employees e
		JOIN users u ON u.id = e.user_id
		JOIN departments d ON d.id = e.department_id
		JOIN positions p ON p.id = e.position_id
		WHERE e.id = $1
	`, employeeID)
	var employee EmployeeProfile
	if err := row.Scan(
		&employee.ID, &employee.OrgID, &employee.UserID, &employee.FirstName,
		&employee.LastName, &employee.Email, &employee.Role,
		&employee.DepartmentName, &employee.PositionName,
	); err != nil {
		return nil, err
	}
	return &employee, nil
}

func (r *payslipRepository) GetOrganizationProfile(ctx context.Context, orgID uuid.UUID) (*OrganizationProfile, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, name, vat_id, address
		FROM organizations
		WHERE id = $1
	`, orgID)
	var org OrganizationProfile
	if err := row.Scan(&org.ID, &org.Name, &org.VATID, &org.Address); err != nil {
		return nil, err
	}
	return &org, nil
}

func (r *payslipRepository) ListPayrollAdjustments(ctx context.Context, cycleID, employeeID uuid.UUID) ([]PayrollAdjustment, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, type, category, amount::text, currency, is_taxable, reason
		FROM payroll_adjustments
		WHERE cycle_id = $1 AND employee_id = $2
		ORDER BY created_at
	`, cycleID, employeeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var adjustments []PayrollAdjustment
	for rows.Next() {
		var adjustment PayrollAdjustment
		var reason sql.NullString
		if err := rows.Scan(
			&adjustment.ID, &adjustment.Type, &adjustment.Category, &adjustment.Amount,
			&adjustment.Currency, &adjustment.IsTaxable, &reason,
		); err != nil {
			return nil, err
		}
		if reason.Valid {
			adjustment.Reason = reason.String
		}
		adjustments = append(adjustments, adjustment)
	}
	return adjustments, rows.Err()
}

func (r *payslipRepository) CreatePayslip(ctx context.Context, params CreatePayslipParams) (*Payslip, error) {
	row := r.db.QueryRowContext(ctx, `
		INSERT INTO payslips (
			id, org_id, employee_id, payroll_cycle_id, payroll_item_id,
			period_start, period_end, status, currency, base_salary, overtime_amount,
			bonuses_total, deductions_total, taxes_total, gross_salary, net_salary,
			employer_taxes_total, total_employer_cost, payload_snapshot, pdf_content,
			pdf_filename, pdf_generated_at, pdf_sha256, generated_by
		)
		VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9,
			$10::numeric, $11::numeric, $12::numeric, $13::numeric, $14::numeric,
			$15::numeric, $16::numeric, $17::numeric, $18::numeric, $19::jsonb,
			$20, $21, $22, $23, $24
		)
		RETURNING `+payslipColumns()+`
	`, params.ID, params.OrgID, params.EmployeeID, params.PayrollCycleID, params.PayrollItemID,
		params.PeriodStart, params.PeriodEnd, params.Status, params.Currency, params.BaseSalary,
		params.OvertimeAmount, params.BonusesTotal, params.DeductionsTotal, params.TaxesTotal,
		params.GrossSalary, params.NetSalary, params.EmployerTaxesTotal, params.TotalEmployerCost,
		string(params.PayloadSnapshot), params.PDFContent, nullableString(params.PDFFilename),
		params.PDFGeneratedAt, nullableString(params.PDFSHA256), params.GeneratedBy)
	return scanPayslip(row)
}

func (r *payslipRepository) GetPayslip(ctx context.Context, id uuid.UUID) (*Payslip, error) {
	row := r.db.QueryRowContext(ctx, `SELECT `+payslipColumns()+` FROM payslips WHERE id = $1`, id)
	return scanPayslip(row)
}

func (r *payslipRepository) GetPayslipByPayrollItem(ctx context.Context, payrollItemID uuid.UUID) (*Payslip, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT `+payslipColumns()+`
		FROM payslips
		WHERE payroll_item_id = $1 AND status <> 'VOID'
		ORDER BY generated_at DESC
		LIMIT 1
	`, payrollItemID)
	return scanPayslip(row)
}

func (r *payslipRepository) ListPayslips(ctx context.Context, filter ListPayslipsFilter) ([]Payslip, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT `+payslipColumns()+`
		FROM payslips
		WHERE org_id = $1
		  AND ($2::uuid IS NULL OR employee_id = $2)
		  AND ($3::uuid IS NULL OR payroll_cycle_id = $3)
		  AND ($4::varchar = '' OR status = $4)
		  AND ($5::date IS NULL OR period_start >= $5)
		  AND ($6::date IS NULL OR period_end <= $6)
		ORDER BY period_start DESC, generated_at DESC
	`, filter.OrgID, filter.EmployeeID, filter.CycleID, filter.Status, filter.PeriodStart, filter.PeriodEnd)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payslips []Payslip
	for rows.Next() {
		payslip, err := scanPayslip(rows)
		if err != nil {
			return nil, err
		}
		payslips = append(payslips, *payslip)
	}
	return payslips, rows.Err()
}

func (r *payslipRepository) StorePayslipPDF(ctx context.Context, id uuid.UUID, content []byte, filename, sha256 string) (*Payslip, error) {
	row := r.db.QueryRowContext(ctx, `
		UPDATE payslips
		SET pdf_content = $2,
			pdf_filename = $3,
			pdf_generated_at = now(),
			pdf_sha256 = $4,
			updated_at = now()
		WHERE id = $1 AND status <> 'VOID'
		RETURNING `+payslipColumns()+`
	`, id, content, nullableString(filename), nullableString(sha256))
	return scanPayslip(row)
}

func (r *payslipRepository) MarkPayslipSent(ctx context.Context, id uuid.UUID, sentToEmail string) (*Payslip, error) {
	row := r.db.QueryRowContext(ctx, `
		UPDATE payslips
		SET status = 'SENT',
			sent_to_email = $2,
			sent_at = now(),
			updated_at = now()
		WHERE id = $1 AND status <> 'VOID'
		RETURNING `+payslipColumns()+`
	`, id, sentToEmail)
	return scanPayslip(row)
}

func (r *payslipRepository) VoidPayslip(ctx context.Context, id uuid.UUID, actorUserID uuid.UUID, reason string) (*Payslip, error) {
	row := r.db.QueryRowContext(ctx, `
		UPDATE payslips
		SET status = 'VOID',
			voided_by = $2,
			voided_at = now(),
			void_reason = $3,
			updated_at = now()
		WHERE id = $1 AND status <> 'VOID'
		RETURNING `+payslipColumns()+`
	`, id, actorUserID, nullableString(reason))
	return scanPayslip(row)
}

func payrollItemSelect() string {
	return `
		SELECT id, cycle_id, org_id, employee_id, base_salary::text, attendance_adjustment::text,
			overtime_amount::text, bonuses_total::text, deductions_total::text, taxes_total::text,
			gross_salary::text, net_salary::text, currency, working_days, paid_days::text, unpaid_days::text,
			late_days, absent_days, missing_days, overtime_minutes, review_required, review_reasons,
			employer_taxes_total::text, total_employer_cost::text, status, calculation_snapshot
		FROM payroll_items`
}

func payslipColumns() string {
	return `id, org_id, employee_id, payroll_cycle_id, payroll_item_id, period_start, period_end,
		status, currency, base_salary::text, overtime_amount::text, bonuses_total::text,
		deductions_total::text, taxes_total::text, gross_salary::text, net_salary::text,
		employer_taxes_total::text, total_employer_cost::text, payload_snapshot,
		pdf_content, pdf_filename, pdf_generated_at, pdf_sha256, sent_to_email,
		sent_at, generated_by, generated_at, voided_by, voided_at,
		void_reason, created_at, updated_at`
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanPayrollItem(row rowScanner) (*PayrollItem, error) {
	var item PayrollItem
	if err := row.Scan(
		&item.ID, &item.CycleID, &item.OrgID, &item.EmployeeID,
		&item.BaseSalary, &item.AttendanceAdjustment, &item.OvertimeAmount,
		&item.BonusesTotal, &item.DeductionsTotal, &item.TaxesTotal,
		&item.GrossSalary, &item.NetSalary, &item.Currency, &item.WorkingDays,
		&item.PaidDays, &item.UnpaidDays, &item.LateDays, &item.AbsentDays,
		&item.MissingDays, &item.OvertimeMinutes, &item.ReviewRequired,
		&item.ReviewReasons, &item.EmployerTaxesTotal, &item.TotalEmployerCost,
		&item.Status, &item.CalculationSnapshot,
	); err != nil {
		return nil, err
	}
	return &item, nil
}

func scanPayslip(row rowScanner) (*Payslip, error) {
	var payslip Payslip
	var pdfFilename, pdfSHA256, sentToEmail, voidReason sql.NullString
	var pdfGeneratedAt, sentAt, voidedAt sql.NullTime
	var voidedBy sql.NullString
	if err := row.Scan(
		&payslip.ID, &payslip.OrgID, &payslip.EmployeeID, &payslip.PayrollCycleID,
		&payslip.PayrollItemID, &payslip.PeriodStart, &payslip.PeriodEnd,
		&payslip.Status, &payslip.Currency, &payslip.BaseSalary, &payslip.OvertimeAmount,
		&payslip.BonusesTotal, &payslip.DeductionsTotal, &payslip.TaxesTotal,
		&payslip.GrossSalary, &payslip.NetSalary, &payslip.EmployerTaxesTotal,
		&payslip.TotalEmployerCost, &payslip.PayloadSnapshot, &payslip.PDFContent,
		&pdfFilename, &pdfGeneratedAt, &pdfSHA256, &sentToEmail, &sentAt,
		&payslip.GeneratedBy, &payslip.GeneratedAt, &voidedBy, &voidedAt, &voidReason,
		&payslip.CreatedAt, &payslip.UpdatedAt,
	); err != nil {
		return nil, err
	}
	if pdfFilename.Valid {
		payslip.PDFFilename = pdfFilename.String
	}
	if pdfGeneratedAt.Valid {
		payslip.PDFGeneratedAt = &pdfGeneratedAt.Time
	}
	if pdfSHA256.Valid {
		payslip.PDFSHA256 = pdfSHA256.String
	}
	if sentToEmail.Valid {
		payslip.SentToEmail = sentToEmail.String
	}
	if sentAt.Valid {
		payslip.SentAt = &sentAt.Time
	}
	if voidedBy.Valid {
		id, err := uuid.Parse(voidedBy.String)
		if err != nil {
			return nil, err
		}
		payslip.VoidedBy = &id
	}
	if voidedAt.Valid {
		payslip.VoidedAt = &voidedAt.Time
	}
	if voidReason.Valid {
		payslip.VoidReason = voidReason.String
	}
	return &payslip, nil
}

func nullableString(value string) any {
	if value == "" {
		return nil
	}
	return value
}
