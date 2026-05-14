package service

import (
	"encoding/json"
	payrollRepository "hrms/internal/feature/payroll/repository"
	"time"
)

func mapCycle(cycle payrollRepository.PayrollCycle) PayrollCycleResponse {
	resp := PayrollCycleResponse{
		ID:          cycle.ID.String(),
		OrgID:       cycle.OrgID.String(),
		PeriodStart: cycle.PeriodStart.Format(dateLayout),
		PeriodEnd:   cycle.PeriodEnd.Format(dateLayout),
		Currency:    cycle.Currency,
		Status:      cycle.Status,
		CreatedBy:   cycle.CreatedBy.String(),
		CreatedAt:   cycle.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   cycle.UpdatedAt.Format(time.RFC3339),
	}
	if cycle.ApprovedBy != nil {
		value := cycle.ApprovedBy.String()
		resp.ApprovedBy = &value
	}
	if cycle.ApprovedAt != nil {
		value := cycle.ApprovedAt.Format(time.RFC3339)
		resp.ApprovedAt = &value
	}
	if cycle.PaidAt != nil {
		value := cycle.PaidAt.Format(time.RFC3339)
		resp.PaidAt = &value
	}
	return resp
}

func mapCycles(cycles []payrollRepository.PayrollCycle) []PayrollCycleResponse {
	out := make([]PayrollCycleResponse, len(cycles))
	for i, cycle := range cycles {
		out[i] = mapCycle(cycle)
	}
	return out
}

func mapAdjustment(adjustment payrollRepository.PayrollAdjustment) PayrollAdjustmentResponse {
	resp := PayrollAdjustmentResponse{
		ID:         adjustment.ID.String(),
		OrgID:      adjustment.OrgID.String(),
		EmployeeID: adjustment.EmployeeID.String(),
		Type:       adjustment.Type,
		Category:   adjustment.Category,
		Amount:     adjustment.Amount,
		Currency:   adjustment.Currency,
		IsTaxable:  adjustment.IsTaxable,
		Reason:     adjustment.Reason,
		CreatedBy:  adjustment.CreatedBy.String(),
		CreatedAt:  adjustment.CreatedAt.Format(time.RFC3339),
	}
	if adjustment.CycleID != nil {
		value := adjustment.CycleID.String()
		resp.CycleID = &value
	}
	return resp
}

func mapSalaryHistory(item payrollRepository.SalaryHistoryItem) SalaryHistoryResponse {
	resp := SalaryHistoryResponse{
		ID:            item.ID.String(),
		EmployeeID:    item.EmployeeID.String(),
		SalaryRate:    item.SalaryRate,
		EffectiveFrom: item.EffectiveFrom.Format(dateLayout),
	}
	if item.EffectiveTo != nil {
		value := item.EffectiveTo.Format(dateLayout)
		resp.EffectiveTo = &value
	}
	return resp
}

func mapAttendanceOverride(override payrollRepository.AttendanceOverride) AttendanceOverrideResponse {
	return AttendanceOverrideResponse{
		ID:               override.ID.String(),
		EmployeeID:       override.EmployeeID.String(),
		Date:             override.Date.Format(dateLayout),
		PaidDay:          override.PaidDay,
		AttendanceType:   override.AttendanceType,
		AttendanceStatus: override.AttendanceStatus,
		OvertimeMinutes:  override.OvertimeMinutes,
		Note:             override.Note,
	}
}

func mapEmploymentPeriod(employee payrollRepository.EmployeePayrollInput) EmploymentPeriodResponse {
	resp := EmploymentPeriodResponse{
		EmployeeID: employee.ID.String(),
	}
	if employee.HireDate != nil {
		value := employee.HireDate.Format(dateLayout)
		resp.HireDate = &value
	}
	if employee.TerminationDate != nil {
		value := employee.TerminationDate.Format(dateLayout)
		resp.TerminationDate = &value
	}
	return resp
}

func mapAdjustments(adjustments []payrollRepository.PayrollAdjustment) []PayrollAdjustmentResponse {
	out := make([]PayrollAdjustmentResponse, len(adjustments))
	for i, adjustment := range adjustments {
		out[i] = mapAdjustment(adjustment)
	}
	return out
}

func mapTaxRule(rule payrollRepository.PayrollTaxRule) PayrollTaxRuleResponse {
	resp := PayrollTaxRuleResponse{
		ID:            rule.ID.String(),
		Country:       rule.Country,
		Name:          rule.Name,
		Rate:          rule.Rate,
		AppliesTo:     rule.AppliesTo,
		Payer:         rule.Payer,
		ThresholdMin:  rule.ThresholdMin,
		ThresholdMax:  rule.ThresholdMax,
		IsActive:      rule.IsActive,
		EffectiveFrom: rule.EffectiveFrom.Format(dateLayout),
	}
	if rule.OrgID != nil {
		value := rule.OrgID.String()
		resp.OrgID = &value
	}
	if rule.EffectiveTo != nil {
		value := rule.EffectiveTo.Format(dateLayout)
		resp.EffectiveTo = &value
	}
	return resp
}

func mapPolicy(policy payrollRepository.PayrollPolicy) PayrollPolicyResponse {
	return PayrollPolicyResponse{
		OrgID:                     policy.OrgID.String(),
		MissingAttendancePolicy:   policy.MissingAttendancePolicy,
		RegularOvertimeMultiplier: policy.RegularOvertimeMultiplier,
		HolidayOvertimeMultiplier: policy.HolidayOvertimeMultiplier,
		LatePenaltyMode:           policy.LatePenaltyMode,
		LatePenaltyAmount:         policy.LatePenaltyAmount,
		RoundingMode:              policy.RoundingMode,
		VacationPayRate:           policy.VacationPayRate,
		SickLeavePayRate:          policy.SickLeavePayRate,
		RemotePayRate:             policy.RemotePayRate,
		BusinessTripPayRate:       policy.BusinessTripPayRate,
		UnpaidLeavePayRate:        policy.UnpaidLeavePayRate,
	}
}

func mapCorrection(correction payrollRepository.PayrollCorrection) PayrollCorrectionResponse {
	resp := PayrollCorrectionResponse{
		ID:            correction.ID.String(),
		OrgID:         correction.OrgID.String(),
		EmployeeID:    correction.EmployeeID.String(),
		TargetCycleID: correction.TargetCycleID.String(),
		AdjustmentID:  correction.AdjustmentID.String(),
		Type:          correction.Type,
		Amount:        correction.Amount,
		Reason:        correction.Reason,
		CreatedBy:     correction.CreatedBy.String(),
		CreatedAt:     correction.CreatedAt.Format(time.RFC3339),
	}
	if correction.SourceItemID != nil {
		value := correction.SourceItemID.String()
		resp.SourceItemID = &value
	}
	return resp
}

func mapTaxRules(rules []payrollRepository.PayrollTaxRule) []PayrollTaxRuleResponse {
	out := make([]PayrollTaxRuleResponse, len(rules))
	for i, rule := range rules {
		out[i] = mapTaxRule(rule)
	}
	return out
}

func mapPayrollItem(item payrollRepository.PayrollItem) PayrollItemResponse {
	resp := PayrollItemResponse{
		ID:                   item.ID.String(),
		CycleID:              item.CycleID.String(),
		OrgID:                item.OrgID.String(),
		EmployeeID:           item.EmployeeID.String(),
		BaseSalary:           item.BaseSalary,
		AttendanceAdjustment: item.AttendanceAdjustment,
		OvertimeAmount:       item.OvertimeAmount,
		BonusesTotal:         item.BonusesTotal,
		DeductionsTotal:      item.DeductionsTotal,
		TaxesTotal:           item.TaxesTotal,
		GrossSalary:          item.GrossSalary,
		NetSalary:            item.NetSalary,
		Currency:             item.Currency,
		WorkingDays:          item.WorkingDays,
		PaidDays:             item.PaidDays,
		UnpaidDays:           item.UnpaidDays,
		LateDays:             item.LateDays,
		AbsentDays:           item.AbsentDays,
		MissingDays:          item.MissingDays,
		OvertimeMinutes:      item.OvertimeMinutes,
		ReviewRequired:       item.ReviewRequired,
		EmployerTaxesTotal:   item.EmployerTaxesTotal,
		TotalEmployerCost:    item.TotalEmployerCost,
		Status:               item.Status,
		CreatedAt:            item.CreatedAt.Format(time.RFC3339),
		UpdatedAt:            item.UpdatedAt.Format(time.RFC3339),
	}
	if len(item.CalculationSnapshot) > 0 {
		var snapshot any
		if err := json.Unmarshal(item.CalculationSnapshot, &snapshot); err == nil {
			resp.Snapshot = snapshot
		}
	}
	if len(item.ReviewReasons) > 0 {
		var reasons any
		if err := json.Unmarshal(item.ReviewReasons, &reasons); err == nil {
			resp.ReviewReasons = reasons
		}
	}
	if item.ReviewedBy != nil {
		value := item.ReviewedBy.String()
		resp.ReviewedBy = &value
	}
	if item.ReviewedAt != nil {
		value := item.ReviewedAt.Format(time.RFC3339)
		resp.ReviewedAt = &value
	}
	resp.ReviewComment = item.ReviewComment
	return resp
}

func mapPayrollItems(items []payrollRepository.PayrollItem) []PayrollItemResponse {
	out := make([]PayrollItemResponse, len(items))
	for i, item := range items {
		out[i] = mapPayrollItem(item)
	}
	return out
}
