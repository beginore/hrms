package service

import (
	"encoding/json"
	payslipRepository "hrms/internal/feature/payslip/repository"
	"time"
)

const dateLayout = "2006-01-02"

func mapPayslip(payslip payslipRepository.Payslip) PayslipResponse {
	resp := PayslipResponse{
		ID:                 payslip.ID.String(),
		OrgID:              payslip.OrgID.String(),
		EmployeeID:         payslip.EmployeeID.String(),
		PayrollCycleID:     payslip.PayrollCycleID.String(),
		PayrollItemID:      payslip.PayrollItemID.String(),
		PeriodStart:        payslip.PeriodStart.Format(dateLayout),
		PeriodEnd:          payslip.PeriodEnd.Format(dateLayout),
		Status:             payslip.Status,
		Currency:           payslip.Currency,
		BaseSalary:         payslip.BaseSalary,
		OvertimeAmount:     payslip.OvertimeAmount,
		BonusesTotal:       payslip.BonusesTotal,
		DeductionsTotal:    payslip.DeductionsTotal,
		TaxesTotal:         payslip.TaxesTotal,
		GrossSalary:        payslip.GrossSalary,
		NetSalary:          payslip.NetSalary,
		EmployerTaxesTotal: payslip.EmployerTaxesTotal,
		TotalEmployerCost:  payslip.TotalEmployerCost,
		PDFFilename:        payslip.PDFFilename,
		PDFSHA256:          payslip.PDFSHA256,
		SentToEmail:        payslip.SentToEmail,
		GeneratedBy:        payslip.GeneratedBy.String(),
		GeneratedAt:        payslip.GeneratedAt.Format(time.RFC3339),
		VoidReason:         payslip.VoidReason,
		CreatedAt:          payslip.CreatedAt.Format(time.RFC3339),
		UpdatedAt:          payslip.UpdatedAt.Format(time.RFC3339),
	}
	if payslip.SentAt != nil {
		value := payslip.SentAt.Format(time.RFC3339)
		resp.SentAt = &value
	}
	if payslip.PDFGeneratedAt != nil {
		value := payslip.PDFGeneratedAt.Format(time.RFC3339)
		resp.PDFGeneratedAt = &value
	}
	if payslip.VoidedBy != nil {
		value := payslip.VoidedBy.String()
		resp.VoidedBy = &value
	}
	if payslip.VoidedAt != nil {
		value := payslip.VoidedAt.Format(time.RFC3339)
		resp.VoidedAt = &value
	}
	if len(payslip.PayloadSnapshot) > 0 {
		var payload any
		if err := json.Unmarshal(payslip.PayloadSnapshot, &payload); err == nil {
			resp.Payload = payload
		}
	}
	return resp
}

func mapPayslips(payslips []payslipRepository.Payslip) []PayslipResponse {
	out := make([]PayslipResponse, len(payslips))
	for i, payslip := range payslips {
		out[i] = mapPayslip(payslip)
	}
	return out
}

func mapSendResponse(payslip payslipRepository.Payslip) PayslipSendResponse {
	sentAt := ""
	if payslip.SentAt != nil {
		sentAt = payslip.SentAt.Format(time.RFC3339)
	}
	return PayslipSendResponse{
		ID:          payslip.ID.String(),
		Status:      payslip.Status,
		SentToEmail: payslip.SentToEmail,
		SentAt:      sentAt,
	}
}
