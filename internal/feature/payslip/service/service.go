package service

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	notificationService "hrms/internal/feature/notification/service"
	payslipRepository "hrms/internal/feature/payslip/repository"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	payslipStatusGenerated = "GENERATED"
	payslipStatusSent      = "SENT"
	payslipStatusVoid      = "VOID"

	cycleStatusApproved = "APPROVED"
	cycleStatusReady    = "READY_FOR_PAYMENT"
	cycleStatusPaid     = "PAID"
)

type PayslipMailer interface {
	SendPayslip(ctx context.Context, toEmail, subject, bodyText, attachmentFilename string, attachment []byte) error
}

type PayslipService interface {
	GeneratePayslip(ctx context.Context, callerUserID uuid.UUID, req GeneratePayslipRequest) (*PayslipResponse, error)
	GenerateCyclePayslips(ctx context.Context, callerUserID uuid.UUID, cycleID uuid.UUID) (*PayslipBatchResponse, error)
	ListPayslips(ctx context.Context, callerUserID uuid.UUID, cycleID, employeeID, status, periodStart, periodEnd string) ([]PayslipResponse, error)
	GetPayslip(ctx context.Context, callerUserID uuid.UUID, id uuid.UUID) (*PayslipResponse, error)
	RenderPayslipPDF(ctx context.Context, callerUserID uuid.UUID, id uuid.UUID) ([]byte, error)
	SendPayslip(ctx context.Context, callerUserID uuid.UUID, id uuid.UUID) (*PayslipSendResponse, error)
	SendCyclePayslips(ctx context.Context, callerUserID uuid.UUID, cycleID uuid.UUID) (*PayslipBatchSendResponse, error)
	VoidPayslip(ctx context.Context, callerUserID uuid.UUID, id uuid.UUID, req VoidPayslipRequest) (*PayslipResponse, error)
	RegeneratePayslip(ctx context.Context, callerUserID uuid.UUID, id uuid.UUID) (*PayslipResponse, error)
	ListMyPayslips(ctx context.Context, callerUserID uuid.UUID) ([]PayslipResponse, error)
	GetMyPayslip(ctx context.Context, callerUserID uuid.UUID, id uuid.UUID) (*PayslipResponse, error)
	RenderMyPayslipPDF(ctx context.Context, callerUserID uuid.UUID, id uuid.UUID) ([]byte, error)
}

type payslipService struct {
	repo      payslipRepository.PayslipRepository
	mailer    PayslipMailer
	notifySvc *notificationService.Service
}

func NewPayslipService(repo payslipRepository.PayslipRepository, mailer PayslipMailer, notifySvc *notificationService.Service) PayslipService {
	return &payslipService{repo: repo, mailer: mailer, notifySvc: notifySvc}
}

func (s *payslipService) GeneratePayslip(ctx context.Context, callerUserID uuid.UUID, req GeneratePayslipRequest) (*PayslipResponse, error) {
	actor, err := s.requirePayslipManager(ctx, callerUserID)
	if err != nil {
		return nil, err
	}
	payrollItemID, err := uuid.Parse(req.PayrollItemID)
	if err != nil {
		return nil, ErrInvalidPayrollItemID
	}
	payslip, err := s.generatePayslipForItem(ctx, actor, payrollItemID)
	if err != nil {
		return nil, err
	}
	resp := mapPayslip(*payslip)
	return &resp, nil
}

func (s *payslipService) GenerateCyclePayslips(ctx context.Context, callerUserID uuid.UUID, cycleID uuid.UUID) (*PayslipBatchResponse, error) {
	actor, err := s.requirePayslipManager(ctx, callerUserID)
	if err != nil {
		return nil, err
	}
	cycle, err := s.getCycleInOrg(ctx, cycleID, actor.OrgID)
	if err != nil {
		return nil, err
	}
	if !canGeneratePayslipForCycle(cycle.Status) {
		return nil, ErrInvalidCycleStatus
	}
	items, err := s.repo.ListPayrollItemsByCycle(ctx, cycleID)
	if err != nil {
		return nil, fmt.Errorf("%w: payroll items: %v", ErrDatabaseQueryFailed, err)
	}

	result := PayslipBatchResponse{}
	for _, item := range items {
		existing, err := s.repo.GetPayslipByPayrollItem(ctx, item.ID)
		if err == nil {
			result.Skipped = append(result.Skipped, mapPayslip(*existing))
			continue
		}
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("%w: existing payslip: %v", ErrDatabaseQueryFailed, err)
		}
		payslip, err := s.generatePayslipForItem(ctx, actor, item.ID)
		if err != nil {
			if errors.Is(err, ErrReviewRequired) {
				continue
			}
			return nil, err
		}
		resp := mapPayslip(*payslip)
		result.Created = append(result.Created, resp)
	}
	return &result, nil
}

func (s *payslipService) ListPayslips(ctx context.Context, callerUserID uuid.UUID, cycleID, employeeID, status, periodStart, periodEnd string) ([]PayslipResponse, error) {
	actor, err := s.requirePayslipManager(ctx, callerUserID)
	if err != nil {
		return nil, err
	}
	filter, err := buildListFilter(actor.OrgID, cycleID, employeeID, status, periodStart, periodEnd)
	if err != nil {
		return nil, err
	}
	payslips, err := s.repo.ListPayslips(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("%w: payslips: %v", ErrDatabaseQueryFailed, err)
	}
	return mapPayslips(payslips), nil
}

func (s *payslipService) GetPayslip(ctx context.Context, callerUserID uuid.UUID, id uuid.UUID) (*PayslipResponse, error) {
	actor, err := s.requirePayslipManager(ctx, callerUserID)
	if err != nil {
		return nil, err
	}
	payslip, err := s.getPayslipInOrg(ctx, id, actor.OrgID)
	if err != nil {
		return nil, err
	}
	resp := mapPayslip(*payslip)
	return &resp, nil
}

func (s *payslipService) RenderPayslipPDF(ctx context.Context, callerUserID uuid.UUID, id uuid.UUID) ([]byte, error) {
	actor, err := s.requirePayslipManager(ctx, callerUserID)
	if err != nil {
		return nil, err
	}
	payslip, err := s.getPayslipInOrg(ctx, id, actor.OrgID)
	if err != nil {
		return nil, err
	}
	return s.ensurePayslipPDF(ctx, *payslip)
}

func (s *payslipService) SendPayslip(ctx context.Context, callerUserID uuid.UUID, id uuid.UUID) (*PayslipSendResponse, error) {
	actor, err := s.requirePayslipManager(ctx, callerUserID)
	if err != nil {
		return nil, err
	}
	payslip, err := s.getPayslipInOrg(ctx, id, actor.OrgID)
	if err != nil {
		return nil, err
	}
	sent, err := s.sendPayslip(ctx, *payslip)
	if err != nil {
		return nil, err
	}
	resp := mapSendResponse(*sent)
	return &resp, nil
}

func (s *payslipService) SendCyclePayslips(ctx context.Context, callerUserID uuid.UUID, cycleID uuid.UUID) (*PayslipBatchSendResponse, error) {
	actor, err := s.requirePayslipManager(ctx, callerUserID)
	if err != nil {
		return nil, err
	}
	if _, err := s.getCycleInOrg(ctx, cycleID, actor.OrgID); err != nil {
		return nil, err
	}
	payslips, err := s.repo.ListPayslips(ctx, payslipRepository.ListPayslipsFilter{
		OrgID:   actor.OrgID,
		CycleID: &cycleID,
		Status:  payslipStatusGenerated,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: payslips: %v", ErrDatabaseQueryFailed, err)
	}
	out := PayslipBatchSendResponse{}
	for _, payslip := range payslips {
		sent, err := s.sendPayslip(ctx, payslip)
		if err != nil {
			out.Failed = append(out.Failed, PayslipSendFailure{
				PayslipID: payslip.ID.String(),
				Error:     err.Error(),
			})
			continue
		}
		out.Sent = append(out.Sent, mapSendResponse(*sent))
	}
	return &out, nil
}

func (s *payslipService) VoidPayslip(ctx context.Context, callerUserID uuid.UUID, id uuid.UUID, req VoidPayslipRequest) (*PayslipResponse, error) {
	actor, err := s.requirePayslipManager(ctx, callerUserID)
	if err != nil {
		return nil, err
	}
	existing, err := s.getPayslipInOrg(ctx, id, actor.OrgID)
	if err != nil {
		return nil, err
	}
	if existing.Status == payslipStatusVoid {
		return nil, ErrPayslipLocked
	}
	payslip, err := s.repo.VoidPayslip(ctx, id, callerUserID, req.Reason)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrPayslipNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("%w: void payslip: %v", ErrDatabaseSaveFailed, err)
	}
	resp := mapPayslip(*payslip)
	return &resp, nil
}

func (s *payslipService) RegeneratePayslip(ctx context.Context, callerUserID uuid.UUID, id uuid.UUID) (*PayslipResponse, error) {
	actor, err := s.requirePayslipManager(ctx, callerUserID)
	if err != nil {
		return nil, err
	}
	existing, err := s.getPayslipInOrg(ctx, id, actor.OrgID)
	if err != nil {
		return nil, err
	}
	if existing.Status != payslipStatusVoid {
		return nil, ErrPayslipLocked
	}
	payslip, err := s.generatePayslipForItem(ctx, actor, existing.PayrollItemID)
	if err != nil {
		return nil, err
	}
	resp := mapPayslip(*payslip)
	return &resp, nil
}

func (s *payslipService) ListMyPayslips(ctx context.Context, callerUserID uuid.UUID) ([]PayslipResponse, error) {
	actor, err := s.getActor(ctx, callerUserID)
	if err != nil {
		return nil, err
	}
	employeeID := actor.EmployeeID
	payslips, err := s.repo.ListPayslips(ctx, payslipRepository.ListPayslipsFilter{
		OrgID:      actor.OrgID,
		EmployeeID: &employeeID,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: my payslips: %v", ErrDatabaseQueryFailed, err)
	}
	return mapPayslips(payslips), nil
}

func (s *payslipService) GetMyPayslip(ctx context.Context, callerUserID uuid.UUID, id uuid.UUID) (*PayslipResponse, error) {
	payslip, err := s.getSelfPayslip(ctx, callerUserID, id)
	if err != nil {
		return nil, err
	}
	resp := mapPayslip(*payslip)
	return &resp, nil
}

func (s *payslipService) RenderMyPayslipPDF(ctx context.Context, callerUserID uuid.UUID, id uuid.UUID) ([]byte, error) {
	payslip, err := s.getSelfPayslip(ctx, callerUserID, id)
	if err != nil {
		return nil, err
	}
	return s.ensurePayslipPDF(ctx, *payslip)
}

func (s *payslipService) generatePayslipForItem(ctx context.Context, actor *payslipRepository.ActorAccess, payrollItemID uuid.UUID) (*payslipRepository.Payslip, error) {
	existing, err := s.repo.GetPayslipByPayrollItem(ctx, payrollItemID)
	if err == nil {
		return existing, nil
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("%w: existing payslip: %v", ErrDatabaseQueryFailed, err)
	}

	item, err := s.repo.GetPayrollItem(ctx, payrollItemID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrPayrollItemNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("%w: payroll item: %v", ErrDatabaseQueryFailed, err)
	}
	if item.OrgID != actor.OrgID {
		return nil, ErrPayrollItemNotFound
	}
	if item.ReviewRequired {
		return nil, ErrReviewRequired
	}
	cycle, err := s.getCycleInOrg(ctx, item.CycleID, actor.OrgID)
	if err != nil {
		return nil, err
	}
	if !canGeneratePayslipForCycle(cycle.Status) {
		return nil, ErrInvalidCycleStatus
	}
	employee, err := s.repo.GetEmployeeProfile(ctx, item.EmployeeID)
	if err != nil {
		return nil, fmt.Errorf("%w: employee: %v", ErrDatabaseQueryFailed, err)
	}
	org, err := s.repo.GetOrganizationProfile(ctx, item.OrgID)
	if err != nil {
		return nil, fmt.Errorf("%w: organization: %v", ErrDatabaseQueryFailed, err)
	}
	adjustments, err := s.repo.ListPayrollAdjustments(ctx, item.CycleID, item.EmployeeID)
	if err != nil {
		return nil, fmt.Errorf("%w: adjustments: %v", ErrDatabaseQueryFailed, err)
	}
	payslipID := uuid.New()
	payload, err := buildPayload(payslipID, *org, *employee, *cycle, *item, adjustments)
	if err != nil {
		return nil, fmt.Errorf("%w: payload: %v", ErrRenderFailed, err)
	}
	pdf, err := renderPayslipPDF(payload)
	if err != nil {
		return nil, fmt.Errorf("%w: pdf: %v", ErrRenderFailed, err)
	}
	pdfGeneratedAt := time.Now().UTC()
	pdfFilename := payslipFilename(payslipID)
	pdfSHA256 := contentSHA256(pdf)
	payslip, err := s.repo.CreatePayslip(ctx, payslipRepository.CreatePayslipParams{
		ID:                 payslipID,
		OrgID:              item.OrgID,
		EmployeeID:         item.EmployeeID,
		PayrollCycleID:     item.CycleID,
		PayrollItemID:      item.ID,
		PeriodStart:        cycle.PeriodStart,
		PeriodEnd:          cycle.PeriodEnd,
		Status:             payslipStatusGenerated,
		Currency:           item.Currency,
		BaseSalary:         item.BaseSalary,
		OvertimeAmount:     item.OvertimeAmount,
		BonusesTotal:       item.BonusesTotal,
		DeductionsTotal:    item.DeductionsTotal,
		TaxesTotal:         item.TaxesTotal,
		GrossSalary:        item.GrossSalary,
		NetSalary:          item.NetSalary,
		EmployerTaxesTotal: item.EmployerTaxesTotal,
		TotalEmployerCost:  item.TotalEmployerCost,
		PayloadSnapshot:    payload,
		PDFContent:         pdf,
		PDFFilename:        pdfFilename,
		PDFGeneratedAt:     &pdfGeneratedAt,
		PDFSHA256:          pdfSHA256,
		GeneratedBy:        actor.UserID,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: payslip: %v", ErrDatabaseSaveFailed, err)
	}
	return payslip, nil
}

func (s *payslipService) sendPayslip(ctx context.Context, payslip payslipRepository.Payslip) (*payslipRepository.Payslip, error) {
	if payslip.Status == payslipStatusVoid {
		return nil, ErrPayslipLocked
	}
	if s.mailer == nil {
		return nil, fmt.Errorf("%w: mailer is not configured", ErrSendFailed)
	}
	var payload map[string]any
	if err := json.Unmarshal(payslip.PayloadSnapshot, &payload); err != nil {
		return nil, fmt.Errorf("%w: payload: %v", ErrRenderFailed, err)
	}
	email := nestedString(payload, "employee", "email")
	if email == "" {
		return nil, fmt.Errorf("%w: employee email is empty", ErrSendFailed)
	}
	body, err := renderPayslipEmailBody(payslip.PayloadSnapshot)
	if err != nil {
		return nil, fmt.Errorf("%w: email body: %v", ErrRenderFailed, err)
	}
	pdf, err := s.ensurePayslipPDF(ctx, payslip)
	if err != nil {
		return nil, err
	}
	filename := payslip.PDFFilename
	if filename == "" {
		filename = payslipFilename(payslip.ID)
	}
	subject := "Your HRMS payslip is ready"
	if err := s.mailer.SendPayslip(ctx, email, subject, body, filename, pdf); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrSendFailed, err)
	}
	sent, err := s.repo.MarkPayslipSent(ctx, payslip.ID, email)
	if err != nil {
		return nil, fmt.Errorf("%w: mark sent: %v", ErrDatabaseSaveFailed, err)
	}

	if s.notifySvc != nil {
		employee, empErr := s.repo.GetEmployeeProfile(ctx, payslip.EmployeeID)
		if empErr == nil {
			orgIDStr := employee.OrgID.String()
			periodStart := payslip.PeriodStart.Format(dateLayout)
			periodEnd := payslip.PeriodEnd.Format(dateLayout)
			_, _ = s.notifySvc.NotifySalaryToEmployee(ctx, notificationService.NotifyUserRequest{
				UserID:  employee.UserID.String(),
				OrgID:   &orgIDStr,
				Title:   "Your payslip is ready",
				Message: fmt.Sprintf("Your payslip for the period %s – %s has been sent to %s.", periodStart, periodEnd, email),
			})
		}
	}
	return sent, nil
}

func (s *payslipService) ensurePayslipPDF(ctx context.Context, payslip payslipRepository.Payslip) ([]byte, error) {
	if payslip.Status == payslipStatusVoid {
		return nil, ErrPayslipLocked
	}
	if len(payslip.PDFContent) > 0 {
		return payslip.PDFContent, nil
	}
	pdf, err := renderPayslipPDF(payslip.PayloadSnapshot)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrRenderFailed, err)
	}
	filename := payslip.PDFFilename
	if filename == "" {
		filename = payslipFilename(payslip.ID)
	}
	updated, err := s.repo.StorePayslipPDF(ctx, payslip.ID, pdf, filename, contentSHA256(pdf))
	if err != nil {
		return nil, fmt.Errorf("%w: store pdf: %v", ErrDatabaseSaveFailed, err)
	}
	return updated.PDFContent, nil
}

func (s *payslipService) getPayslipInOrg(ctx context.Context, id uuid.UUID, orgID uuid.UUID) (*payslipRepository.Payslip, error) {
	payslip, err := s.repo.GetPayslip(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrPayslipNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("%w: payslip: %v", ErrDatabaseQueryFailed, err)
	}
	if payslip.OrgID != orgID {
		return nil, ErrPayslipNotFound
	}
	return payslip, nil
}

func (s *payslipService) getSelfPayslip(ctx context.Context, callerUserID uuid.UUID, id uuid.UUID) (*payslipRepository.Payslip, error) {
	actor, err := s.getActor(ctx, callerUserID)
	if err != nil {
		return nil, err
	}
	payslip, err := s.getPayslipInOrg(ctx, id, actor.OrgID)
	if err != nil {
		return nil, err
	}
	if payslip.EmployeeID != actor.EmployeeID {
		return nil, ErrPayslipNotFound
	}
	return payslip, nil
}

func (s *payslipService) getCycleInOrg(ctx context.Context, id uuid.UUID, orgID uuid.UUID) (*payslipRepository.PayrollCycle, error) {
	cycle, err := s.repo.GetPayrollCycle(ctx, id)
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

func (s *payslipService) getActor(ctx context.Context, callerUserID uuid.UUID) (*payslipRepository.ActorAccess, error) {
	actor, err := s.repo.GetActorAccess(ctx, callerUserID)
	if err != nil {
		return nil, fmt.Errorf("%w: actor: %v", ErrDatabaseQueryFailed, err)
	}
	return actor, nil
}

func (s *payslipService) requirePayslipManager(ctx context.Context, callerUserID uuid.UUID) (*payslipRepository.ActorAccess, error) {
	actor, err := s.getActor(ctx, callerUserID)
	if err != nil {
		return nil, err
	}
	if !canManagePayslips(actor.Role) {
		return nil, ErrForbidden
	}
	return actor, nil
}

func buildListFilter(orgID uuid.UUID, cycleID, employeeID, status, periodStart, periodEnd string) (payslipRepository.ListPayslipsFilter, error) {
	filter := payslipRepository.ListPayslipsFilter{OrgID: orgID, Status: status}
	if status != "" && status != payslipStatusGenerated && status != payslipStatusSent && status != payslipStatusVoid {
		return filter, ErrInvalidStatus
	}
	if cycleID != "" {
		id, err := uuid.Parse(cycleID)
		if err != nil {
			return filter, ErrInvalidCycleID
		}
		filter.CycleID = &id
	}
	if employeeID != "" {
		id, err := uuid.Parse(employeeID)
		if err != nil {
			return filter, ErrInvalidEmployeeID
		}
		filter.EmployeeID = &id
	}
	if periodStart != "" {
		date, err := parseDate(periodStart)
		if err != nil {
			return filter, ErrInvalidDateFormat
		}
		filter.PeriodStart = &date
	}
	if periodEnd != "" {
		date, err := parseDate(periodEnd)
		if err != nil {
			return filter, ErrInvalidDateFormat
		}
		filter.PeriodEnd = &date
	}
	return filter, nil
}

func buildPayload(
	payslipID uuid.UUID,
	org payslipRepository.OrganizationProfile,
	employee payslipRepository.EmployeeProfile,
	cycle payslipRepository.PayrollCycle,
	item payslipRepository.PayrollItem,
	adjustments []payslipRepository.PayrollAdjustment,
) ([]byte, error) {
	var calculationSnapshot any
	if len(item.CalculationSnapshot) > 0 {
		_ = json.Unmarshal(item.CalculationSnapshot, &calculationSnapshot)
	}
	var reviewReasons any
	if len(item.ReviewReasons) > 0 {
		_ = json.Unmarshal(item.ReviewReasons, &reviewReasons)
	}

	payload := map[string]any{
		"payslip_id": payslipID.String(),
		"organization": map[string]any{
			"id":      org.ID.String(),
			"name":    org.Name,
			"vat_id":  org.VATID,
			"address": org.Address,
		},
		"employee": map[string]any{
			"id":         employee.ID.String(),
			"user_id":    employee.UserID.String(),
			"full_name":  employee.FirstName + " " + employee.LastName,
			"first_name": employee.FirstName,
			"last_name":  employee.LastName,
			"email":      employee.Email,
			"role":       employee.Role,
			"department": employee.DepartmentName,
			"position":   employee.PositionName,
		},
		"period": map[string]any{
			"cycle_id": cycle.ID.String(),
			"start":    cycle.PeriodStart.Format(dateLayout),
			"end":      cycle.PeriodEnd.Format(dateLayout),
			"status":   cycle.Status,
			"currency": item.Currency,
		},
		"attendance": map[string]any{
			"working_days":     item.WorkingDays,
			"paid_days":        item.PaidDays,
			"unpaid_days":      item.UnpaidDays,
			"late_days":        item.LateDays,
			"absent_days":      item.AbsentDays,
			"missing_days":     item.MissingDays,
			"overtime_minutes": item.OvertimeMinutes,
		},
		"amounts": map[string]any{
			"base_salary":           item.BaseSalary,
			"attendance_adjustment": item.AttendanceAdjustment,
			"overtime_amount":       item.OvertimeAmount,
			"bonuses_total":         item.BonusesTotal,
			"deductions_total":      item.DeductionsTotal,
			"taxes_total":           item.TaxesTotal,
			"gross_salary":          item.GrossSalary,
			"net_salary":            item.NetSalary,
			"employer_taxes_total":  item.EmployerTaxesTotal,
			"total_employer_cost":   item.TotalEmployerCost,
		},
		"review": map[string]any{
			"review_required": item.ReviewRequired,
			"review_reasons":  reviewReasons,
		},
		"adjustments":          adjustmentsPayload(adjustments),
		"calculation_snapshot": calculationSnapshot,
		"generated_at":         time.Now().UTC().Format(time.RFC3339),
	}
	return json.Marshal(payload)
}

func adjustmentsPayload(adjustments []payslipRepository.PayrollAdjustment) []map[string]any {
	out := make([]map[string]any, 0, len(adjustments))
	for _, adjustment := range adjustments {
		out = append(out, map[string]any{
			"id":         adjustment.ID.String(),
			"type":       adjustment.Type,
			"category":   adjustment.Category,
			"amount":     adjustment.Amount,
			"currency":   adjustment.Currency,
			"is_taxable": adjustment.IsTaxable,
			"reason":     adjustment.Reason,
		})
	}
	return out
}

func renderPDFForPayslip(payslip payslipRepository.Payslip) ([]byte, error) {
	if payslip.Status == payslipStatusVoid {
		return nil, ErrPayslipLocked
	}
	bytes, err := renderPayslipPDF(payslip.PayloadSnapshot)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrRenderFailed, err)
	}
	return bytes, nil
}

func payslipFilename(id uuid.UUID) string {
	return "payslip-" + id.String() + ".pdf"
}

func contentSHA256(content []byte) string {
	sum := sha256.Sum256(content)
	return hex.EncodeToString(sum[:])
}

func canManagePayslips(role string) bool {
	switch normalizeRole(role) {
	case "sysadmin", "admin", "hr":
		return true
	default:
		return false
	}
}

func normalizeRole(role string) string {
	switch strings.ToLower(strings.TrimSpace(role)) {
	case "sysadmin", "superadmin", "super_admin":
		return "sysadmin"
	default:
		return strings.ToLower(strings.TrimSpace(role))
	}
}

func canGeneratePayslipForCycle(status string) bool {
	return status == cycleStatusReady || status == cycleStatusPaid
}

func parseDate(value string) (time.Time, error) {
	return time.Parse(dateLayout, value)
}
