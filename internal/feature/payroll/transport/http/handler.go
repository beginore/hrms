package http

import (
	"context"
	"errors"
	payrollService "hrms/internal/feature/payroll/service"
	"hrms/internal/infrastructure/middleware"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type PayrollHandler struct {
	service payrollService.PayrollService
}

func NewPayrollHandler(service payrollService.PayrollService) *PayrollHandler {
	return &PayrollHandler{service: service}
}

// CreateCycle godoc
// @Summary      Create payroll cycle
// @Description  Create a payroll cycle for an organisation period
// @Tags         Payroll
// @Accept       json
// @Produce      json
// @Param        body  body      payrollService.CreateCycleRequest  true  "Payroll cycle period"
// @Success      201   {object}  payrollService.PayrollCycleResponse
// @Failure      400   {object}  map[string]string
// @Failure      401   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Security     BearerAuth
// @Router       /payroll/cycles [post]
func (h *PayrollHandler) CreateCycle(c *gin.Context) {
	callerUserID, ok := h.extractCallerUserID(c)
	if !ok {
		return
	}
	var req payrollService.CreateCycleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	resp, err := h.service.CreateCycle(c.Request.Context(), callerUserID, req)
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

// ListCycles godoc
// @Summary      List payroll cycles
// @Description  List payroll cycles for the caller's organisation
// @Tags         Payroll
// @Produce      json
// @Success      200  {array}   payrollService.PayrollCycleResponse
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /payroll/cycles [get]
func (h *PayrollHandler) ListCycles(c *gin.Context) {
	callerUserID, ok := h.extractCallerUserID(c)
	if !ok {
		return
	}
	resp, err := h.service.ListCycles(c.Request.Context(), callerUserID)
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// GetCycle godoc
// @Summary      Get payroll cycle
// @Description  Retrieve one payroll cycle
// @Tags         Payroll
// @Produce      json
// @Param        id   path      string  true  "Payroll cycle ID"
// @Success      200  {object}  payrollService.PayrollCycleResponse
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /payroll/cycles/{id} [get]
func (h *PayrollHandler) GetCycle(c *gin.Context) {
	callerUserID, ok := h.extractCallerUserID(c)
	if !ok {
		return
	}
	id, ok := h.parseUUIDParam(c, "id")
	if !ok {
		return
	}
	resp, err := h.service.GetCycle(c.Request.Context(), callerUserID, id)
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// CalculateCycle godoc
// @Summary      Calculate payroll cycle
// @Description  Calculate payroll items from salary, attendance, overtime, bonuses, deductions, and taxes
// @Tags         Payroll
// @Accept       json
// @Produce      json
// @Param        id    path      string                                  true  "Payroll cycle ID"
// @Param        body  body      payrollService.CalculateCycleRequest    true  "Calculate options"
// @Success      200   {array}   payrollService.PayrollItemResponse
// @Failure      400   {object}  map[string]string
// @Failure      404   {object}  map[string]string
// @Failure      409   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Security     BearerAuth
// @Router       /payroll/cycles/{id}/calculate [patch]
func (h *PayrollHandler) CalculateCycle(c *gin.Context) {
	callerUserID, ok := h.extractCallerUserID(c)
	if !ok {
		return
	}
	id, ok := h.parseUUIDParam(c, "id")
	if !ok {
		return
	}
	var req payrollService.CalculateCycleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	resp, err := h.service.CalculateCycle(c.Request.Context(), callerUserID, id, req)
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// ApproveCycle godoc
// @Summary      Approve payroll cycle
// @Description  Lock a calculated payroll cycle for payment
// @Tags         Payroll
// @Param        id  path  string  true  "Payroll cycle ID"
// @Success      204
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      409  {object}  map[string]string
// @Security     BearerAuth
// @Router       /payroll/cycles/{id}/approve [patch]
func (h *PayrollHandler) ApproveCycle(c *gin.Context) {
	h.mutateCycle(c, h.service.ApproveCycle)
}

// PreparePayment godoc
// @Summary      Prepare payroll cycle payment batch
// @Description  Move an approved payroll cycle to READY_FOR_PAYMENT so it can be exported and marked as paid after external payment
// @Tags         Payroll
// @Param        id  path  string  true  "Payroll cycle ID"
// @Success      204
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      409  {object}  map[string]string
// @Security     BearerAuth
// @Router       /payroll/cycles/{id}/prepare-payment [patch]
func (h *PayrollHandler) PreparePayment(c *gin.Context) {
	h.mutateCycle(c, h.service.PreparePayment)
}

// MarkCyclePaid godoc
// @Summary      Mark payroll cycle as paid
// @Description  Mark a READY_FOR_PAYMENT cycle as paid after external payment is completed
// @Tags         Payroll
// @Param        id  path  string  true  "Payroll cycle ID"
// @Success      204
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      409  {object}  map[string]string
// @Security     BearerAuth
// @Router       /payroll/cycles/{id}/mark-paid [patch]
func (h *PayrollHandler) MarkCyclePaid(c *gin.Context) {
	h.mutateCycle(c, h.service.MarkCyclePaid)
}

// ReopenCycle godoc
// @Summary      Reopen payroll cycle
// @Description  Reopen an approved or READY_FOR_PAYMENT cycle back to calculated status. Paid cycles cannot be reopened.
// @Tags         Payroll
// @Param        id  path  string  true  "Payroll cycle ID"
// @Success      204
// @Failure      400  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      409  {object}  map[string]string
// @Security     BearerAuth
// @Router       /payroll/cycles/{id}/reopen [patch]
func (h *PayrollHandler) ReopenCycle(c *gin.Context) {
	h.mutateCycle(c, h.service.ReopenCycle)
}

// PreviewPayroll godoc
// @Summary      Preview payroll calculation
// @Description  Calculate one employee payroll without saving a payroll item
// @Tags         Payroll
// @Accept       json
// @Produce      json
// @Param        body  body      payrollService.PreviewPayrollRequest  true  "Preview data"
// @Success      200   {object}  payrollService.PayrollItemResponse
// @Failure      400   {object}  map[string]string
// @Failure      404   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Security     BearerAuth
// @Router       /payroll/preview [post]
func (h *PayrollHandler) PreviewPayroll(c *gin.Context) {
	callerUserID, ok := h.extractCallerUserID(c)
	if !ok {
		return
	}
	var req payrollService.PreviewPayrollRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	resp, err := h.service.PreviewPayroll(c.Request.Context(), callerUserID, req)
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// PreviewCycle godoc
// @Summary      Preview payroll cycle
// @Description  Calculate all employee payroll items for a cycle without saving them
// @Tags         Payroll
// @Produce      json
// @Param        id  path  string  true  "Payroll cycle ID"
// @Success      200  {array}   payrollService.PayrollItemResponse
// @Failure      400  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /payroll/cycles/{id}/preview [post]
func (h *PayrollHandler) PreviewCycle(c *gin.Context) {
	callerUserID, ok := h.extractCallerUserID(c)
	if !ok {
		return
	}
	id, ok := h.parseUUIDParam(c, "id")
	if !ok {
		return
	}
	resp, err := h.service.PreviewCycle(c.Request.Context(), callerUserID, id)
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// CreateAdjustment godoc
// @Summary      Create payroll adjustment
// @Description  Add a bonus or deduction to an employee payroll cycle
// @Tags         Payroll
// @Accept       json
// @Produce      json
// @Param        body  body      payrollService.CreateAdjustmentRequest  true  "Adjustment data"
// @Success      201   {object}  payrollService.PayrollAdjustmentResponse
// @Failure      400   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Security     BearerAuth
// @Router       /payroll/adjustments [post]
func (h *PayrollHandler) CreateAdjustment(c *gin.Context) {
	callerUserID, ok := h.extractCallerUserID(c)
	if !ok {
		return
	}
	var req payrollService.CreateAdjustmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	resp, err := h.service.CreateAdjustment(c.Request.Context(), callerUserID, req)
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

// ListAdjustments godoc
// @Summary      List payroll adjustments
// @Description  List bonuses and deductions, optionally filtered by cycle or employee
// @Tags         Payroll
// @Produce      json
// @Param        cycle_id     query  string  false  "Payroll cycle ID"
// @Param        employee_id  query  string  false  "Employee ID"
// @Success      200  {array}   payrollService.PayrollAdjustmentResponse
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /payroll/adjustments [get]
func (h *PayrollHandler) ListAdjustments(c *gin.Context) {
	callerUserID, ok := h.extractCallerUserID(c)
	if !ok {
		return
	}
	resp, err := h.service.ListAdjustments(c.Request.Context(), callerUserID, c.Query("cycle_id"), c.Query("employee_id"))
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// DeleteAdjustment godoc
// @Summary      Delete payroll adjustment
// @Description  Remove a payroll adjustment while cycle is still editable
// @Tags         Payroll
// @Param        id  path  string  true  "Adjustment ID"
// @Success      204
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /payroll/adjustments/{id} [delete]
func (h *PayrollHandler) DeleteAdjustment(c *gin.Context) {
	callerUserID, ok := h.extractCallerUserID(c)
	if !ok {
		return
	}
	id, ok := h.parseUUIDParam(c, "id")
	if !ok {
		return
	}
	if err := h.service.DeleteAdjustment(c.Request.Context(), callerUserID, id); err != nil {
		h.handleError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// CreateTaxRule godoc
// @Summary      Create payroll tax rule
// @Description  Create an organisation-specific tax rule
// @Tags         Payroll
// @Accept       json
// @Produce      json
// @Param        body  body      payrollService.CreateTaxRuleRequest  true  "Tax rule"
// @Success      201   {object}  payrollService.PayrollTaxRuleResponse
// @Failure      400   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Security     BearerAuth
// @Router       /payroll/tax-rules [post]
func (h *PayrollHandler) CreateTaxRule(c *gin.Context) {
	callerUserID, ok := h.extractCallerUserID(c)
	if !ok {
		return
	}
	var req payrollService.CreateTaxRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	resp, err := h.service.CreateTaxRule(c.Request.Context(), callerUserID, req)
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

// ListTaxRules godoc
// @Summary      List active payroll tax rules
// @Description  List default and organisation-specific active tax rules
// @Tags         Payroll
// @Produce      json
// @Success      200  {array}   payrollService.PayrollTaxRuleResponse
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /payroll/tax-rules [get]
func (h *PayrollHandler) ListTaxRules(c *gin.Context) {
	callerUserID, ok := h.extractCallerUserID(c)
	if !ok {
		return
	}
	resp, err := h.service.ListTaxRules(c.Request.Context(), callerUserID)
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// GetPolicy godoc
// @Summary      Get payroll policy
// @Description  Return payroll calculation policy for the caller's organisation
// @Tags         Payroll
// @Produce      json
// @Success      200  {object}  payrollService.PayrollPolicyResponse
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /payroll/policy [get]
func (h *PayrollHandler) GetPolicy(c *gin.Context) {
	callerUserID, ok := h.extractCallerUserID(c)
	if !ok {
		return
	}
	resp, err := h.service.GetPolicy(c.Request.Context(), callerUserID)
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// UpsertPolicy godoc
// @Summary      Upsert payroll policy
// @Description  Create or update payroll calculation policy for the caller's organisation
// @Tags         Payroll
// @Accept       json
// @Produce      json
// @Param        body  body      payrollService.UpsertPolicyRequest  true  "Payroll policy"
// @Success      200   {object}  payrollService.PayrollPolicyResponse
// @Failure      400   {object}  map[string]string
// @Failure      401   {object}  map[string]string
// @Failure      403   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Security     BearerAuth
// @Router       /payroll/policy [put]
func (h *PayrollHandler) UpsertPolicy(c *gin.Context) {
	callerUserID, ok := h.extractCallerUserID(c)
	if !ok {
		return
	}
	var req payrollService.UpsertPolicyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	resp, err := h.service.UpsertPolicy(c.Request.Context(), callerUserID, req)
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// CreateCorrection godoc
// @Summary      Create payroll correction
// @Description  Create a correction adjustment in an editable target cycle
// @Tags         Payroll
// @Accept       json
// @Produce      json
// @Param        body  body      payrollService.CreateCorrectionRequest  true  "Correction data"
// @Success      201   {object}  payrollService.PayrollCorrectionResponse
// @Failure      400   {object}  map[string]string
// @Failure      401   {object}  map[string]string
// @Failure      403   {object}  map[string]string
// @Failure      409   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Security     BearerAuth
// @Router       /payroll/corrections [post]
func (h *PayrollHandler) CreateCorrection(c *gin.Context) {
	callerUserID, ok := h.extractCallerUserID(c)
	if !ok {
		return
	}
	var req payrollService.CreateCorrectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	resp, err := h.service.CreateCorrection(c.Request.Context(), callerUserID, req)
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

// CreateSalaryHistory godoc
// @Summary      Create salary history record
// @Description  Add salary history for an employee and effective date range
// @Tags         Payroll
// @Accept       json
// @Produce      json
// @Param        body  body      payrollService.CreateSalaryHistoryRequest  true  "Salary history"
// @Success      201   {object}  payrollService.SalaryHistoryResponse
// @Failure      400   {object}  map[string]string
// @Failure      403   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Security     BearerAuth
// @Router       /payroll/salary-history [post]
func (h *PayrollHandler) CreateSalaryHistory(c *gin.Context) {
	callerUserID, ok := h.extractCallerUserID(c)
	if !ok {
		return
	}
	var req payrollService.CreateSalaryHistoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	resp, err := h.service.CreateSalaryHistory(c.Request.Context(), callerUserID, req)
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

// UpsertAttendanceOverride godoc
// @Summary      Upsert payroll attendance override
// @Description  Add a manual payroll attendance override for one employee/date
// @Tags         Payroll
// @Accept       json
// @Produce      json
// @Param        body  body      payrollService.UpsertAttendanceOverrideRequest  true  "Attendance override"
// @Success      200   {object}  payrollService.AttendanceOverrideResponse
// @Failure      400   {object}  map[string]string
// @Failure      403   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Security     BearerAuth
// @Router       /payroll/attendance-overrides [put]
func (h *PayrollHandler) UpsertAttendanceOverride(c *gin.Context) {
	callerUserID, ok := h.extractCallerUserID(c)
	if !ok {
		return
	}
	var req payrollService.UpsertAttendanceOverrideRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	resp, err := h.service.UpsertAttendanceOverride(c.Request.Context(), callerUserID, req)
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// UpdateEmploymentPeriod godoc
// @Summary      Update employee payroll employment period
// @Description  Set hire and termination dates used by payroll calculations
// @Tags         Payroll
// @Accept       json
// @Produce      json
// @Param        id    path      string                                        true  "Employee ID"
// @Param        body  body      payrollService.UpdateEmploymentPeriodRequest  true  "Employment period"
// @Success      200   {object}  payrollService.EmploymentPeriodResponse
// @Failure      400   {object}  map[string]string
// @Failure      403   {object}  map[string]string
// @Failure      404   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Security     BearerAuth
// @Router       /payroll/employees/{id}/employment-period [patch]
func (h *PayrollHandler) UpdateEmploymentPeriod(c *gin.Context) {
	callerUserID, ok := h.extractCallerUserID(c)
	if !ok {
		return
	}
	id, ok := h.parseUUIDParam(c, "id")
	if !ok {
		return
	}
	var req payrollService.UpdateEmploymentPeriodRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	resp, err := h.service.UpdateEmploymentPeriod(c.Request.Context(), callerUserID, id, req)
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// ReviewPayrollItem godoc
// @Summary      Review payroll item
// @Description  Mark a payroll item as reviewed and save reviewer comment
// @Tags         Payroll
// @Accept       json
// @Produce      json
// @Param        id    path  string                                  true  "Payroll item ID"
// @Param        body  body  payrollService.ReviewPayrollItemRequest  true  "Review comment"
// @Success      200   {object}  payrollService.PayrollItemResponse
// @Failure      400   {object}  map[string]string
// @Failure      403   {object}  map[string]string
// @Failure      404   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Security     BearerAuth
// @Router       /payroll/items/{id}/review [patch]
func (h *PayrollHandler) ReviewPayrollItem(c *gin.Context) {
	callerUserID, ok := h.extractCallerUserID(c)
	if !ok {
		return
	}
	id, ok := h.parseUUIDParam(c, "id")
	if !ok {
		return
	}
	var req payrollService.ReviewPayrollItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	resp, err := h.service.ReviewPayrollItem(c.Request.Context(), callerUserID, id, req)
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// ExportCycleCSV godoc
// @Summary      Export payroll cycle CSV
// @Description  Export READY_FOR_PAYMENT or PAID payroll items as CSV payment batch for accounting or bank upload
// @Tags         Payroll
// @Produce      text/csv
// @Param        id  path  string  true  "Payroll cycle ID"
// @Success      200  {file}  file
// @Failure      400  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /payroll/cycles/{id}/export [get]
func (h *PayrollHandler) ExportCycleCSV(c *gin.Context) {
	callerUserID, ok := h.extractCallerUserID(c)
	if !ok {
		return
	}
	id, ok := h.parseUUIDParam(c, "id")
	if !ok {
		return
	}
	data, err := h.service.ExportCycleCSV(c.Request.Context(), callerUserID, id)
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.Header("Content-Disposition", "attachment; filename=payroll.csv")
	c.Data(http.StatusOK, "text/csv; charset=utf-8", data)
}

// CreateKZTaxPreset godoc
// @Summary      Create KZ payroll tax preset
// @Description  Add basic Kazakhstan employee and employer tax rules for the organisation
// @Tags         Payroll
// @Produce      json
// @Success      201  {array}   payrollService.PayrollTaxRuleResponse
// @Failure      403  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /payroll/tax-rules/kz-preset [post]
func (h *PayrollHandler) CreateKZTaxPreset(c *gin.Context) {
	callerUserID, ok := h.extractCallerUserID(c)
	if !ok {
		return
	}
	resp, err := h.service.CreateKZTaxPreset(c.Request.Context(), callerUserID)
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

// ListPayrollItems godoc
// @Summary      List payroll items
// @Description  List calculated payroll items for a cycle
// @Tags         Payroll
// @Produce      json
// @Param        id  path  string  true  "Payroll cycle ID"
// @Success      200  {array}   payrollService.PayrollItemResponse
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /payroll/cycles/{id}/items [get]
func (h *PayrollHandler) ListPayrollItems(c *gin.Context) {
	callerUserID, ok := h.extractCallerUserID(c)
	if !ok {
		return
	}
	cycleID, ok := h.parseUUIDParam(c, "id")
	if !ok {
		return
	}
	resp, err := h.service.ListPayrollItems(c.Request.Context(), callerUserID, cycleID)
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// GetPayrollItem godoc
// @Summary      Get payroll item
// @Description  Retrieve one calculated payroll item
// @Tags         Payroll
// @Produce      json
// @Param        id  path  string  true  "Payroll item ID"
// @Success      200  {object}  payrollService.PayrollItemResponse
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /payroll/items/{id} [get]
func (h *PayrollHandler) GetPayrollItem(c *gin.Context) {
	callerUserID, ok := h.extractCallerUserID(c)
	if !ok {
		return
	}
	id, ok := h.parseUUIDParam(c, "id")
	if !ok {
		return
	}
	resp, err := h.service.GetPayrollItem(c.Request.Context(), callerUserID, id)
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *PayrollHandler) mutateCycle(c *gin.Context, fn func(ctx context.Context, callerUserID uuid.UUID, id uuid.UUID) error) {
	callerUserID, ok := h.extractCallerUserID(c)
	if !ok {
		return
	}
	id, ok := h.parseUUIDParam(c, "id")
	if !ok {
		return
	}
	if err := fn(c.Request.Context(), callerUserID, id); err != nil {
		h.handleError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *PayrollHandler) extractCallerUserID(c *gin.Context) (uuid.UUID, bool) {
	value, exists := c.Get(middleware.UserIDKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return uuid.Nil, false
	}
	userID, ok := value.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return uuid.Nil, false
	}
	return userID, true
}

func (h *PayrollHandler) parseUUIDParam(c *gin.Context, param string) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param(param))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid " + param})
		return uuid.Nil, false
	}
	return id, true
}

func (h *PayrollHandler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, payrollService.ErrInvalidDateFormat):
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date format, use YYYY-MM-DD"})
	case errors.Is(err, payrollService.ErrInvalidDateRange):
		c.JSON(http.StatusBadRequest, gin.H{"error": "end date must be after or equal to start date"})
	case errors.Is(err, payrollService.ErrInvalidCycleID):
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid cycle id"})
	case errors.Is(err, payrollService.ErrInvalidEmployeeID):
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid employee id"})
	case errors.Is(err, payrollService.ErrInvalidAmount):
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid amount"})
	case errors.Is(err, payrollService.ErrInvalidRate):
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tax rate"})
	case errors.Is(err, payrollService.ErrInvalidAdjustmentType):
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid adjustment type, use BONUS or DEDUCTION"})
	case errors.Is(err, payrollService.ErrInvalidTaxAppliesTo):
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid applies_to, use GROSS, TAXABLE_INCOME, or BASE_ONLY"})
	case errors.Is(err, payrollService.ErrInvalidTaxPayer):
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payer, use EMPLOYEE or EMPLOYER"})
	case errors.Is(err, payrollService.ErrInvalidPolicy):
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payroll policy"})
	case errors.Is(err, payrollService.ErrForbidden):
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
	case errors.Is(err, payrollService.ErrCycleNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "payroll cycle not found"})
	case errors.Is(err, payrollService.ErrPayrollItemNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "payroll item not found"})
	case errors.Is(err, payrollService.ErrEmployeeNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "employee not found"})
	case errors.Is(err, payrollService.ErrCycleLocked):
		c.JSON(http.StatusConflict, gin.H{"error": "payroll cycle is locked"})
	case errors.Is(err, payrollService.ErrNoWorkingDays):
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "payroll period has no working days"})
	case errors.Is(err, payrollService.ErrInvalidSalaryRate):
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "invalid salary rate"})
	default:
		log.Printf("[PayrollHandler] error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}
