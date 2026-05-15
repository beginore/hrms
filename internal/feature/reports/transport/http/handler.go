package http

import (
	"errors"
	"log"
	"net/http"

	reportsService "hrms/internal/feature/reports/service"
	"hrms/internal/infrastructure/middleware"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ReportsHandler struct {
	service reportsService.ReportsService
}

func NewReportsHandler(service reportsService.ReportsService) *ReportsHandler {
	return &ReportsHandler{service: service}
}

// GetDashboard godoc
// @Summary      Get dashboard analytics
// @Description  Return dashboard cards for employees, attendance, and monthly payroll
// @Tags         Reports
// @Produce      json
// @Success      200  {object}  reportsService.DashboardResponse
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /reports/dashboard [get]
func (h *ReportsHandler) GetDashboard(c *gin.Context) {
	callerUserID, ok := currentUserID(c)
	if !ok {
		return
	}
	resp, err := h.service.GetDashboard(c.Request.Context(), callerUserID)
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// GetPayrollSummary godoc
// @Summary      Get payroll report summary
// @Description  Return payroll totals for approved, ready, and paid payroll cycles
// @Tags         Reports
// @Produce      json
// @Param        period_start  query     string  false  "Period start YYYY-MM-DD"
// @Param        period_end    query     string  false  "Period end YYYY-MM-DD"
// @Success      200           {object}  reportsService.PayrollSummaryResponse
// @Failure      400           {object}  map[string]string
// @Failure      401           {object}  map[string]string
// @Failure      403           {object}  map[string]string
// @Failure      500           {object}  map[string]string
// @Security     BearerAuth
// @Router       /reports/payroll [get]
func (h *ReportsHandler) GetPayrollSummary(c *gin.Context) {
	callerUserID, ok := currentUserID(c)
	if !ok {
		return
	}
	resp, err := h.service.GetPayrollSummary(c.Request.Context(), callerUserID, c.Query("period_start"), c.Query("period_end"))
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// GetPayrollTrends godoc
// @Summary      Get payroll trend chart
// @Description  Return monthly payroll totals as chart-ready labels and series
// @Tags         Reports
// @Produce      json
// @Param        period_start  query     string  false  "Period start YYYY-MM-DD"
// @Param        period_end    query     string  false  "Period end YYYY-MM-DD"
// @Success      200           {object}  reportsService.LineChartResponse
// @Failure      400           {object}  map[string]string
// @Failure      401           {object}  map[string]string
// @Failure      403           {object}  map[string]string
// @Failure      500           {object}  map[string]string
// @Security     BearerAuth
// @Router       /reports/payroll/trends [get]
func (h *ReportsHandler) GetPayrollTrends(c *gin.Context) {
	callerUserID, ok := currentUserID(c)
	if !ok {
		return
	}
	resp, err := h.service.GetPayrollTrends(c.Request.Context(), callerUserID, c.Query("period_start"), c.Query("period_end"))
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// GetDepartmentPayroll godoc
// @Summary      Get department payroll report
// @Description  Return payroll breakdown by department with chart-ready data
// @Tags         Reports
// @Produce      json
// @Param        period_start  query     string  false  "Period start YYYY-MM-DD"
// @Param        period_end    query     string  false  "Period end YYYY-MM-DD"
// @Success      200           {object}  reportsService.DepartmentPayrollResponse
// @Failure      400           {object}  map[string]string
// @Failure      401           {object}  map[string]string
// @Failure      403           {object}  map[string]string
// @Failure      500           {object}  map[string]string
// @Security     BearerAuth
// @Router       /reports/payroll/departments [get]
func (h *ReportsHandler) GetDepartmentPayroll(c *gin.Context) {
	callerUserID, ok := currentUserID(c)
	if !ok {
		return
	}
	resp, err := h.service.GetDepartmentPayroll(c.Request.Context(), callerUserID, c.Query("period_start"), c.Query("period_end"))
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// GetAttendanceToday godoc
// @Summary      Get attendance breakdown for a day
// @Description  Return present, late, leave, and absent counts as chart-ready data
// @Tags         Reports
// @Produce      json
// @Param        date  query     string  false  "Date YYYY-MM-DD"
// @Success      200   {object}  reportsService.AttendanceTodayResponse
// @Failure      400   {object}  map[string]string
// @Failure      401   {object}  map[string]string
// @Failure      403   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Security     BearerAuth
// @Router       /reports/attendance/today [get]
func (h *ReportsHandler) GetAttendanceToday(c *gin.Context) {
	callerUserID, ok := currentUserID(c)
	if !ok {
		return
	}
	resp, err := h.service.GetAttendanceToday(c.Request.Context(), callerUserID, c.Query("date"))
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// GetAttendanceWeekly godoc
// @Summary      Get weekly attendance chart
// @Description  Return attendance counts for a 7-day period
// @Tags         Reports
// @Produce      json
// @Param        week_start  query     string  false  "Week start YYYY-MM-DD"
// @Success      200         {object}  reportsService.LineChartResponse
// @Failure      400         {object}  map[string]string
// @Failure      401         {object}  map[string]string
// @Failure      403         {object}  map[string]string
// @Failure      500         {object}  map[string]string
// @Security     BearerAuth
// @Router       /reports/attendance/weekly [get]
func (h *ReportsHandler) GetAttendanceWeekly(c *gin.Context) {
	callerUserID, ok := currentUserID(c)
	if !ok {
		return
	}
	resp, err := h.service.GetAttendanceWeekly(c.Request.Context(), callerUserID, c.Query("week_start"))
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// GetEmployeeStatistics godoc
// @Summary      Get employee statistics
// @Description  Return employee distribution by department, role, and status
// @Tags         Reports
// @Produce      json
// @Success      200  {object}  reportsService.EmployeeStatisticsResponse
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /reports/employees/statistics [get]
func (h *ReportsHandler) GetEmployeeStatistics(c *gin.Context) {
	callerUserID, ok := currentUserID(c)
	if !ok {
		return
	}
	resp, err := h.service.GetEmployeeStatistics(c.Request.Context(), callerUserID)
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// GetDepartmentStatistics godoc
// @Summary      Get department statistics
// @Description  Return department employee, payroll, and attendance statistics
// @Tags         Reports
// @Produce      json
// @Param        period_start  query     string  false  "Period start YYYY-MM-DD"
// @Param        period_end    query     string  false  "Period end YYYY-MM-DD"
// @Success      200           {object}  reportsService.DepartmentStatisticsResponse
// @Failure      400           {object}  map[string]string
// @Failure      401           {object}  map[string]string
// @Failure      403           {object}  map[string]string
// @Failure      500           {object}  map[string]string
// @Security     BearerAuth
// @Router       /reports/departments/statistics [get]
func (h *ReportsHandler) GetDepartmentStatistics(c *gin.Context) {
	callerUserID, ok := currentUserID(c)
	if !ok {
		return
	}
	resp, err := h.service.GetDepartmentStatistics(c.Request.Context(), callerUserID, c.Query("period_start"), c.Query("period_end"))
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// ExportReportCSV godoc
// @Summary      Export report as CSV
// @Description  Export payroll or attendance report as CSV
// @Tags         Reports
// @Produce      text/csv
// @Param        type          query     string  true   "Report type: payroll or attendance"
// @Param        period_start  query     string  false  "Period start YYYY-MM-DD"
// @Param        period_end    query     string  false  "Period end YYYY-MM-DD"
// @Success      200           {file}    file
// @Failure      400           {object}  map[string]string
// @Failure      401           {object}  map[string]string
// @Failure      403           {object}  map[string]string
// @Failure      500           {object}  map[string]string
// @Security     BearerAuth
// @Router       /reports/export/csv [get]
func (h *ReportsHandler) ExportReportCSV(c *gin.Context) {
	h.exportReport(c, "csv")
}

// ExportReportPDF godoc
// @Summary      Export report as PDF
// @Description  Export payroll or attendance report as a downloadable PDF file
// @Tags         Reports
// @Produce      application/pdf
// @Param        type          query     string  true   "Report type: payroll or attendance"
// @Param        period_start  query     string  false  "Period start YYYY-MM-DD"
// @Param        period_end    query     string  false  "Period end YYYY-MM-DD"
// @Success      200           {file}    file
// @Failure      400           {object}  map[string]string
// @Failure      401           {object}  map[string]string
// @Failure      403           {object}  map[string]string
// @Failure      500           {object}  map[string]string
// @Security     BearerAuth
// @Router       /reports/export/pdf [get]
func (h *ReportsHandler) ExportReportPDF(c *gin.Context) {
	h.exportReport(c, "pdf")
}

func (h *ReportsHandler) exportReport(c *gin.Context, format string) {
	callerUserID, ok := currentUserID(c)
	if !ok {
		return
	}
	resp, err := h.service.ExportReport(c.Request.Context(), callerUserID, c.Query("type"), format, c.Query("period_start"), c.Query("period_end"))
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.Header("Content-Disposition", `attachment; filename="`+resp.Filename+`"`)
	c.Header("Access-Control-Expose-Headers", "Content-Disposition, Content-Length, Content-Type")
	c.Header("X-Content-Type-Options", "nosniff")
	c.Data(http.StatusOK, resp.ContentType, resp.Content)
}

func currentUserID(c *gin.Context) (uuid.UUID, bool) {
	value, exists := c.Get(middleware.UserIDKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authenticated user"})
		return uuid.Nil, false
	}
	userID, ok := value.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authenticated user"})
		return uuid.Nil, false
	}
	return userID, true
}

func (h *ReportsHandler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, reportsService.ErrInvalidDateFormat):
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date format, use YYYY-MM-DD"})
	case errors.Is(err, reportsService.ErrInvalidDateRange):
		c.JSON(http.StatusBadRequest, gin.H{"error": "period_end must be greater than or equal to period_start"})
	case errors.Is(err, reportsService.ErrInvalidReportType):
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid report type, use payroll or attendance"})
	case errors.Is(err, reportsService.ErrInvalidExportFormat):
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid export format, use csv or pdf"})
	case errors.Is(err, reportsService.ErrForbidden):
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
	default:
		log.Printf("[ReportsHandler] error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}
