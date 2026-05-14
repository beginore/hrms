package http

import (
	"context"
	"errors"
	payslipService "hrms/internal/feature/payslip/service"
	"hrms/internal/infrastructure/app/cognito"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type PayslipHandler struct {
	service    payslipService.PayslipService
	cognitoSvc *cognito.Service
}

func NewPayslipHandler(service payslipService.PayslipService, cognitoSvc *cognito.Service) *PayslipHandler {
	return &PayslipHandler{service: service, cognitoSvc: cognitoSvc}
}

// GeneratePayslip godoc
// @Summary      Generate payslip
// @Description  Generate a payslip from a READY_FOR_PAYMENT or PAID payroll item
// @Tags         Payslips
// @Accept       json
// @Produce      json
// @Param        body  body      payslipService.GeneratePayslipRequest  true  "Payroll item"
// @Success      201   {object}  payslipService.PayslipResponse
// @Failure      400   {object}  map[string]string
// @Failure      403   {object}  map[string]string
// @Failure      404   {object}  map[string]string
// @Failure      409   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Security     BearerAuth
// @Router       /payslips [post]
func (h *PayslipHandler) GeneratePayslip(c *gin.Context) {
	callerUserID, ok := h.extractCallerUserID(c)
	if !ok {
		return
	}
	var req payslipService.GeneratePayslipRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	resp, err := h.service.GeneratePayslip(c.Request.Context(), callerUserID, req)
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

// GenerateCyclePayslips godoc
// @Summary      Generate cycle payslips
// @Description  Generate payslips for all payroll items in a READY_FOR_PAYMENT or PAID cycle
// @Tags         Payslips
// @Produce      json
// @Param        id   path      string  true  "Payroll cycle ID"
// @Success      201  {object}  payslipService.PayslipBatchResponse
// @Failure      400  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      409  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /payroll/cycles/{id}/payslips/generate [post]
func (h *PayslipHandler) GenerateCyclePayslips(c *gin.Context) {
	callerUserID, ok := h.extractCallerUserID(c)
	if !ok {
		return
	}
	id, ok := h.parseUUIDParam(c, "id")
	if !ok {
		return
	}
	resp, err := h.service.GenerateCyclePayslips(c.Request.Context(), callerUserID, id)
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

// ListPayslips godoc
// @Summary      List payslips
// @Description  List payslips for the caller's organization with optional filters
// @Tags         Payslips
// @Produce      json
// @Param        cycle_id      query  string  false  "Payroll cycle ID"
// @Param        employee_id   query  string  false  "Employee ID"
// @Param        status        query  string  false  "Payslip status"
// @Param        period_start  query  string  false  "Period start YYYY-MM-DD"
// @Param        period_end    query  string  false  "Period end YYYY-MM-DD"
// @Success      200           {array}   payslipService.PayslipResponse
// @Failure      400           {object}  map[string]string
// @Failure      403           {object}  map[string]string
// @Failure      500           {object}  map[string]string
// @Security     BearerAuth
// @Router       /payslips [get]
func (h *PayslipHandler) ListPayslips(c *gin.Context) {
	callerUserID, ok := h.extractCallerUserID(c)
	if !ok {
		return
	}
	resp, err := h.service.ListPayslips(
		c.Request.Context(),
		callerUserID,
		c.Query("cycle_id"),
		c.Query("employee_id"),
		c.Query("status"),
		c.Query("period_start"),
		c.Query("period_end"),
	)
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// GetPayslip godoc
// @Summary      Get payslip
// @Description  Retrieve one payslip by ID
// @Tags         Payslips
// @Produce      json
// @Param        id   path      string  true  "Payslip ID"
// @Success      200  {object}  payslipService.PayslipResponse
// @Failure      400  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /payslips/{id} [get]
func (h *PayslipHandler) GetPayslip(c *gin.Context) {
	h.getPayslip(c, h.service.GetPayslip)
}

// ExportPayslipPDF godoc
// @Summary      Export payslip PDF
// @Description  Export payslip as PDF
// @Tags         Payslips
// @Produce      application/pdf
// @Param        id   path  string  true  "Payslip ID"
// @Success      200  {file}  file
// @Failure      400  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      409  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /payslips/{id}/pdf [get]
func (h *PayslipHandler) ExportPayslipPDF(c *gin.Context) {
	h.renderPDF(c, h.service.RenderPayslipPDF)
}

// SendPayslip godoc
// @Summary      Send payslip
// @Description  Send a payslip notification to the employee email and mark it as SENT
// @Tags         Payslips
// @Produce      json
// @Param        id   path      string  true  "Payslip ID"
// @Success      200  {object}  payslipService.PayslipSendResponse
// @Failure      400  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      409  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /payslips/{id}/send [post]
func (h *PayslipHandler) SendPayslip(c *gin.Context) {
	callerUserID, ok := h.extractCallerUserID(c)
	if !ok {
		return
	}
	id, ok := h.parseUUIDParam(c, "id")
	if !ok {
		return
	}
	resp, err := h.service.SendPayslip(c.Request.Context(), callerUserID, id)
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// SendCyclePayslips godoc
// @Summary      Send cycle payslips
// @Description  Send all generated payslips for a cycle
// @Tags         Payslips
// @Produce      json
// @Param        id   path      string  true  "Payroll cycle ID"
// @Success      200  {object}  payslipService.PayslipBatchSendResponse
// @Failure      400  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /payroll/cycles/{id}/payslips/send [post]
func (h *PayslipHandler) SendCyclePayslips(c *gin.Context) {
	callerUserID, ok := h.extractCallerUserID(c)
	if !ok {
		return
	}
	id, ok := h.parseUUIDParam(c, "id")
	if !ok {
		return
	}
	resp, err := h.service.SendCyclePayslips(c.Request.Context(), callerUserID, id)
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// VoidPayslip godoc
// @Summary      Void payslip
// @Description  Mark a generated or sent payslip as VOID without deleting it
// @Tags         Payslips
// @Accept       json
// @Produce      json
// @Param        id    path      string                             true  "Payslip ID"
// @Param        body  body      payslipService.VoidPayslipRequest  true  "Void reason"
// @Success      200   {object}  payslipService.PayslipResponse
// @Failure      400   {object}  map[string]string
// @Failure      403   {object}  map[string]string
// @Failure      404   {object}  map[string]string
// @Failure      409   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Security     BearerAuth
// @Router       /payslips/{id}/void [patch]
func (h *PayslipHandler) VoidPayslip(c *gin.Context) {
	callerUserID, ok := h.extractCallerUserID(c)
	if !ok {
		return
	}
	id, ok := h.parseUUIDParam(c, "id")
	if !ok {
		return
	}
	var req payslipService.VoidPayslipRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	resp, err := h.service.VoidPayslip(c.Request.Context(), callerUserID, id, req)
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// RegeneratePayslip godoc
// @Summary      Regenerate payslip
// @Description  Create a new payslip from the same payroll item after the old payslip was voided
// @Tags         Payslips
// @Produce      json
// @Param        id   path      string  true  "Voided payslip ID"
// @Success      201  {object}  payslipService.PayslipResponse
// @Failure      400  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      409  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /payslips/{id}/regenerate [post]
func (h *PayslipHandler) RegeneratePayslip(c *gin.Context) {
	callerUserID, ok := h.extractCallerUserID(c)
	if !ok {
		return
	}
	id, ok := h.parseUUIDParam(c, "id")
	if !ok {
		return
	}
	resp, err := h.service.RegeneratePayslip(c.Request.Context(), callerUserID, id)
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

// ListMyPayslips godoc
// @Summary      List my payslips
// @Description  List payslips for the authenticated employee
// @Tags         Payslips
// @Produce      json
// @Success      200  {array}   payslipService.PayslipResponse
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /me/payslips [get]
func (h *PayslipHandler) ListMyPayslips(c *gin.Context) {
	callerUserID, ok := h.extractCallerUserID(c)
	if !ok {
		return
	}
	resp, err := h.service.ListMyPayslips(c.Request.Context(), callerUserID)
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// GetMyPayslip godoc
// @Summary      Get my payslip
// @Description  Retrieve one payslip that belongs to the authenticated employee
// @Tags         Payslips
// @Produce      json
// @Param        id   path      string  true  "Payslip ID"
// @Success      200  {object}  payslipService.PayslipResponse
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /me/payslips/{id} [get]
func (h *PayslipHandler) GetMyPayslip(c *gin.Context) {
	h.getPayslip(c, h.service.GetMyPayslip)
}

// ExportMyPayslipPDF godoc
// @Summary      Export my payslip PDF
// @Description  Export authenticated employee payslip as PDF
// @Tags         Payslips
// @Produce      application/pdf
// @Param        id   path  string  true  "Payslip ID"
// @Success      200  {file}  file
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      409  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /me/payslips/{id}/pdf [get]
func (h *PayslipHandler) ExportMyPayslipPDF(c *gin.Context) {
	h.renderPDF(c, h.service.RenderMyPayslipPDF)
}

func (h *PayslipHandler) getPayslip(c *gin.Context, fn func(context.Context, uuid.UUID, uuid.UUID) (*payslipService.PayslipResponse, error)) {
	callerUserID, ok := h.extractCallerUserID(c)
	if !ok {
		return
	}
	id, ok := h.parseUUIDParam(c, "id")
	if !ok {
		return
	}
	resp, err := fn(c.Request.Context(), callerUserID, id)
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *PayslipHandler) renderPDF(c *gin.Context, fn func(context.Context, uuid.UUID, uuid.UUID) ([]byte, error)) {
	callerUserID, ok := h.extractCallerUserID(c)
	if !ok {
		return
	}
	id, ok := h.parseUUIDParam(c, "id")
	if !ok {
		return
	}
	data, err := fn(c.Request.Context(), callerUserID, id)
	if err != nil {
		h.handleError(c, err)
		return
	}
	filename := "payslip-" + id.String() + ".pdf"
	c.Header("Content-Disposition", `attachment; filename="`+filename+`"`)
	c.Header("Access-Control-Expose-Headers", "Content-Disposition, Content-Length, Content-Type")
	c.Header("Cache-Control", "no-store")
	c.Header("X-Content-Type-Options", "nosniff")
	c.Data(http.StatusOK, "application/pdf", data)
}

func (h *PayslipHandler) extractCallerUserID(c *gin.Context) (uuid.UUID, bool) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
		return uuid.Nil, false
	}
	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token == authHeader {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
		return uuid.Nil, false
	}
	userID, _, err := h.cognitoSvc.ParseTokenClaims(token)
	if err != nil {
		log.Printf("[PayslipHandler] ParseTokenClaims failed: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
		return uuid.Nil, false
	}
	return userID, true
}

func (h *PayslipHandler) parseUUIDParam(c *gin.Context, param string) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param(param))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid " + param})
		return uuid.Nil, false
	}
	return id, true
}

func (h *PayslipHandler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, payslipService.ErrInvalidPayslipID):
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payslip id"})
	case errors.Is(err, payslipService.ErrInvalidPayrollItemID):
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payroll item id"})
	case errors.Is(err, payslipService.ErrInvalidCycleID):
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid cycle id"})
	case errors.Is(err, payslipService.ErrInvalidEmployeeID):
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid employee id"})
	case errors.Is(err, payslipService.ErrInvalidDateFormat):
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date format, use YYYY-MM-DD"})
	case errors.Is(err, payslipService.ErrInvalidStatus):
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payslip status"})
	case errors.Is(err, payslipService.ErrForbidden):
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
	case errors.Is(err, payslipService.ErrPayslipNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "payslip not found"})
	case errors.Is(err, payslipService.ErrPayrollItemNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "payroll item not found"})
	case errors.Is(err, payslipService.ErrCycleNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "payroll cycle not found"})
	case errors.Is(err, payslipService.ErrInvalidCycleStatus):
		c.JSON(http.StatusConflict, gin.H{"error": "payroll cycle must be READY_FOR_PAYMENT or PAID"})
	case errors.Is(err, payslipService.ErrReviewRequired):
		c.JSON(http.StatusConflict, gin.H{"error": "payroll item requires review before payslip generation"})
	case errors.Is(err, payslipService.ErrPayslipLocked):
		c.JSON(http.StatusConflict, gin.H{"error": "payslip is locked"})
	case errors.Is(err, payslipService.ErrSendFailed):
		log.Printf("[PayslipHandler] send error: %v", err)
		c.JSON(http.StatusBadGateway, gin.H{"error": "failed to send payslip"})
	default:
		log.Printf("[PayslipHandler] error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}
