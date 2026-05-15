package http

import (
	"errors"
	attendanceService "hrms/internal/feature/attendance/service"
	"hrms/internal/infrastructure/middleware"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AttendanceHandler struct {
	service attendanceService.AttendanceService
}

func NewAttendanceHandler(service attendanceService.AttendanceService) *AttendanceHandler {
	return &AttendanceHandler{service: service}
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
	var req attendanceService.SetWorkScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if err := h.service.SetWorkSchedule(c.Request.Context(), h.callerUserID(c), req); err != nil {
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

// CheckIn godoc
// @Summary      Manual check-in
// @Description  Employee manually checks in for the day. Use work_type=REMOTE for remote work, OFFICE for office.
// @Tags         Attendance
// @Accept       json
// @Produce      json
// @Param        body  body      attendanceService.CheckInRequest  false  "Check-in options"
// @Success      200   {object}  attendanceService.CheckInResponse
// @Failure      400   {object}  map[string]string
// @Failure      401   {object}  map[string]string
// @Failure      409   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Security     BearerAuth
// @Router       /attendance/check-in [post]
func (h *AttendanceHandler) CheckIn(c *gin.Context) {
	var req attendanceService.CheckInRequest
	_ = c.ShouldBindJSON(&req)
	resp, err := h.service.CheckIn(c.Request.Context(), h.callerUserID(c), req)
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// CheckOut godoc
// @Summary      Manual check-out
// @Description  Employee manually checks out for the day. Requires a prior check-in.
// @Tags         Attendance
// @Produce      json
// @Success      200  {object}  attendanceService.CheckOutResponse
// @Failure      401  {object}  map[string]string
// @Failure      422  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /attendance/check-out [post]
func (h *AttendanceHandler) CheckOut(c *gin.Context) {
	resp, err := h.service.CheckOut(c.Request.Context(), h.callerUserID(c))
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

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
	var req attendanceService.CreateLeaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	resp, err := h.service.CreateLeaveRequest(c.Request.Context(), h.callerUserID(c), req)
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
	resp, err := h.service.ListLeaveRequests(c.Request.Context(), h.callerUserID(c), h.callerRole(c))
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
	id, ok := h.parseUUIDParam(c, "id")
	if !ok {
		return
	}

	var req attendanceService.ReviewLeaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if err := h.service.ReviewLeaveRequest(c.Request.Context(), h.callerUserID(c), id, req); err != nil {
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
	resp, err := h.service.ListAttendance(
		c.Request.Context(), h.callerUserID(c), h.callerRole(c),
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
		c.Request.Context(), h.callerUserID(c), h.callerRole(c), employeeID,
		c.Query("start_date"), c.Query("end_date"),
	)
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// ─── Private helpers ──────────────────────────────────────────────────────────

func (h *AttendanceHandler) callerUserID(c *gin.Context) uuid.UUID {
	if value, exists := c.Get(middleware.UserIDKey); exists {
		if id, ok := value.(uuid.UUID); ok {
			return id
		}
	}
	return uuid.Nil
}

func (h *AttendanceHandler) callerRole(c *gin.Context) string {
	if value, exists := c.Get(middleware.RoleKey); exists {
		if role, ok := value.(string); ok {
			return role
		}
	}
	return ""
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
	case errors.Is(err, attendanceService.ErrLeaveAlreadyReviewed):
		c.JSON(http.StatusConflict, gin.H{"error": "leave request has already been reviewed"})
	case errors.Is(err, attendanceService.ErrAlreadyCheckedIn):
		c.JSON(http.StatusConflict, gin.H{"error": "already checked in today"})
	case errors.Is(err, attendanceService.ErrAlreadyCheckedOut):
		c.JSON(http.StatusConflict, gin.H{"error": "already checked out today"})
	case errors.Is(err, attendanceService.ErrNotCheckedIn):
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "no check-in found for today"})
	case errors.Is(err, attendanceService.ErrForbidden):
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
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
	case errors.Is(err, attendanceService.ErrInvalidWorkType):
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid work type, use OFFICE or REMOTE"})
	case errors.Is(err, attendanceService.ErrInvalidSkudEventType):
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event type, use ENTER or EXIT"})
	default:
		log.Printf("[AttendanceHandler] error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}
