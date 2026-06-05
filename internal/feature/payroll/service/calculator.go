package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	payrollRepository "hrms/internal/feature/payroll/repository"
	"time"

	"github.com/google/uuid"
)

const (
	dateLayout = "2006-01-02"

	statusDraft      = "DRAFT"
	statusCalculated = "CALCULATED"
	statusApproved   = "APPROVED"
	statusReady      = "READY_FOR_PAYMENT"
	statusPaid       = "PAID"

	typeBonus     = "BONUS"
	typeDeduction = "DEDUCTION"
)

type calculationResult struct {
	Item     payrollRepository.UpsertPayrollItemParams
	Response PayrollItemResponse
}

type calculationSnapshot struct {
	EmployeeID        string                 `json:"employee_id"`
	PeriodStart       string                 `json:"period_start"`
	PeriodEnd         string                 `json:"period_end"`
	MonthlySalary     string                 `json:"monthly_salary"`
	DailyRate         string                 `json:"daily_rate"`
	WorkingDays       int32                  `json:"working_days"`
	PaidDates         []string               `json:"paid_dates"`
	UnpaidDates       []string               `json:"unpaid_dates"`
	MissingDates      []string               `json:"missing_dates"`
	LateDates         []string               `json:"late_dates"`
	OvertimeByDate    map[string]int32       `json:"overtime_by_date"`
	AttendanceSources []string               `json:"attendance_sources"`
	Adjustments       []snapshotAdjustment   `json:"adjustments"`
	TaxRules          []snapshotTaxRule      `json:"tax_rules"`
	ReviewReasons     []string               `json:"review_reasons"`
	Policy            map[string]interface{} `json:"policy"`
}

type snapshotAdjustment struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	Category  string `json:"category"`
	Amount    string `json:"amount"`
	IsTaxable bool   `json:"is_taxable"`
}

type snapshotTaxRule struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Rate      string `json:"rate"`
	AppliesTo string `json:"applies_to"`
	Payer     string `json:"payer"`
	TaxAmount string `json:"tax_amount"`
}

func (s *payrollService) calculateEmployee(ctx context.Context, cycle payrollRepository.PayrollCycle, employee payrollRepository.EmployeePayrollInput) (*calculationResult, error) {
	monthlySalary, err := ParseMoney(employee.SalaryRate)
	if err != nil {
		return nil, ErrInvalidSalaryRate
	}
	salaryHistory, err := s.repo.ListSalaryHistory(ctx, employee.ID, cycle.PeriodStart, cycle.PeriodEnd)
	if err != nil {
		return nil, fmt.Errorf("%w: salary history: %v", ErrDatabaseQueryFailed, err)
	}

	policy, err := s.getPolicy(ctx, cycle.OrgID)
	if err != nil {
		return nil, err
	}

	calendarDays, err := s.repo.ListCalendarDays(ctx, cycle.OrgID, cycle.PeriodStart, cycle.PeriodEnd)
	if err != nil {
		return nil, fmt.Errorf("%w: calendar: %v", ErrDatabaseQueryFailed, err)
	}
	workingDates := workingDateList(cycle.PeriodStart, cycle.PeriodEnd, calendarDays)
	workingDates = filterEmploymentDates(workingDates, employee.HireDate, employee.TerminationDate)
	if len(workingDates) == 0 {
		return nil, ErrNoWorkingDays
	}
	workingDateSet := make(map[string]time.Time, len(workingDates))
	for _, date := range workingDates {
		workingDateSet[date.Format(dateLayout)] = date
	}

	attendance, err := s.repo.ListAttendance(ctx, employee.ID, cycle.PeriodStart, cycle.PeriodEnd)
	if err != nil {
		return nil, fmt.Errorf("%w: attendance: %v", ErrDatabaseQueryFailed, err)
	}
	attendanceByDate := make(map[string]payrollRepository.AttendanceRecord, len(attendance))
	for _, record := range attendance {
		attendanceByDate[record.Date.Format(dateLayout)] = record
	}

	leaves, err := s.repo.ListApprovedLeaves(ctx, employee.ID, cycle.PeriodStart, cycle.PeriodEnd)
	if err != nil {
		return nil, fmt.Errorf("%w: leaves: %v", ErrDatabaseQueryFailed, err)
	}
	leaveByDate := approvedLeavesByDate(leaves, cycle.PeriodStart, cycle.PeriodEnd)
	overrides, err := s.repo.ListAttendanceOverrides(ctx, employee.ID, cycle.PeriodStart, cycle.PeriodEnd)
	if err != nil {
		return nil, fmt.Errorf("%w: attendance overrides: %v", ErrDatabaseQueryFailed, err)
	}
	overrideByDate := attendanceOverridesByDate(overrides)

	schedule := defaultWorkSchedule(employee.ID)
	if configured, err := s.repo.GetWorkSchedule(ctx, employee.ID); err == nil {
		schedule = *configured
	} else if !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("%w: work schedule: %v", ErrDatabaseQueryFailed, err)
	}

	workingDaysCount := int32(len(workingDates))
	dailyRate := monthlySalary.DivInt(int64(workingDaysCount))
	baseSalary := ZeroMoney()
	attendanceAdjustment := ZeroMoney()
	overtimeAmount := ZeroMoney()
	latePenaltyAmount := ZeroMoney()
	var paidDays, unpaidDays, lateDays, absentDays, overtimeMinutes int32
	var paidDates, unpaidDates, missingDates, lateDates, attendanceSources []string
	reviewReasons := []string{}
	overtimeByDate := map[string]int32{}

	for _, date := range workingDates {
		dateKey := date.Format(dateLayout)
		record, hasRecord := attendanceByDate[dateKey]
		leave, hasLeave := leaveByDate[dateKey]
		override, hasOverride := overrideByDate[dateKey]
		daySalary := salaryForDate(date, monthlySalary, salaryHistory)
		currentDailyRate := daySalary.DivInt(int64(workingDaysCount))

		isPaid := false
		payRate := int64(rateScale)
		if hasOverride {
			isPaid = override.PaidDay
			if !override.PaidDay {
				reviewReasons = append(reviewReasons, "manual unpaid override on "+dateKey)
			}
		} else if hasLeave {
			payRate = leavePayRate(policy, leave.Type)
			isPaid = payRate > 0
		} else if hasRecord {
			switch record.Status {
			case "PRESENT", "LATE", "ON_LEAVE":
				isPaid = true
			}
		} else {
			missingDates = append(missingDates, dateKey)
			reviewReasons = append(reviewReasons, "missing attendance on "+dateKey)
			isPaid = policy.MissingAttendancePolicy != "ABSENT_UNPAID"
		}

		if isPaid {
			paidDays++
			paidDates = append(paidDates, dateKey)
			baseSalary = baseSalary.Add(currentDailyRate.MulRate(payRate))
		} else {
			unpaidDays++
			absentDays++
			unpaidDates = append(unpaidDates, dateKey)
			attendanceAdjustment = attendanceAdjustment.Sub(currentDailyRate)
		}

		if hasOverride {
			attendanceSources = append(attendanceSources, "override:"+override.ID.String())
			if override.AttendanceStatus == "LATE" {
				lateDays++
				lateDates = append(lateDates, dateKey)
			}
			if override.OvertimeMinutes > 0 {
				multiplier, err := ParseRateMicros(policy.RegularOvertimeMultiplier)
				if err != nil {
					return nil, ErrInvalidRate
				}
				overtimeMinutes += override.OvertimeMinutes
				overtimeByDate[dateKey] = override.OvertimeMinutes
				overtimeAmount = overtimeAmount.Add(overtimePay(daySalary, workingDaysCount, schedule, override.OvertimeMinutes, multiplier))
			}
		} else if hasRecord {
			attendanceSources = append(attendanceSources, record.ID.String())
			if record.Status == "LATE" {
				lateDays++
				lateDates = append(lateDates, dateKey)
				if policy.LatePenaltyMode == "FIXED_PER_LATE_DAY" {
					penalty, err := ParseMoney(policy.LatePenaltyAmount)
					if err != nil {
						return nil, ErrInvalidAmount
					}
					latePenaltyAmount = latePenaltyAmount.Add(penalty)
				}
			}
			minutes := calculateOvertimeMinutes(record, schedule, false)
			if minutes > 0 {
				multiplier, err := ParseRateMicros(policy.RegularOvertimeMultiplier)
				if err != nil {
					return nil, ErrInvalidRate
				}
				overtimeMinutes += minutes
				overtimeByDate[dateKey] = minutes
				overtimeAmount = overtimeAmount.Add(overtimePay(daySalary, workingDaysCount, schedule, minutes, multiplier))
			}
		}
	}

	for _, record := range attendance {
		dateKey := record.Date.Format(dateLayout)
		if _, isWorking := workingDateSet[dateKey]; isWorking {
			continue
		}
		minutes := calculateOvertimeMinutes(record, schedule, true)
		if minutes > 0 {
			daySalary := salaryForDate(record.Date, monthlySalary, salaryHistory)
			multiplier, err := ParseRateMicros(policy.HolidayOvertimeMultiplier)
			if err != nil {
				return nil, ErrInvalidRate
			}
			overtimeMinutes += minutes
			overtimeByDate[dateKey] = minutes
			overtimeAmount = overtimeAmount.Add(overtimePay(daySalary, workingDaysCount, schedule, minutes, multiplier))
		}
	}

	adjustments, err := s.repo.ListAdjustments(ctx, cycle.OrgID, &cycle.ID, &employee.ID)
	if err != nil {
		return nil, fmt.Errorf("%w: adjustments: %v", ErrDatabaseQueryFailed, err)
	}

	bonusesTotal := ZeroMoney()
	taxableBonuses := ZeroMoney()
	deductionsTotal := ZeroMoney()
	snapshotAdjustments := make([]snapshotAdjustment, 0, len(adjustments))
	for _, adjustment := range adjustments {
		amount, err := ParseMoney(adjustment.Amount)
		if err != nil {
			return nil, ErrInvalidAmount
		}
		switch adjustment.Type {
		case typeBonus:
			bonusesTotal = bonusesTotal.Add(amount)
			if adjustment.IsTaxable {
				taxableBonuses = taxableBonuses.Add(amount)
			}
		case typeDeduction:
			deductionsTotal = deductionsTotal.Add(amount)
		}
		snapshotAdjustments = append(snapshotAdjustments, snapshotAdjustment{
			ID:        adjustment.ID.String(),
			Type:      adjustment.Type,
			Category:  adjustment.Category,
			Amount:    amount.String(),
			IsTaxable: adjustment.IsTaxable,
		})
	}
	deductionsTotal = deductionsTotal.Add(latePenaltyAmount)

	grossSalary := baseSalary.Add(overtimeAmount).Add(bonusesTotal)
	taxableIncome := baseSalary.Add(overtimeAmount).Add(taxableBonuses)
	taxesTotal, employerTaxesTotal, snapshotTaxRules, err := s.calculateTaxes(ctx, cycle.OrgID, cycle.PeriodEnd, taxableIncome, baseSalary, grossSalary)
	if err != nil {
		return nil, err
	}

	netSalary := grossSalary.Sub(deductionsTotal).Sub(taxesTotal)
	if netSalary.IsNegative() {
		netSalary = ZeroMoney()
		reviewReasons = append(reviewReasons, "net salary was below zero and clamped to zero")
	}
	totalEmployerCost := grossSalary.Add(employerTaxesTotal)
	baseSalary = roundMoney(baseSalary, policy.RoundingMode)
	overtimeAmount = roundMoney(overtimeAmount, policy.RoundingMode)
	bonusesTotal = roundMoney(bonusesTotal, policy.RoundingMode)
	deductionsTotal = roundMoney(deductionsTotal, policy.RoundingMode)
	taxesTotal = roundMoney(taxesTotal, policy.RoundingMode)
	employerTaxesTotal = roundMoney(employerTaxesTotal, policy.RoundingMode)
	grossSalary = roundMoney(grossSalary, policy.RoundingMode)
	netSalary = grossSalary.Sub(deductionsTotal).Sub(taxesTotal)
	if netSalary.IsNegative() {
		netSalary = ZeroMoney()
		reviewReasons = append(reviewReasons, "net salary was below zero after rounding and clamped to zero")
	}
	netSalary = roundMoney(netSalary, policy.RoundingMode)
	totalEmployerCost = roundMoney(grossSalary.Add(employerTaxesTotal), policy.RoundingMode)
	reviewRequired := len(reviewReasons) > 0
	reviewReasonBytes, err := json.Marshal(reviewReasons)
	if err != nil {
		return nil, fmt.Errorf("%w: review reasons: %v", ErrDatabaseSaveFailed, err)
	}

	snapshot := calculationSnapshot{
		EmployeeID:        employee.ID.String(),
		PeriodStart:       cycle.PeriodStart.Format(dateLayout),
		PeriodEnd:         cycle.PeriodEnd.Format(dateLayout),
		MonthlySalary:     monthlySalary.String(),
		DailyRate:         dailyRate.String(),
		WorkingDays:       workingDaysCount,
		PaidDates:         paidDates,
		UnpaidDates:       unpaidDates,
		MissingDates:      missingDates,
		LateDates:         lateDates,
		OvertimeByDate:    overtimeByDate,
		AttendanceSources: attendanceSources,
		Adjustments:       snapshotAdjustments,
		TaxRules:          snapshotTaxRules,
		ReviewReasons:     reviewReasons,
		Policy: map[string]interface{}{
			"missing_workday_policy":      policy.MissingAttendancePolicy,
			"regular_overtime_multiplier": policy.RegularOvertimeMultiplier,
			"holiday_overtime_multiplier": policy.HolidayOvertimeMultiplier,
			"late_penalty_mode":           policy.LatePenaltyMode,
			"late_penalty_amount":         policy.LatePenaltyAmount,
			"rounding_mode":               policy.RoundingMode,
			"vacation_pay_rate":           policy.VacationPayRate,
			"sick_leave_pay_rate":         policy.SickLeavePayRate,
			"remote_pay_rate":             policy.RemotePayRate,
			"business_trip_pay_rate":      policy.BusinessTripPayRate,
			"unpaid_leave_pay_rate":       policy.UnpaidLeavePayRate,
		},
	}

	snapshotBytes, err := json.Marshal(snapshot)
	if err != nil {
		return nil, fmt.Errorf("%w: snapshot: %v", ErrDatabaseSaveFailed, err)
	}

	item := payrollRepository.UpsertPayrollItemParams{
		ID:                   uuid.New(),
		CycleID:              cycle.ID,
		OrgID:                cycle.OrgID,
		EmployeeID:           employee.ID,
		BaseSalary:           baseSalary.String(),
		AttendanceAdjustment: attendanceAdjustment.String(),
		OvertimeAmount:       overtimeAmount.String(),
		BonusesTotal:         bonusesTotal.String(),
		DeductionsTotal:      deductionsTotal.String(),
		TaxesTotal:           taxesTotal.String(),
		GrossSalary:          grossSalary.String(),
		NetSalary:            netSalary.String(),
		Currency:             withDefault(cycle.Currency, "KZT"),
		WorkingDays:          workingDaysCount,
		PaidDays:             fmt.Sprintf("%d.00", paidDays),
		UnpaidDays:           fmt.Sprintf("%d.00", unpaidDays),
		LateDays:             lateDays,
		AbsentDays:           absentDays,
		MissingDays:          int32(len(missingDates)),
		OvertimeMinutes:      overtimeMinutes,
		ReviewRequired:       reviewRequired,
		ReviewReasons:        reviewReasonBytes,
		EmployerTaxesTotal:   employerTaxesTotal.String(),
		TotalEmployerCost:    totalEmployerCost.String(),
		Status:               statusCalculated,
		CalculationSnapshot:  snapshotBytes,
	}

	return &calculationResult{
		Item:     item,
		Response: mapPayrollItemParams(item),
	}, nil
}

func (s *payrollService) calculateTaxes(ctx context.Context, orgID uuid.UUID, asOf time.Time, taxableIncome, baseSalary, grossSalary Money) (Money, Money, []snapshotTaxRule, error) {
	rules, err := s.repo.ListActiveTaxRules(ctx, orgID, asOf)
	if err != nil {
		return ZeroMoney(), ZeroMoney(), nil, fmt.Errorf("%w: tax rules: %v", ErrDatabaseQueryFailed, err)
	}

	employeeTotal := ZeroMoney()
	employerTotal := ZeroMoney()
	snapshotRules := make([]snapshotTaxRule, 0, len(rules))
	for _, rule := range rules {
		rate, err := ParseRateMicros(rule.Rate)
		if err != nil {
			return ZeroMoney(), ZeroMoney(), nil, ErrInvalidRate
		}

		var base Money
		switch rule.AppliesTo {
		case "TAXABLE_INCOME":
			base = taxableIncome
		case "GROSS":
			base = grossSalary
		case "BASE_ONLY":
			base = baseSalary
		default:
			return ZeroMoney(), ZeroMoney(), nil, ErrInvalidTaxAppliesTo
		}

		if rule.ThresholdMin != "" {
			min, err := ParseMoney(rule.ThresholdMin)
			if err != nil {
				return ZeroMoney(), ZeroMoney(), nil, ErrInvalidAmount
			}
			if base.cents < min.cents {
				continue
			}
		}
		if rule.ThresholdMax != "" {
			max, err := ParseMoney(rule.ThresholdMax)
			if err != nil {
				return ZeroMoney(), ZeroMoney(), nil, ErrInvalidAmount
			}
			if base.cents > max.cents {
				continue
			}
		}

		tax := base.MulRate(rate)
		if rule.Payer == "EMPLOYER" {
			employerTotal = employerTotal.Add(tax)
		} else {
			employeeTotal = employeeTotal.Add(tax)
		}
		snapshotRules = append(snapshotRules, snapshotTaxRule{
			ID:        rule.ID.String(),
			Name:      rule.Name,
			Rate:      rule.Rate,
			AppliesTo: rule.AppliesTo,
			Payer:     rule.Payer,
			TaxAmount: tax.String(),
		})
	}
	return employeeTotal, employerTotal, snapshotRules, nil
}

func (s *payrollService) getPolicy(ctx context.Context, orgID uuid.UUID) (payrollRepository.PayrollPolicy, error) {
	policy, err := s.repo.GetPolicy(ctx, orgID)
	if err == nil {
		return *policy, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return payrollRepository.PayrollPolicy{}, fmt.Errorf("%w: policy: %v", ErrDatabaseQueryFailed, err)
	}
	return defaultPolicy(orgID), nil
}

func defaultPolicy(orgID uuid.UUID) payrollRepository.PayrollPolicy {
	return payrollRepository.PayrollPolicy{
		OrgID:                     orgID,
		MissingAttendancePolicy:   "PAID_REVIEW_REQUIRED",
		RegularOvertimeMultiplier: "1.500000",
		HolidayOvertimeMultiplier: "2.000000",
		LatePenaltyMode:           "NONE",
		LatePenaltyAmount:         "0.00",
		RoundingMode:              "CENT",
		VacationPayRate:           "1.000000",
		SickLeavePayRate:          "1.000000",
		RemotePayRate:             "1.000000",
		BusinessTripPayRate:       "1.000000",
		UnpaidLeavePayRate:        "0.000000",
	}
}

func workingDateList(start, end time.Time, calendarDays []payrollRepository.CalendarDay) []time.Time {
	overrides := make(map[string]string, len(calendarDays))
	for _, day := range calendarDays {
		overrides[day.Date.Format(dateLayout)] = day.Type
	}

	var result []time.Time
	for date := truncateDate(start); !date.After(truncateDate(end)); date = date.AddDate(0, 0, 1) {
		key := date.Format(dateLayout)
		switch overrides[key] {
		case "HOLIDAY":
			continue
		case "WORKDAY":
			result = append(result, date)
			continue
		}
		if date.Weekday() != time.Saturday && date.Weekday() != time.Sunday {
			result = append(result, date)
		}
	}
	return result
}

func approvedLeavesByDate(leaves []payrollRepository.ApprovedLeave, start, end time.Time) map[string]payrollRepository.ApprovedLeave {
	result := map[string]payrollRepository.ApprovedLeave{}
	for _, leave := range leaves {
		from := maxDate(truncateDate(leave.StartDate), truncateDate(start))
		to := minDate(truncateDate(leave.EndDate), truncateDate(end))
		for date := from; !date.After(to); date = date.AddDate(0, 0, 1) {
			result[date.Format(dateLayout)] = leave
		}
	}
	return result
}

func attendanceOverridesByDate(overrides []payrollRepository.AttendanceOverride) map[string]payrollRepository.AttendanceOverride {
	result := make(map[string]payrollRepository.AttendanceOverride, len(overrides))
	for _, override := range overrides {
		result[override.Date.Format(dateLayout)] = override
	}
	return result
}

func filterEmploymentDates(dates []time.Time, hireDate, terminationDate *time.Time) []time.Time {
	filtered := make([]time.Time, 0, len(dates))
	for _, date := range dates {
		if hireDate != nil && date.Before(truncateDate(*hireDate)) {
			continue
		}
		if terminationDate != nil && date.After(truncateDate(*terminationDate)) {
			continue
		}
		filtered = append(filtered, date)
	}
	return filtered
}

func salaryForDate(date time.Time, fallback Money, history []payrollRepository.SalaryHistoryItem) Money {
	for i := len(history) - 1; i >= 0; i-- {
		item := history[i]
		from := truncateDate(item.EffectiveFrom)
		if date.Before(from) {
			continue
		}
		if item.EffectiveTo != nil && date.After(truncateDate(*item.EffectiveTo)) {
			continue
		}
		money, err := ParseMoney(item.SalaryRate)
		if err == nil {
			return money
		}
	}
	return fallback
}

func roundMoney(value Money, mode string) Money {
	switch mode {
	case "WHOLE_TENGE":
		if value.cents >= 0 {
			return Money{cents: ((value.cents + 50) / 100) * 100}
		}
		return Money{cents: ((value.cents - 50) / 100) * 100}
	case "FLOOR":
		return Money{cents: (value.cents / 100) * 100}
	case "CEIL":
		if value.cents%100 == 0 {
			return value
		}
		if value.cents > 0 {
			return Money{cents: (value.cents/100 + 1) * 100}
		}
		return Money{cents: (value.cents / 100) * 100}
	default:
		return value
	}
}

func leavePayRate(policy payrollRepository.PayrollPolicy, leaveType string) int64 {
	var raw string
	switch leaveType {
	case "VACATION":
		raw = policy.VacationPayRate
	case "SICK_LEAVE":
		raw = policy.SickLeavePayRate
	case "REMOTE":
		raw = policy.RemotePayRate
	case "BUSINESS_TRIP":
		raw = policy.BusinessTripPayRate
	default:
		raw = policy.UnpaidLeavePayRate
	}
	rate, err := ParseRateMicros(raw)
	if err != nil {
		return rateScale
	}
	return rate
}

func defaultWorkSchedule(employeeID uuid.UUID) payrollRepository.WorkSchedule {
	return payrollRepository.WorkSchedule{
		EmployeeID:           employeeID,
		WorkStart:            "09:00:00",
		WorkEnd:              "18:00:00",
		LateThresholdMinutes: 15,
	}
}

func calculateOvertimeMinutes(record payrollRepository.AttendanceRecord, schedule payrollRepository.WorkSchedule, allWorkedTime bool) int32 {
	if record.CheckIn == nil || record.CheckOut == nil || !record.CheckOut.After(*record.CheckIn) {
		return 0
	}
	workedMinutes := int32(record.CheckOut.Sub(*record.CheckIn).Minutes())
	if allWorkedTime {
		return workedMinutes
	}
	scheduledMinutes := scheduleMinutes(schedule)
	if workedMinutes <= scheduledMinutes {
		return 0
	}
	return workedMinutes - scheduledMinutes
}

func overtimePay(monthlySalary Money, workingDays int32, schedule payrollRepository.WorkSchedule, minutes int32, multiplierMicros int64) Money {
	denominator := int64(workingDays) * int64(scheduleMinutes(schedule))
	if denominator <= 0 {
		return ZeroMoney()
	}
	base := Money{cents: monthlySalary.cents * int64(minutes) / denominator}
	return base.MulRate(multiplierMicros)
}

func scheduleMinutes(schedule payrollRepository.WorkSchedule) int32 {
	start, err := parseClock(schedule.WorkStart)
	if err != nil {
		return 8 * 60
	}
	end, err := parseClock(schedule.WorkEnd)
	if err != nil || !end.After(start) {
		return 8 * 60
	}
	return int32(end.Sub(start).Minutes())
}

func parseClock(value string) (time.Time, error) {
	if t, err := time.Parse("15:04:05", value); err == nil {
		return t, nil
	}
	return time.Parse("15:04", value)
}

func truncateDate(value time.Time) time.Time {
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, time.UTC)
}

func maxDate(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}

func minDate(a, b time.Time) time.Time {
	if a.Before(b) {
		return a
	}
	return b
}

func mapPayrollItemParams(item payrollRepository.UpsertPayrollItemParams) PayrollItemResponse {
	return PayrollItemResponse{
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
		ReviewReasons:        json.RawMessage(item.ReviewReasons),
		EmployerTaxesTotal:   item.EmployerTaxesTotal,
		TotalEmployerCost:    item.TotalEmployerCost,
		Status:               item.Status,
		Snapshot:             json.RawMessage(item.CalculationSnapshot),
	}
}
