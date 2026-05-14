package service

type CreateCycleRequest struct {
	PeriodStart string `json:"period_start" binding:"required"`
	PeriodEnd   string `json:"period_end" binding:"required"`
	Currency    string `json:"currency"`
}

type CalculateCycleRequest struct {
	Recalculate bool `json:"recalculate"`
}

type CreateAdjustmentRequest struct {
	EmployeeID string `json:"employee_id" binding:"required"`
	CycleID    string `json:"cycle_id"`
	Type       string `json:"type" binding:"required"` // BONUS | DEDUCTION
	Category   string `json:"category" binding:"required"`
	Amount     string `json:"amount" binding:"required"`
	Currency   string `json:"currency"`
	IsTaxable  bool   `json:"is_taxable"`
	Reason     string `json:"reason"`
}

type CreateSalaryHistoryRequest struct {
	EmployeeID    string `json:"employee_id" binding:"required"`
	SalaryRate    string `json:"salary_rate" binding:"required"`
	EffectiveFrom string `json:"effective_from" binding:"required"`
	EffectiveTo   string `json:"effective_to"`
}

type UpsertAttendanceOverrideRequest struct {
	EmployeeID       string `json:"employee_id" binding:"required"`
	Date             string `json:"date" binding:"required"`
	PaidDay          *bool  `json:"paid_day"`
	AttendanceType   string `json:"attendance_type"`
	AttendanceStatus string `json:"attendance_status"`
	OvertimeMinutes  int32  `json:"overtime_minutes"`
	Note             string `json:"note"`
}

type UpdateEmploymentPeriodRequest struct {
	HireDate        string `json:"hire_date"`
	TerminationDate string `json:"termination_date"`
}

type ReviewPayrollItemRequest struct {
	Comment string `json:"comment"`
}

type CreateTaxRuleRequest struct {
	Country       string `json:"country"`
	Name          string `json:"name" binding:"required"`
	Rate          string `json:"rate" binding:"required"`
	AppliesTo     string `json:"applies_to" binding:"required"` // GROSS | TAXABLE_INCOME | BASE_ONLY
	Payer         string `json:"payer"`                         // EMPLOYEE | EMPLOYER
	ThresholdMin  string `json:"threshold_min"`
	ThresholdMax  string `json:"threshold_max"`
	EffectiveFrom string `json:"effective_from" binding:"required"`
	EffectiveTo   string `json:"effective_to"`
}

type UpsertPolicyRequest struct {
	MissingAttendancePolicy   string `json:"missing_attendance_policy"` // PAID_REVIEW_REQUIRED | ABSENT_UNPAID
	RegularOvertimeMultiplier string `json:"regular_overtime_multiplier"`
	HolidayOvertimeMultiplier string `json:"holiday_overtime_multiplier"`
	LatePenaltyMode           string `json:"late_penalty_mode"` // NONE | FIXED_PER_LATE_DAY
	LatePenaltyAmount         string `json:"late_penalty_amount"`
	RoundingMode              string `json:"rounding_mode"`
	VacationPayRate           string `json:"vacation_pay_rate"`
	SickLeavePayRate          string `json:"sick_leave_pay_rate"`
	RemotePayRate             string `json:"remote_pay_rate"`
	BusinessTripPayRate       string `json:"business_trip_pay_rate"`
	UnpaidLeavePayRate        string `json:"unpaid_leave_pay_rate"`
}

type CreateCorrectionRequest struct {
	EmployeeID    string `json:"employee_id" binding:"required"`
	SourceItemID  string `json:"source_item_id"`
	TargetCycleID string `json:"target_cycle_id" binding:"required"`
	Type          string `json:"type" binding:"required"` // BONUS | DEDUCTION
	Amount        string `json:"amount" binding:"required"`
	Reason        string `json:"reason" binding:"required"`
	Taxable       bool   `json:"taxable"`
}

type PreviewPayrollRequest struct {
	EmployeeID  string `json:"employee_id" binding:"required"`
	PeriodStart string `json:"period_start" binding:"required"`
	PeriodEnd   string `json:"period_end" binding:"required"`
}

type PayrollCycleResponse struct {
	ID          string  `json:"id"`
	OrgID       string  `json:"org_id"`
	PeriodStart string  `json:"period_start"`
	PeriodEnd   string  `json:"period_end"`
	Currency    string  `json:"currency"`
	Status      string  `json:"status"`
	CreatedBy   string  `json:"created_by"`
	ApprovedBy  *string `json:"approved_by,omitempty"`
	ApprovedAt  *string `json:"approved_at,omitempty"`
	PaidAt      *string `json:"paid_at,omitempty"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

type PayrollAdjustmentResponse struct {
	ID         string  `json:"id"`
	OrgID      string  `json:"org_id"`
	EmployeeID string  `json:"employee_id"`
	CycleID    *string `json:"cycle_id,omitempty"`
	Type       string  `json:"type"`
	Category   string  `json:"category"`
	Amount     string  `json:"amount"`
	Currency   string  `json:"currency"`
	IsTaxable  bool    `json:"is_taxable"`
	Reason     string  `json:"reason,omitempty"`
	CreatedBy  string  `json:"created_by"`
	CreatedAt  string  `json:"created_at"`
}

type SalaryHistoryResponse struct {
	ID            string  `json:"id"`
	EmployeeID    string  `json:"employee_id"`
	SalaryRate    string  `json:"salary_rate"`
	EffectiveFrom string  `json:"effective_from"`
	EffectiveTo   *string `json:"effective_to,omitempty"`
}

type AttendanceOverrideResponse struct {
	ID               string `json:"id"`
	EmployeeID       string `json:"employee_id"`
	Date             string `json:"date"`
	PaidDay          bool   `json:"paid_day"`
	AttendanceType   string `json:"attendance_type"`
	AttendanceStatus string `json:"attendance_status"`
	OvertimeMinutes  int32  `json:"overtime_minutes"`
	Note             string `json:"note,omitempty"`
}

type EmploymentPeriodResponse struct {
	EmployeeID      string  `json:"employee_id"`
	HireDate        *string `json:"hire_date,omitempty"`
	TerminationDate *string `json:"termination_date,omitempty"`
}

type PayrollTaxRuleResponse struct {
	ID            string  `json:"id"`
	OrgID         *string `json:"org_id,omitempty"`
	Country       string  `json:"country"`
	Name          string  `json:"name"`
	Rate          string  `json:"rate"`
	AppliesTo     string  `json:"applies_to"`
	Payer         string  `json:"payer"`
	ThresholdMin  string  `json:"threshold_min,omitempty"`
	ThresholdMax  string  `json:"threshold_max,omitempty"`
	IsActive      bool    `json:"is_active"`
	EffectiveFrom string  `json:"effective_from"`
	EffectiveTo   *string `json:"effective_to,omitempty"`
}

type PayrollPolicyResponse struct {
	OrgID                     string `json:"org_id"`
	MissingAttendancePolicy   string `json:"missing_attendance_policy"`
	RegularOvertimeMultiplier string `json:"regular_overtime_multiplier"`
	HolidayOvertimeMultiplier string `json:"holiday_overtime_multiplier"`
	LatePenaltyMode           string `json:"late_penalty_mode"`
	LatePenaltyAmount         string `json:"late_penalty_amount"`
	RoundingMode              string `json:"rounding_mode"`
	VacationPayRate           string `json:"vacation_pay_rate"`
	SickLeavePayRate          string `json:"sick_leave_pay_rate"`
	RemotePayRate             string `json:"remote_pay_rate"`
	BusinessTripPayRate       string `json:"business_trip_pay_rate"`
	UnpaidLeavePayRate        string `json:"unpaid_leave_pay_rate"`
}

type PayrollCorrectionResponse struct {
	ID            string  `json:"id"`
	OrgID         string  `json:"org_id"`
	EmployeeID    string  `json:"employee_id"`
	SourceItemID  *string `json:"source_item_id,omitempty"`
	TargetCycleID string  `json:"target_cycle_id"`
	AdjustmentID  string  `json:"adjustment_id"`
	Type          string  `json:"type"`
	Amount        string  `json:"amount"`
	Reason        string  `json:"reason"`
	CreatedBy     string  `json:"created_by"`
	CreatedAt     string  `json:"created_at"`
}

type PayrollItemResponse struct {
	ID                   string  `json:"id"`
	CycleID              string  `json:"cycle_id"`
	OrgID                string  `json:"org_id"`
	EmployeeID           string  `json:"employee_id"`
	BaseSalary           string  `json:"base_salary"`
	AttendanceAdjustment string  `json:"attendance_adjustment"`
	OvertimeAmount       string  `json:"overtime_amount"`
	BonusesTotal         string  `json:"bonuses_total"`
	DeductionsTotal      string  `json:"deductions_total"`
	TaxesTotal           string  `json:"taxes_total"`
	GrossSalary          string  `json:"gross_salary"`
	NetSalary            string  `json:"net_salary"`
	Currency             string  `json:"currency"`
	WorkingDays          int32   `json:"working_days"`
	PaidDays             string  `json:"paid_days"`
	UnpaidDays           string  `json:"unpaid_days"`
	LateDays             int32   `json:"late_days"`
	AbsentDays           int32   `json:"absent_days"`
	MissingDays          int32   `json:"missing_days"`
	OvertimeMinutes      int32   `json:"overtime_minutes"`
	ReviewRequired       bool    `json:"review_required"`
	ReviewReasons        any     `json:"review_reasons,omitempty"`
	EmployerTaxesTotal   string  `json:"employer_taxes_total"`
	TotalEmployerCost    string  `json:"total_employer_cost"`
	Status               string  `json:"status"`
	ReviewedBy           *string `json:"reviewed_by,omitempty"`
	ReviewedAt           *string `json:"reviewed_at,omitempty"`
	ReviewComment        string  `json:"review_comment,omitempty"`
	Snapshot             any     `json:"calculation_snapshot,omitempty"`
	CreatedAt            string  `json:"created_at,omitempty"`
	UpdatedAt            string  `json:"updated_at,omitempty"`
}
