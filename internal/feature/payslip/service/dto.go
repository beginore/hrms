package service

type GeneratePayslipRequest struct {
	PayrollItemID string `json:"payroll_item_id" binding:"required"`
}

type VoidPayslipRequest struct {
	Reason string `json:"reason"`
}

type PayslipResponse struct {
	ID                 string  `json:"id"`
	OrgID              string  `json:"org_id"`
	EmployeeID         string  `json:"employee_id"`
	PayrollCycleID     string  `json:"payroll_cycle_id"`
	PayrollItemID      string  `json:"payroll_item_id"`
	PeriodStart        string  `json:"period_start"`
	PeriodEnd          string  `json:"period_end"`
	Status             string  `json:"status"`
	Currency           string  `json:"currency"`
	BaseSalary         string  `json:"base_salary"`
	OvertimeAmount     string  `json:"overtime_amount"`
	BonusesTotal       string  `json:"bonuses_total"`
	DeductionsTotal    string  `json:"deductions_total"`
	TaxesTotal         string  `json:"taxes_total"`
	GrossSalary        string  `json:"gross_salary"`
	NetSalary          string  `json:"net_salary"`
	EmployerTaxesTotal string  `json:"employer_taxes_total"`
	TotalEmployerCost  string  `json:"total_employer_cost"`
	PDFFilename        string  `json:"pdf_filename,omitempty"`
	PDFGeneratedAt     *string `json:"pdf_generated_at,omitempty"`
	PDFSHA256          string  `json:"pdf_sha256,omitempty"`
	SentToEmail        string  `json:"sent_to_email,omitempty"`
	SentAt             *string `json:"sent_at,omitempty"`
	GeneratedBy        string  `json:"generated_by"`
	GeneratedAt        string  `json:"generated_at"`
	VoidedBy           *string `json:"voided_by,omitempty"`
	VoidedAt           *string `json:"voided_at,omitempty"`
	VoidReason         string  `json:"void_reason,omitempty"`
	Payload            any     `json:"payload_snapshot,omitempty"`
	CreatedAt          string  `json:"created_at"`
	UpdatedAt          string  `json:"updated_at"`
}

type PayslipBatchResponse struct {
	Created []PayslipResponse `json:"created"`
	Skipped []PayslipResponse `json:"skipped"`
}

type PayslipSendResponse struct {
	ID          string `json:"id"`
	Status      string `json:"status"`
	SentToEmail string `json:"sent_to_email"`
	SentAt      string `json:"sent_at"`
}

type PayslipBatchSendResponse struct {
	Sent   []PayslipSendResponse `json:"sent"`
	Failed []PayslipSendFailure  `json:"failed"`
}

type PayslipSendFailure struct {
	PayslipID string `json:"payslip_id"`
	Error     string `json:"error"`
}
