package http

import (
	"errors"
	attendanceService "hrms/internal/feature/attendance/service"
	"hrms/internal/infrastructure/app/cognito"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AttendanceHandler struct {
	service    attendanceService.AttendanceService
	cognitoSvc *cognito.Service
}

func NewAttendanceHandler(service attendanceService.AttendanceService, cognitoSvc *cognito.Service) *AttendanceHandler {
	return &AttendanceHandler{service: service, cognitoSvc: cognitoSvc}
}

// ─── Work Schedule ────────────────────────────────────────────────────────────

// SetWorkSchedule godoc
// @Summary      Set work schedule
// @Description  Create or update the flexible work schedule for an employee
// @Tags         Attendance
// @Accept       json
// @Produce      json
// @Param        body  body      attendanceService.SetWorkScheduleRequest  true  "Schedule data"
// @Success      204
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      422  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /attendance/work-schedules [post]
func (h *AttendanceHandler) SetWorkSchedule(c *gin.Context) {
	callerUserID, ok := h.extractCallerUserID(c)
	if !ok {
		return
	}

	var req attendanceService.SetWorkScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if err := h.service.SetWorkSchedule(c.Request.Context(), callerUserID, req); err != nil {
		h.handleError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// GetWorkSchedule godoc
// @Summary      Get work schedule
// @Description  Retrieve the work schedule for a specific employee
// @Tags         Attendance
// @Produce      json
// @Param        employee_id  path      string  true  "Employee ID (UUID)"
// @Success      200  {object}  attendanceService.WorkScheduleResponse
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /attendance/work-schedules/{employee_id} [get]
func (h *AttendanceHandler) GetWorkSchedule(c *gin.Context) {
	employeeID, ok := h.parseUUIDParam(c, "employee_id")
	if !ok {
		return
	}

	resp, err := h.service.GetWorkSchedule(c.Request.Context(), employeeID)
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// ─── SKUD ─────────────────────────────────────────────────────────────────────

// CreateSkudEvent godoc
// @Summary      Register SKUD event
// @Description  Accept an entry/exit event from the access control system (СКУД). Automatically creates or updates the attendance record for the day.
// @Tags         Attendance
// @Accept       json
// @Produce      json
// @Param        body  body      attendanceService.CreateSkudEventRequest  true  "SKUD event"
// @Success      200  {object}  attendanceService.CreateSkudEventResponse
// @Failure      400  {object}  map[string]string
// @Failure      422  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /attendance/skud-events [post]
func (h *AttendanceHandler) CreateSkudEvent(c *gin.Context) {
	var req attendanceService.CreateSkudEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	resp, err := h.service.CreateSkudEvent(c.Request.Context(), req)
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// ─── Leave Requests ───────────────────────────────────────────────────────────

// CreateLeaveRequest godoc
// @Summary      Submit leave request
// @Description  Submit a leave request (sick leave, vacation, remote work, etc.)
// @Tags         Attendance
// @Accept       json
// @Produce      json
// @Param        body  body      attendanceService.CreateLeaveRequest  true  "Leave request data"
// @Success      201  {object}  attendanceService.CreateLeaveResponse
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      422  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /attendance/leave-requests [post]
func (h *AttendanceHandler) CreateLeaveRequest(c *gin.Context) {
	callerUserID, ok := h.extractCallerUserID(c)
	if !ok {
		return
	}

	var req attendanceService.CreateLeaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	resp, err := h.service.CreateLeaveRequest(c.Request.Context(), callerUserID, req)
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

// ListLeaveRequests godoc
// @Summary      List leave requests
// @Description  List all leave requests in the caller's organisation
// @Tags         Attendance
// @Produce      json
// @Success      200  {array}   attendanceService.LeaveRequestResponse
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /attendance/leave-requests [get]
func (h *AttendanceHandler) ListLeaveRequests(c *gin.Context) {
	callerUserID, ok := h.extractCallerUserID(c)
	if !ok {
		return
	}

	resp, err := h.service.ListLeaveRequestsByOrg(c.Request.Context(), callerUserID)
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// GetLeaveRequest godoc
// @Summary      Get leave request
// @Description  Retrieve details of a specific leave request
// @Tags         Attendance
// @Produce      json
// @Param        id  path      string  true  "Leave Request ID (UUID)"
// @Success      200  {object}  attendanceService.LeaveRequestResponse
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /attendance/leave-requests/{id} [get]
func (h *AttendanceHandler) GetLeaveRequest(c *gin.Context) {
	id, ok := h.parseUUIDParam(c, "id")
	if !ok {
		return
	}

	resp, err := h.service.GetLeaveRequestByID(c.Request.Context(), id)
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// ReviewLeaveRequest godoc
// @Summary      Approve or reject leave request
// @Description  HR or manager approves or rejects a leave request
// @Tags         Attendance
// @Accept       json
// @Param        id    path  string                              true  "Leave Request ID (UUID)"
// @Param        body  body  attendanceService.ReviewLeaveRequest  true  "Action (approve|reject)"
// @Success      204
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /attendance/leave-requests/{id}/review [patch]
func (h *AttendanceHandler) ReviewLeaveRequest(c *gin.Context) {
	callerUserID, ok := h.extractCallerUserID(c)
	if !ok {
		return
	}

	id, ok := h.parseUUIDParam(c, "id")
	if !ok {
		return
	}

	var req attendanceService.ReviewLeaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if err := h.service.ReviewLeaveRequest(c.Request.Context(), callerUserID, id, req); err != nil {
		h.handleError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ─── Attendance Records ───────────────────────────────────────────────────────

// ListAttendance godoc
// @Summary      List attendance records
// @Description  List attendance records for the caller's organisation filtered by period. Dates default to current month if not provided.
// @Tags         Attendance
// @Produce      json
// @Param        start_date  query     string  false  "Start date (YYYY-MM-DD)"
// @Param        end_date    query     string  false  "End date (YYYY-MM-DD)"
// @Success      200  {array}   attendanceService.AttendanceResponse
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /attendance [get]
func (h *AttendanceHandler) ListAttendance(c *gin.Context) {
	callerUserID, ok := h.extractCallerUserID(c)
	if !ok {
		return
	}

	resp, err := h.service.ListAttendanceByOrg(
		c.Request.Context(), callerUserID,
		c.Query("start_date"), c.Query("end_date"),
	)
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// ListEmployeeAttendance godoc
// @Summary      List employee attendance
// @Description  List attendance records for a specific employee filtered by period
// @Tags         Attendance
// @Produce      json
// @Param        employee_id  path      string  true   "Employee ID (UUID)"
// @Param        start_date   query     string  false  "Start date (YYYY-MM-DD)"
// @Param        end_date     query     string  false  "End date (YYYY-MM-DD)"
// @Success      200  {array}   attendanceService.AttendanceResponse
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /attendance/employees/{employee_id} [get]
func (h *AttendanceHandler) ListEmployeeAttendance(c *gin.Context) {
	employeeID, ok := h.parseUUIDParam(c, "employee_id")
	if !ok {
		return
	}

	resp, err := h.service.ListAttendanceByEmployee(
		c.Request.Context(), employeeID,
		c.Query("start_date"), c.Query("end_date"),
	)
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// ─── Private helpers ──────────────────────────────────────────────────────────

func (h *AttendanceHandler) extractCallerUserID(c *gin.Context) (uuid.UUID, bool) {
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
		log.Printf("[AttendanceHandler] ParseTokenClaims failed: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
		return uuid.Nil, false
	}
	return userID, true
}

func (h *AttendanceHandler) parseUUIDParam(c *gin.Context, param string) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param(param))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid " + param})
		return uuid.Nil, false
	}
	return id, true
}

func (h *AttendanceHandler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, attendanceService.ErrEmployeeNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "employee not found"})
	case errors.Is(err, attendanceService.ErrLeaveRequestNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "leave request not found"})
	case errors.Is(err, attendanceService.ErrInvalidEmployeeID):
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "invalid employee id"})
	case errors.Is(err, attendanceService.ErrInvalidLeaveRequestID):
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "invalid leave request id"})
	case errors.Is(err, attendanceService.ErrInvalidDateFormat):
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date format, use YYYY-MM-DD"})
	case errors.Is(err, attendanceService.ErrInvalidDateRange):
		c.JSON(http.StatusBadRequest, gin.H{"error": "end date must be after or equal to start date"})
	case errors.Is(err, attendanceService.ErrInvalidLeaveType):
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid leave type"})
	case errors.Is(err, attendanceService.ErrInvalidSkudEventType):
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event type, use ENTER or EXIT"})
	default:
		log.Printf("[AttendanceHandler] error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}
