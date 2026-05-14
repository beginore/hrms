package service

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	payrollRepository "hrms/internal/feature/payroll/repository"
	"time"

	"github.com/google/uuid"
)

type PayrollService interface {
	CreateCycle(ctx context.Context, callerUserID uuid.UUID, req CreateCycleRequest) (*PayrollCycleResponse, error)
	ListCycles(ctx context.Context, callerUserID uuid.UUID) ([]PayrollCycleResponse, error)
	GetCycle(ctx context.Context, callerUserID uuid.UUID, id uuid.UUID) (*PayrollCycleResponse, error)
	CalculateCycle(ctx context.Context, callerUserID uuid.UUID, id uuid.UUID, req CalculateCycleRequest) ([]PayrollItemResponse, error)
	ApproveCycle(ctx context.Context, callerUserID uuid.UUID, id uuid.UUID) error
	PreparePayment(ctx context.Context, callerUserID uuid.UUID, id uuid.UUID) error
	MarkCyclePaid(ctx context.Context, callerUserID uuid.UUID, id uuid.UUID) error

	CreateAdjustment(ctx context.Context, callerUserID uuid.UUID, req CreateAdjustmentRequest) (*PayrollAdjustmentResponse, error)
	ListAdjustments(ctx context.Context, callerUserID uuid.UUID, cycleID, employeeID string) ([]PayrollAdjustmentResponse, error)
	DeleteAdjustment(ctx context.Context, callerUserID uuid.UUID, id uuid.UUID) error

	CreateTaxRule(ctx context.Context, callerUserID uuid.UUID, req CreateTaxRuleRequest) (*PayrollTaxRuleResponse, error)
	ListTaxRules(ctx context.Context, callerUserID uuid.UUID) ([]PayrollTaxRuleResponse, error)
	GetPolicy(ctx context.Context, callerUserID uuid.UUID) (*PayrollPolicyResponse, error)
	UpsertPolicy(ctx context.Context, callerUserID uuid.UUID, req UpsertPolicyRequest) (*PayrollPolicyResponse, error)
	CreateCorrection(ctx context.Context, callerUserID uuid.UUID, req CreateCorrectionRequest) (*PayrollCorrectionResponse, error)
	CreateSalaryHistory(ctx context.Context, callerUserID uuid.UUID, req CreateSalaryHistoryRequest) (*SalaryHistoryResponse, error)
	UpsertAttendanceOverride(ctx context.Context, callerUserID uuid.UUID, req UpsertAttendanceOverrideRequest) (*AttendanceOverrideResponse, error)
	UpdateEmploymentPeriod(ctx context.Context, callerUserID uuid.UUID, employeeID uuid.UUID, req UpdateEmploymentPeriodRequest) (*EmploymentPeriodResponse, error)

	PreviewPayroll(ctx context.Context, callerUserID uuid.UUID, req PreviewPayrollRequest) (*PayrollItemResponse, error)
	PreviewCycle(ctx context.Context, callerUserID uuid.UUID, cycleID uuid.UUID) ([]PayrollItemResponse, error)
	ListPayrollItems(ctx context.Context, callerUserID uuid.UUID, cycleID uuid.UUID) ([]PayrollItemResponse, error)
	GetPayrollItem(ctx context.Context, callerUserID uuid.UUID, id uuid.UUID) (*PayrollItemResponse, error)
	ReviewPayrollItem(ctx context.Context, callerUserID uuid.UUID, id uuid.UUID, req ReviewPayrollItemRequest) (*PayrollItemResponse, error)
	ReopenCycle(ctx context.Context, callerUserID uuid.UUID, id uuid.UUID) error
	ExportCycleCSV(ctx context.Context, callerUserID uuid.UUID, id uuid.UUID) ([]byte, error)
	CreateKZTaxPreset(ctx context.Context, callerUserID uuid.UUID) ([]PayrollTaxRuleResponse, error)
}

type payrollService struct {
	repo payrollRepository.PayrollRepository
}

func NewPayrollService(repo payrollRepository.PayrollRepository) PayrollService {
	return &payrollService{repo: repo}
}

func (s *payrollService) CreateCycle(ctx context.Context, callerUserID uuid.UUID, req CreateCycleRequest) (*PayrollCycleResponse, error) {
	actor, err := s.requirePayrollManager(ctx, callerUserID)
	if err != nil {
		return nil, err
	}
	start, end, err := parseDateRange(req.PeriodStart, req.PeriodEnd)
	if err != nil {
		return nil, err
	}

	cycle, err := s.repo.CreateCycle(ctx, payrollRepository.CreateCycleParams{
		ID:          uuid.New(),
		OrgID:       actor.OrgID,
		PeriodStart: start,
		PeriodEnd:   end,
		Currency:    withDefault(req.Currency, "KZT"),
		CreatedBy:   callerUserID,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: create cycle: %v", ErrDatabaseSaveFailed, err)
	}
	_ = s.repo.InsertAuditLog(ctx, actor.OrgID, &cycle.ID, nil, callerUserID, "CREATE_CYCLE", nil, mustJSON(cycle))
	resp := mapCycle(*cycle)
	return &resp, nil
}

func (s *payrollService) ListCycles(ctx context.Context, callerUserID uuid.UUID) ([]PayrollCycleResponse, error) {
	actor, err := s.requirePayrollManager(ctx, callerUserID)
	if err != nil {
		return nil, err
	}
	cycles, err := s.repo.ListCycles(ctx, actor.OrgID)
	if err != nil {
		return nil, fmt.Errorf("%w: list cycles: %v", ErrDatabaseQueryFailed, err)
	}
	return mapCycles(cycles), nil
}

func (s *payrollService) GetCycle(ctx context.Context, callerUserID uuid.UUID, id uuid.UUID) (*PayrollCycleResponse, error) {
	actor, err := s.requirePayrollManager(ctx, callerUserID)
	if err != nil {
		return nil, err
	}
	cycle, err := s.repo.GetCycle(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrCycleNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("%w: get cycle: %v", ErrDatabaseQueryFailed, err)
	}
	if cycle.OrgID != actor.OrgID {
		return nil, ErrCycleNotFound
	}
	resp := mapCycle(*cycle)
	return &resp, nil
}

func (s *payrollService) CalculateCycle(ctx context.Context, callerUserID uuid.UUID, id uuid.UUID, req CalculateCycleRequest) ([]PayrollItemResponse, error) {
	actor, err := s.requirePayrollManager(ctx, callerUserID)
	if err != nil {
		return nil, err
	}
	cycle, err := s.repo.GetCycle(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrCycleNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("%w: cycle: %v", ErrDatabaseQueryFailed, err)
	}
	if cycle.OrgID != actor.OrgID {
		return nil, ErrCycleNotFound
	}
	if isPayrollLocked(cycle.Status) {
		return nil, ErrCycleLocked
	}

	employees, err := s.repo.ListActiveEmployees(ctx, actor.OrgID)
	if err != nil {
		return nil, fmt.Errorf("%w: employees: %v", ErrDatabaseQueryFailed, err)
	}

	responses := make([]PayrollItemResponse, 0, len(employees))
	for _, employee := range employees {
		calculated, err := s.calculateEmployee(ctx, *cycle, employee)
		if err != nil {
			return nil, err
		}
		item, err := s.repo.UpsertPayrollItem(ctx, calculated.Item)
		if err != nil {
			return nil, fmt.Errorf("%w: payroll item: %v", ErrDatabaseSaveFailed, err)
		}
		_ = s.repo.InsertAuditLog(ctx, actor.OrgID, &cycle.ID, &item.ID, callerUserID, "CALCULATE_ITEM", nil, item.CalculationSnapshot)
		responses = append(responses, mapPayrollItem(*item))
	}
	if err := s.repo.UpdateCycleStatus(ctx, cycle.ID, statusCalculated, callerUserID); err != nil {
		return nil, fmt.Errorf("%w: cycle status: %v", ErrDatabaseSaveFailed, err)
	}
	_ = s.repo.InsertAuditLog(ctx, actor.OrgID, &cycle.ID, nil, callerUserID, "CALCULATE_CYCLE", nil, mustJSON(responses))
	return responses, nil
}

func (s *payrollService) ApproveCycle(ctx context.Context, callerUserID uuid.UUID, id uuid.UUID) error {
	cycle, orgID, err := s.getCycleForActor(ctx, callerUserID, id)
	if err != nil {
		return err
	}
	if cycle.Status != statusCalculated {
		return ErrCycleLocked
	}
	items, err := s.repo.ListPayrollItems(ctx, id)
	if err != nil {
		return fmt.Errorf("%w: approve items: %v", ErrDatabaseQueryFailed, err)
	}
	if len(items) == 0 {
		return ErrCycleLocked
	}
	if hasReviewRequiredItems(items) {
		return ErrCycleLocked
	}
	if err := s.repo.UpdateCycleStatus(ctx, id, statusApproved, callerUserID); err != nil {
		return fmt.Errorf("%w: approve: %v", ErrDatabaseSaveFailed, err)
	}
	_ = s.repo.InsertAuditLog(ctx, orgID, &id, nil, callerUserID, "APPROVE_CYCLE", mustJSON(cycle), nil)
	return nil
}

func (s *payrollService) PreparePayment(ctx context.Context, callerUserID uuid.UUID, id uuid.UUID) error {
	cycle, orgID, err := s.getCycleForActor(ctx, callerUserID, id)
	if err != nil {
		return err
	}
	if cycle.Status != statusApproved {
		return ErrCycleLocked
	}
	items, err := s.repo.ListPayrollItems(ctx, id)
	if err != nil {
		return fmt.Errorf("%w: prepare payment items: %v", ErrDatabaseQueryFailed, err)
	}
	if len(items) == 0 || hasReviewRequiredItems(items) {
		return ErrCycleLocked
	}
	if err := s.repo.UpdateCycleStatus(ctx, id, statusReady, callerUserID); err != nil {
		return fmt.Errorf("%w: prepare payment: %v", ErrDatabaseSaveFailed, err)
	}
	_ = s.repo.InsertAuditLog(ctx, orgID, &id, nil, callerUserID, "PREPARE_PAYMENT", mustJSON(cycle), mustJSON(items))
	return nil
}

func (s *payrollService) MarkCyclePaid(ctx context.Context, callerUserID uuid.UUID, id uuid.UUID) error {
	cycle, orgID, err := s.getCycleForActor(ctx, callerUserID, id)
	if err != nil {
		return err
	}
	if cycle.Status != statusReady {
		return ErrCycleLocked
	}
	if err := s.repo.MarkCyclePaid(ctx, id); err != nil {
		return fmt.Errorf("%w: mark paid: %v", ErrDatabaseSaveFailed, err)
	}
	_ = s.repo.InsertAuditLog(ctx, orgID, &id, nil, callerUserID, "MARK_CYCLE_PAID", mustJSON(cycle), nil)
	return nil
}

func (s *payrollService) CreateAdjustment(ctx context.Context, callerUserID uuid.UUID, req CreateAdjustmentRequest) (*PayrollAdjustmentResponse, error) {
	actor, err := s.requirePayrollManager(ctx, callerUserID)
	if err != nil {
		return nil, err
	}
	if req.Type != typeBonus && req.Type != typeDeduction {
		return nil, ErrInvalidAdjustmentType
	}
	if _, err := ParseMoney(req.Amount); err != nil {
		return nil, ErrInvalidAmount
	}
	employeeID, err := uuid.Parse(req.EmployeeID)
	if err != nil {
		return nil, ErrInvalidEmployeeID
	}
	if _, err := s.ensureEmployeeInOrg(ctx, employeeID, actor.OrgID); err != nil {
		return nil, err
	}
	var cycleID *uuid.UUID
	if req.CycleID != "" {
		id, err := uuid.Parse(req.CycleID)
		if err != nil {
			return nil, ErrInvalidCycleID
		}
		cycle, err := s.getCycleInOrg(ctx, id, actor.OrgID)
		if err != nil {
			return nil, err
		}
		if isPayrollLocked(cycle.Status) {
			return nil, ErrCycleLocked
		}
		cycleID = &id
	}
	adjustment, err := s.repo.CreateAdjustment(ctx, payrollRepository.CreateAdjustmentParams{
		ID:         uuid.New(),
		OrgID:      actor.OrgID,
		EmployeeID: employeeID,
		CycleID:    cycleID,
		Type:       req.Type,
		Category:   req.Category,
		Amount:     req.Amount,
		Currency:   withDefault(req.Currency, "KZT"),
		IsTaxable:  req.IsTaxable,
		Reason:     req.Reason,
		CreatedBy:  callerUserID,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: adjustment: %v", ErrDatabaseSaveFailed, err)
	}
	_ = s.repo.InsertAuditLog(ctx, actor.OrgID, cycleID, nil, callerUserID, "CREATE_ADJUSTMENT", nil, mustJSON(adjustment))
	resp := mapAdjustment(*adjustment)
	return &resp, nil
}

func (s *payrollService) ListAdjustments(ctx context.Context, callerUserID uuid.UUID, cycleID, employeeID string) ([]PayrollAdjustmentResponse, error) {
	actor, err := s.requirePayrollManager(ctx, callerUserID)
	if err != nil {
		return nil, err
	}
	parsedCycleID, err := optionalUUID(cycleID, ErrInvalidCycleID)
	if err != nil {
		return nil, err
	}
	parsedEmployeeID, err := optionalUUID(employeeID, ErrInvalidEmployeeID)
	if err != nil {
		return nil, err
	}
	adjustments, err := s.repo.ListAdjustments(ctx, actor.OrgID, parsedCycleID, parsedEmployeeID)
	if err != nil {
		return nil, fmt.Errorf("%w: adjustments: %v", ErrDatabaseQueryFailed, err)
	}
	return mapAdjustments(adjustments), nil
}

func (s *payrollService) DeleteAdjustment(ctx context.Context, callerUserID uuid.UUID, id uuid.UUID) error {
	actor, err := s.requirePayrollManager(ctx, callerUserID)
	if err != nil {
		return err
	}
	if err := s.repo.DeleteAdjustment(ctx, id, actor.OrgID); err != nil {
		return fmt.Errorf("%w: delete adjustment: %v", ErrDatabaseSaveFailed, err)
	}
	_ = s.repo.InsertAuditLog(ctx, actor.OrgID, nil, nil, callerUserID, "DELETE_ADJUSTMENT", mustJSON(map[string]string{"id": id.String()}), nil)
	return nil
}

func (s *payrollService) CreateTaxRule(ctx context.Context, callerUserID uuid.UUID, req CreateTaxRuleRequest) (*PayrollTaxRuleResponse, error) {
	actor, err := s.requirePayrollManager(ctx, callerUserID)
	if err != nil {
		return nil, err
	}
	if _, err := ParseRateMicros(req.Rate); err != nil {
		return nil, ErrInvalidRate
	}
	if req.AppliesTo != "GROSS" && req.AppliesTo != "TAXABLE_INCOME" && req.AppliesTo != "BASE_ONLY" {
		return nil, ErrInvalidTaxAppliesTo
	}
	payer := req.Payer
	if payer == "" {
		payer = "EMPLOYEE"
	}
	if payer != "EMPLOYEE" && payer != "EMPLOYER" {
		return nil, ErrInvalidTaxPayer
	}
	effectiveFrom, err := parseDate(req.EffectiveFrom)
	if err != nil {
		return nil, ErrInvalidDateFormat
	}
	var effectiveTo *time.Time
	if req.EffectiveTo != "" {
		date, err := parseDate(req.EffectiveTo)
		if err != nil {
			return nil, ErrInvalidDateFormat
		}
		effectiveTo = &date
	}
	country := req.Country
	if country == "" {
		country = "KZ"
	}
	rule, err := s.repo.CreateTaxRule(ctx, payrollRepository.CreateTaxRuleParams{
		ID:            uuid.New(),
		OrgID:         &actor.OrgID,
		Country:       country,
		Name:          req.Name,
		Rate:          req.Rate,
		AppliesTo:     req.AppliesTo,
		Payer:         payer,
		ThresholdMin:  req.ThresholdMin,
		ThresholdMax:  req.ThresholdMax,
		EffectiveFrom: effectiveFrom,
		EffectiveTo:   effectiveTo,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: tax rule: %v", ErrDatabaseSaveFailed, err)
	}
	_ = s.repo.InsertAuditLog(ctx, actor.OrgID, nil, nil, callerUserID, "CREATE_TAX_RULE", nil, mustJSON(rule))
	resp := mapTaxRule(*rule)
	return &resp, nil
}

func (s *payrollService) ListTaxRules(ctx context.Context, callerUserID uuid.UUID) ([]PayrollTaxRuleResponse, error) {
	actor, err := s.requirePayrollManager(ctx, callerUserID)
	if err != nil {
		return nil, err
	}
	rules, err := s.repo.ListActiveTaxRules(ctx, actor.OrgID, time.Now())
	if err != nil {
		return nil, fmt.Errorf("%w: tax rules: %v", ErrDatabaseQueryFailed, err)
	}
	return mapTaxRules(rules), nil
}

func (s *payrollService) PreviewPayroll(ctx context.Context, callerUserID uuid.UUID, req PreviewPayrollRequest) (*PayrollItemResponse, error) {
	actor, err := s.getActor(ctx, callerUserID)
	if err != nil {
		return nil, err
	}
	employeeID, err := uuid.Parse(req.EmployeeID)
	if err != nil {
		return nil, ErrInvalidEmployeeID
	}
	start, end, err := parseDateRange(req.PeriodStart, req.PeriodEnd)
	if err != nil {
		return nil, err
	}
	employee, err := s.repo.GetEmployee(ctx, employeeID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrEmployeeNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("%w: employee: %v", ErrDatabaseQueryFailed, err)
	}
	if employee.OrgID != actor.OrgID {
		return nil, ErrEmployeeNotFound
	}
	if !canManagePayroll(actor.Role) && employee.ID != actor.EmployeeID {
		return nil, ErrForbidden
	}
	cycle := payrollRepository.PayrollCycle{
		ID:          uuid.New(),
		OrgID:       actor.OrgID,
		PeriodStart: start,
		PeriodEnd:   end,
		Currency:    "KZT",
		Status:      statusDraft,
		CreatedBy:   callerUserID,
	}
	result, err := s.calculateEmployee(ctx, cycle, *employee)
	if err != nil {
		return nil, err
	}
	return &result.Response, nil
}

func (s *payrollService) ListPayrollItems(ctx context.Context, callerUserID uuid.UUID, cycleID uuid.UUID) ([]PayrollItemResponse, error) {
	actor, err := s.requirePayrollManager(ctx, callerUserID)
	if err != nil {
		return nil, err
	}
	cycle, err := s.getCycleInOrg(ctx, cycleID, actor.OrgID)
	if err != nil {
		return nil, err
	}
	items, err := s.repo.ListPayrollItems(ctx, cycle.ID)
	if err != nil {
		return nil, fmt.Errorf("%w: items: %v", ErrDatabaseQueryFailed, err)
	}
	return mapPayrollItems(items), nil
}

func (s *payrollService) GetPayrollItem(ctx context.Context, callerUserID uuid.UUID, id uuid.UUID) (*PayrollItemResponse, error) {
	actor, err := s.getActor(ctx, callerUserID)
	if err != nil {
		return nil, err
	}
	item, err := s.repo.GetPayrollItem(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrPayrollItemNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("%w: item: %v", ErrDatabaseQueryFailed, err)
	}
	if item.OrgID != actor.OrgID {
		return nil, ErrPayrollItemNotFound
	}
	if !canManagePayroll(actor.Role) && item.EmployeeID != actor.EmployeeID {
		return nil, ErrForbidden
	}
	resp := mapPayrollItem(*item)
	return &resp, nil
}

func (s *payrollService) GetPolicy(ctx context.Context, callerUserID uuid.UUID) (*PayrollPolicyResponse, error) {
	actor, err := s.requirePayrollManager(ctx, callerUserID)
	if err != nil {
		return nil, err
	}
	policy, err := s.getPolicy(ctx, actor.OrgID)
	if err != nil {
		return nil, err
	}
	resp := mapPolicy(policy)
	return &resp, nil
}

func (s *payrollService) UpsertPolicy(ctx context.Context, callerUserID uuid.UUID, req UpsertPolicyRequest) (*PayrollPolicyResponse, error) {
	actor, err := s.requirePayrollManager(ctx, callerUserID)
	if err != nil {
		return nil, err
	}
	params := payrollRepository.UpsertPolicyParams{
		OrgID:                     actor.OrgID,
		MissingAttendancePolicy:   withDefault(req.MissingAttendancePolicy, "PAID_REVIEW_REQUIRED"),
		RegularOvertimeMultiplier: withDefault(req.RegularOvertimeMultiplier, "1.500000"),
		HolidayOvertimeMultiplier: withDefault(req.HolidayOvertimeMultiplier, "2.000000"),
		LatePenaltyMode:           withDefault(req.LatePenaltyMode, "NONE"),
		LatePenaltyAmount:         withDefault(req.LatePenaltyAmount, "0.00"),
		RoundingMode:              withDefault(req.RoundingMode, "CENT"),
		VacationPayRate:           withDefault(req.VacationPayRate, "1.000000"),
		SickLeavePayRate:          withDefault(req.SickLeavePayRate, "1.000000"),
		RemotePayRate:             withDefault(req.RemotePayRate, "1.000000"),
		BusinessTripPayRate:       withDefault(req.BusinessTripPayRate, "1.000000"),
		UnpaidLeavePayRate:        withDefault(req.UnpaidLeavePayRate, "0.000000"),
	}
	if params.MissingAttendancePolicy != "PAID_REVIEW_REQUIRED" && params.MissingAttendancePolicy != "ABSENT_UNPAID" {
		return nil, ErrInvalidPolicy
	}
	if params.LatePenaltyMode != "NONE" && params.LatePenaltyMode != "FIXED_PER_LATE_DAY" {
		return nil, ErrInvalidPolicy
	}
	for _, rate := range []string{params.RegularOvertimeMultiplier, params.HolidayOvertimeMultiplier, params.VacationPayRate, params.SickLeavePayRate, params.RemotePayRate, params.BusinessTripPayRate, params.UnpaidLeavePayRate} {
		if _, err := ParseRateMicros(rate); err != nil {
			return nil, ErrInvalidRate
		}
	}
	if _, err := ParseMoney(params.LatePenaltyAmount); err != nil {
		return nil, ErrInvalidAmount
	}
	policy, err := s.repo.UpsertPolicy(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("%w: policy: %v", ErrDatabaseSaveFailed, err)
	}
	_ = s.repo.InsertAuditLog(ctx, actor.OrgID, nil, nil, callerUserID, "UPSERT_PAYROLL_POLICY", nil, mustJSON(policy))
	resp := mapPolicy(*policy)
	return &resp, nil
}

func (s *payrollService) CreateCorrection(ctx context.Context, callerUserID uuid.UUID, req CreateCorrectionRequest) (*PayrollCorrectionResponse, error) {
	actor, err := s.requirePayrollManager(ctx, callerUserID)
	if err != nil {
		return nil, err
	}
	if req.Type != typeBonus && req.Type != typeDeduction {
		return nil, ErrInvalidAdjustmentType
	}
	if _, err := ParseMoney(req.Amount); err != nil {
		return nil, ErrInvalidAmount
	}
	employeeID, err := uuid.Parse(req.EmployeeID)
	if err != nil {
		return nil, ErrInvalidEmployeeID
	}
	if _, err := s.ensureEmployeeInOrg(ctx, employeeID, actor.OrgID); err != nil {
		return nil, err
	}
	targetCycleID, err := uuid.Parse(req.TargetCycleID)
	if err != nil {
		return nil, ErrInvalidCycleID
	}
	targetCycle, err := s.getCycleInOrg(ctx, targetCycleID, actor.OrgID)
	if err != nil {
		return nil, err
	}
	if isPayrollLocked(targetCycle.Status) {
		return nil, ErrCycleLocked
	}
	var sourceItemID *uuid.UUID
	if req.SourceItemID != "" {
		id, err := uuid.Parse(req.SourceItemID)
		if err != nil {
			return nil, ErrPayrollItemNotFound
		}
		sourceItemID = &id
	}
	adjustment, err := s.repo.CreateAdjustment(ctx, payrollRepository.CreateAdjustmentParams{
		ID:         uuid.New(),
		OrgID:      actor.OrgID,
		EmployeeID: employeeID,
		CycleID:    &targetCycleID,
		Type:       req.Type,
		Category:   "CORRECTION",
		Amount:     req.Amount,
		Currency:   "KZT",
		IsTaxable:  req.Taxable,
		Reason:     req.Reason,
		CreatedBy:  callerUserID,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: correction adjustment: %v", ErrDatabaseSaveFailed, err)
	}
	correction, err := s.repo.CreateCorrection(ctx, payrollRepository.CreateCorrectionParams{
		ID:            uuid.New(),
		OrgID:         actor.OrgID,
		EmployeeID:    employeeID,
		SourceItemID:  sourceItemID,
		TargetCycleID: targetCycleID,
		AdjustmentID:  adjustment.ID,
		Amount:        req.Amount,
		Type:          req.Type,
		Reason:        req.Reason,
		CreatedBy:     callerUserID,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: correction: %v", ErrDatabaseSaveFailed, err)
	}
	_ = s.repo.InsertAuditLog(ctx, actor.OrgID, &targetCycleID, nil, callerUserID, "CREATE_CORRECTION", nil, mustJSON(correction))
	resp := mapCorrection(*correction)
	return &resp, nil
}

func (s *payrollService) CreateSalaryHistory(ctx context.Context, callerUserID uuid.UUID, req CreateSalaryHistoryRequest) (*SalaryHistoryResponse, error) {
	actor, err := s.requirePayrollManager(ctx, callerUserID)
	if err != nil {
		return nil, err
	}
	employeeID, err := uuid.Parse(req.EmployeeID)
	if err != nil {
		return nil, ErrInvalidEmployeeID
	}
	if _, err := s.ensureEmployeeInOrg(ctx, employeeID, actor.OrgID); err != nil {
		return nil, err
	}
	if _, err := ParseMoney(req.SalaryRate); err != nil {
		return nil, ErrInvalidAmount
	}
	effectiveFrom, err := parseDate(req.EffectiveFrom)
	if err != nil {
		return nil, ErrInvalidDateFormat
	}
	var effectiveTo *time.Time
	if req.EffectiveTo != "" {
		date, err := parseDate(req.EffectiveTo)
		if err != nil {
			return nil, ErrInvalidDateFormat
		}
		if date.Before(effectiveFrom) {
			return nil, ErrInvalidDateRange
		}
		effectiveTo = &date
	}
	existing, err := s.repo.ListSalaryHistory(ctx, employeeID, effectiveFrom, effectiveToOrFarFuture(effectiveTo))
	if err != nil {
		return nil, fmt.Errorf("%w: salary history overlap: %v", ErrDatabaseQueryFailed, err)
	}
	for _, item := range existing {
		if salaryPeriodsOverlap(effectiveFrom, effectiveTo, item.EffectiveFrom, item.EffectiveTo) {
			return nil, ErrInvalidDateRange
		}
	}
	item, err := s.repo.CreateSalaryHistory(ctx, payrollRepository.CreateSalaryHistoryParams{
		ID:            uuid.New(),
		OrgID:         actor.OrgID,
		EmployeeID:    employeeID,
		SalaryRate:    req.SalaryRate,
		EffectiveFrom: effectiveFrom,
		EffectiveTo:   effectiveTo,
		CreatedBy:     callerUserID,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: salary history: %v", ErrDatabaseSaveFailed, err)
	}
	_ = s.repo.InsertAuditLog(ctx, actor.OrgID, nil, nil, callerUserID, "CREATE_SALARY_HISTORY", nil, mustJSON(item))
	resp := mapSalaryHistory(*item)
	return &resp, nil
}

func (s *payrollService) UpsertAttendanceOverride(ctx context.Context, callerUserID uuid.UUID, req UpsertAttendanceOverrideRequest) (*AttendanceOverrideResponse, error) {
	actor, err := s.requirePayrollManager(ctx, callerUserID)
	if err != nil {
		return nil, err
	}
	employeeID, err := uuid.Parse(req.EmployeeID)
	if err != nil {
		return nil, ErrInvalidEmployeeID
	}
	if _, err := s.ensureEmployeeInOrg(ctx, employeeID, actor.OrgID); err != nil {
		return nil, err
	}
	date, err := parseDate(req.Date)
	if err != nil {
		return nil, ErrInvalidDateFormat
	}
	if req.OvertimeMinutes < 0 {
		return nil, ErrInvalidAmount
	}
	if req.AttendanceStatus != "" && req.AttendanceStatus != "PRESENT" && req.AttendanceStatus != "LATE" && req.AttendanceStatus != "ABSENT" && req.AttendanceStatus != "ON_LEAVE" {
		return nil, ErrInvalidPolicy
	}
	paidDay := true
	if req.PaidDay != nil {
		paidDay = *req.PaidDay
	}
	override, err := s.repo.UpsertAttendanceOverride(ctx, payrollRepository.UpsertAttendanceOverrideParams{
		ID:               uuid.New(),
		OrgID:            actor.OrgID,
		EmployeeID:       employeeID,
		Date:             date,
		PaidDay:          paidDay,
		AttendanceType:   withDefault(req.AttendanceType, "MANUAL"),
		AttendanceStatus: withDefault(req.AttendanceStatus, "PRESENT"),
		OvertimeMinutes:  req.OvertimeMinutes,
		Note:             req.Note,
		CreatedBy:        callerUserID,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: attendance override: %v", ErrDatabaseSaveFailed, err)
	}
	_ = s.repo.InsertAuditLog(ctx, actor.OrgID, nil, nil, callerUserID, "UPSERT_ATTENDANCE_OVERRIDE", nil, mustJSON(override))
	resp := mapAttendanceOverride(*override)
	return &resp, nil
}

func (s *payrollService) UpdateEmploymentPeriod(ctx context.Context, callerUserID uuid.UUID, employeeID uuid.UUID, req UpdateEmploymentPeriodRequest) (*EmploymentPeriodResponse, error) {
	actor, err := s.requirePayrollManager(ctx, callerUserID)
	if err != nil {
		return nil, err
	}
	if _, err := s.ensureEmployeeInOrg(ctx, employeeID, actor.OrgID); err != nil {
		return nil, err
	}
	var hireDate *time.Time
	if req.HireDate != "" {
		date, err := parseDate(req.HireDate)
		if err != nil {
			return nil, ErrInvalidDateFormat
		}
		hireDate = &date
	}
	var terminationDate *time.Time
	if req.TerminationDate != "" {
		date, err := parseDate(req.TerminationDate)
		if err != nil {
			return nil, ErrInvalidDateFormat
		}
		terminationDate = &date
	}
	if hireDate != nil && terminationDate != nil && terminationDate.Before(*hireDate) {
		return nil, ErrInvalidDateRange
	}
	employee, err := s.repo.UpdateEmploymentPeriod(ctx, payrollRepository.UpdateEmploymentPeriodParams{
		EmployeeID:      employeeID,
		OrgID:           actor.OrgID,
		HireDate:        hireDate,
		TerminationDate: terminationDate,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrEmployeeNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("%w: employment period: %v", ErrDatabaseSaveFailed, err)
	}
	_ = s.repo.InsertAuditLog(ctx, actor.OrgID, nil, nil, callerUserID, "UPDATE_EMPLOYMENT_PERIOD", nil, mustJSON(employee))
	resp := mapEmploymentPeriod(*employee)
	return &resp, nil
}

func (s *payrollService) PreviewCycle(ctx context.Context, callerUserID uuid.UUID, cycleID uuid.UUID) ([]PayrollItemResponse, error) {
	actor, err := s.requirePayrollManager(ctx, callerUserID)
	if err != nil {
		return nil, err
	}
	cycle, err := s.getCycleInOrg(ctx, cycleID, actor.OrgID)
	if err != nil {
		return nil, err
	}
	employees, err := s.repo.ListActiveEmployees(ctx, actor.OrgID)
	if err != nil {
		return nil, fmt.Errorf("%w: employees: %v", ErrDatabaseQueryFailed, err)
	}
	responses := make([]PayrollItemResponse, 0, len(employees))
	for _, employee := range employees {
		calculated, err := s.calculateEmployee(ctx, *cycle, employee)
		if err != nil {
			return nil, err
		}
		responses = append(responses, calculated.Response)
	}
	return responses, nil
}

func (s *payrollService) ReviewPayrollItem(ctx context.Context, callerUserID uuid.UUID, id uuid.UUID, req ReviewPayrollItemRequest) (*PayrollItemResponse, error) {
	actor, err := s.requirePayrollManager(ctx, callerUserID)
	if err != nil {
		return nil, err
	}
	existing, err := s.repo.GetPayrollItem(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrPayrollItemNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("%w: item: %v", ErrDatabaseQueryFailed, err)
	}
	if existing.OrgID != actor.OrgID {
		return nil, ErrPayrollItemNotFound
	}
	cycle, err := s.getCycleInOrg(ctx, existing.CycleID, actor.OrgID)
	if err != nil {
		return nil, err
	}
	if isPayrollLocked(cycle.Status) {
		return nil, ErrCycleLocked
	}
	item, err := s.repo.ReviewPayrollItem(ctx, id, callerUserID, req.Comment)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrPayrollItemNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("%w: review item: %v", ErrDatabaseSaveFailed, err)
	}
	_ = s.repo.InsertAuditLog(ctx, actor.OrgID, &item.CycleID, &item.ID, callerUserID, "REVIEW_PAYROLL_ITEM", mustJSON(existing), mustJSON(item))
	resp := mapPayrollItem(*item)
	return &resp, nil
}

func (s *payrollService) ReopenCycle(ctx context.Context, callerUserID uuid.UUID, id uuid.UUID) error {
	cycle, orgID, err := s.getCycleForActor(ctx, callerUserID, id)
	if err != nil {
		return err
	}
	if cycle.Status != statusApproved && cycle.Status != statusReady {
		return ErrCycleLocked
	}
	if err := s.repo.UpdateCycleStatus(ctx, id, statusCalculated, callerUserID); err != nil {
		return fmt.Errorf("%w: reopen: %v", ErrDatabaseSaveFailed, err)
	}
	_ = s.repo.InsertAuditLog(ctx, orgID, &id, nil, callerUserID, "REOPEN_CYCLE", mustJSON(cycle), nil)
	return nil
}

func (s *payrollService) ExportCycleCSV(ctx context.Context, callerUserID uuid.UUID, id uuid.UUID) ([]byte, error) {
	cycle, _, err := s.getCycleForActor(ctx, callerUserID, id)
	if err != nil {
		return nil, err
	}
	if cycle.Status != statusReady && cycle.Status != statusPaid {
		return nil, ErrCycleLocked
	}
	items, err := s.repo.ListPayrollItems(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("%w: export items: %v", ErrDatabaseQueryFailed, err)
	}
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)
	_ = writer.Write([]string{"cycle_id", "period_start", "period_end", "employee_id", "gross_salary", "deductions", "taxes", "net_salary", "currency", "status", "review_required"})
	for _, item := range items {
		_ = writer.Write([]string{
			cycle.ID.String(), cycle.PeriodStart.Format(dateLayout), cycle.PeriodEnd.Format(dateLayout),
			item.EmployeeID.String(), item.GrossSalary, item.DeductionsTotal, item.TaxesTotal,
			item.NetSalary, item.Currency, item.Status,
			fmt.Sprintf("%t", item.ReviewRequired),
		})
	}
	writer.Flush()
	return buf.Bytes(), writer.Error()
}

func (s *payrollService) CreateKZTaxPreset(ctx context.Context, callerUserID uuid.UUID) ([]PayrollTaxRuleResponse, error) {
	actor, err := s.requirePayrollManager(ctx, callerUserID)
	if err != nil {
		return nil, err
	}
	presets := []CreateTaxRuleRequest{
		{Country: "KZ", Name: "KZ Individual Income Tax", Rate: "0.100000", AppliesTo: "TAXABLE_INCOME", Payer: "EMPLOYEE", EffectiveFrom: "2026-01-01"},
		{Country: "KZ", Name: "KZ Employer Social Contribution", Rate: "0.035000", AppliesTo: "GROSS", Payer: "EMPLOYER", EffectiveFrom: "2026-01-01"},
		{Country: "KZ", Name: "KZ Employer Medical Insurance", Rate: "0.030000", AppliesTo: "GROSS", Payer: "EMPLOYER", EffectiveFrom: "2026-01-01"},
	}
	out := make([]PayrollTaxRuleResponse, 0, len(presets))
	for _, preset := range presets {
		rule, err := s.CreateTaxRule(ctx, actor.UserID, preset)
		if err != nil {
			return nil, err
		}
		out = append(out, *rule)
	}
	return out, nil
}

func (s *payrollService) getCycleForActor(ctx context.Context, callerUserID uuid.UUID, id uuid.UUID) (*payrollRepository.PayrollCycle, uuid.UUID, error) {
	actor, err := s.requirePayrollManager(ctx, callerUserID)
	if err != nil {
		return nil, uuid.Nil, err
	}
	cycle, err := s.getCycleInOrg(ctx, id, actor.OrgID)
	return cycle, actor.OrgID, err
}

func (s *payrollService) getCycleInOrg(ctx context.Context, id uuid.UUID, orgID uuid.UUID) (*payrollRepository.PayrollCycle, error) {
	cycle, err := s.repo.GetCycle(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrCycleNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("%w: cycle: %v", ErrDatabaseQueryFailed, err)
	}
	if cycle.OrgID != orgID {
		return nil, ErrCycleNotFound
	}
	return cycle, nil
}

func (s *payrollService) ensureEmployeeInOrg(ctx context.Context, employeeID uuid.UUID, orgID uuid.UUID) (*payrollRepository.EmployeePayrollInput, error) {
	employee, err := s.repo.GetEmployee(ctx, employeeID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrEmployeeNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("%w: employee: %v", ErrDatabaseQueryFailed, err)
	}
	if employee.OrgID != orgID {
		return nil, ErrEmployeeNotFound
	}
	return employee, nil
}

func (s *payrollService) getActor(ctx context.Context, callerUserID uuid.UUID) (*payrollRepository.ActorAccess, error) {
	actor, err := s.repo.GetActorAccess(ctx, callerUserID)
	if err != nil {
		return nil, fmt.Errorf("%w: actor: %v", ErrDatabaseQueryFailed, err)
	}
	return actor, nil
}

func (s *payrollService) requirePayrollManager(ctx context.Context, callerUserID uuid.UUID) (*payrollRepository.ActorAccess, error) {
	actor, err := s.getActor(ctx, callerUserID)
	if err != nil {
		return nil, err
	}
	if !canManagePayroll(actor.Role) {
		return nil, ErrForbidden
	}
	return actor, nil
}

func canManagePayroll(role string) bool {
	switch role {
	case "SysAdmin", "Admin", "HR", "Manager":
		return true
	default:
		return false
	}
}

func isPayrollLocked(status string) bool {
	switch status {
	case statusApproved, statusReady, statusPaid:
		return true
	default:
		return false
	}
}

func hasReviewRequiredItems(items []payrollRepository.PayrollItem) bool {
	for _, item := range items {
		if item.ReviewRequired {
			return true
		}
	}
	return false
}

func parseDateRange(startValue, endValue string) (time.Time, time.Time, error) {
	start, err := parseDate(startValue)
	if err != nil {
		return time.Time{}, time.Time{}, ErrInvalidDateFormat
	}
	end, err := parseDate(endValue)
	if err != nil {
		return time.Time{}, time.Time{}, ErrInvalidDateFormat
	}
	if end.Before(start) {
		return time.Time{}, time.Time{}, ErrInvalidDateRange
	}
	return start, end, nil
}

func parseDate(value string) (time.Time, error) {
	return time.Parse(dateLayout, value)
}

func optionalUUID(value string, invalidErr error) (*uuid.UUID, error) {
	if value == "" {
		return nil, nil
	}
	id, err := uuid.Parse(value)
	if err != nil {
		return nil, invalidErr
	}
	return &id, nil
}

func mustJSON(value any) []byte {
	out, _ := json.Marshal(value)
	return out
}

func withDefault(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func effectiveToOrFarFuture(value *time.Time) time.Time {
	if value != nil {
		return *value
	}
	return time.Date(9999, 12, 31, 0, 0, 0, 0, time.UTC)
}

func salaryPeriodsOverlap(aFrom time.Time, aTo *time.Time, bFrom time.Time, bTo *time.Time) bool {
	aEnd := effectiveToOrFarFuture(aTo)
	bEnd := effectiveToOrFarFuture(bTo)
	return !aFrom.After(bEnd) && !bFrom.After(aEnd)
}
