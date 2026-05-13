package http

import (
	"errors"
	calendarService "hrms/internal/feature/calendar/service"
	"hrms/internal/infrastructure/app/cognito"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CalendarHandler struct {
	service    calendarService.CalendarService
	cognitoSvc *cognito.Service
}

func NewCalendarHandler(service calendarService.CalendarService, cognitoSvc *cognito.Service) *CalendarHandler {
	return &CalendarHandler{service: service, cognitoSvc: cognitoSvc}
}

// AddDay godoc
// @Summary      Add calendar day
// @Description  Mark a date as a holiday (HOLIDAY) or override a weekend as a working day (WORKDAY)
// @Tags         Calendar
// @Accept       json
// @Produce      json
// @Param        body  body      calendarService.AddCalendarDayRequest  true  "Day data"
// @Success      201   {object}  calendarService.AddCalendarDayResponse
// @Failure      400   {object}  map[string]string
// @Failure      401   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Security     BearerAuth
// @Router       /calendar [post]
func (h *CalendarHandler) AddDay(c *gin.Context) {
	callerUserID, ok := h.extractCallerUserID(c)
	if !ok {
		return
	}

	var req calendarService.AddCalendarDayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	resp, err := h.service.AddDay(c.Request.Context(), callerUserID, req)
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

// DeleteDay godoc
// @Summary      Delete calendar day
// @Description  Remove a holiday or workday override from the calendar
// @Tags         Calendar
// @Param        id  path  string  true  "Calendar Day ID (UUID)"
// @Success      204
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /calendar/{id} [delete]
func (h *CalendarHandler) DeleteDay(c *gin.Context) {
	callerUserID, ok := h.extractCallerUserID(c)
	if !ok {
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.service.DeleteDay(c.Request.Context(), callerUserID, id); err != nil {
		h.handleError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// GetMonthSummary godoc
// @Summary      Get month summary
// @Description  Returns a full breakdown of working days, weekends, and holidays for a given month
// @Tags         Calendar
// @Produce      json
// @Param        month  query     string  false  "Month in YYYY-MM format (default: current month)"
// @Success      200    {object}  calendarService.MonthSummaryResponse
// @Failure      400    {object}  map[string]string
// @Failure      401    {object}  map[string]string
// @Failure      500    {object}  map[string]string
// @Security     BearerAuth
// @Router       /calendar/summary [get]
func (h *CalendarHandler) GetMonthSummary(c *gin.Context) {
	callerUserID, ok := h.extractCallerUserID(c)
	if !ok {
		return
	}

	month := c.Query("month")
	if month == "" {
		month = time.Now().Format("2006-01")
	}

	resp, err := h.service.GetMonthSummary(c.Request.Context(), callerUserID, month)
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *CalendarHandler) extractCallerUserID(c *gin.Context) (uuid.UUID, bool) {
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
		log.Printf("[CalendarHandler] ParseTokenClaims failed: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
		return uuid.Nil, false
	}
	return userID, true
}

func (h *CalendarHandler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, calendarService.ErrCalendarDayNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "calendar day not found"})
	case errors.Is(err, calendarService.ErrInvalidDate):
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date format, use YYYY-MM-DD"})
	case errors.Is(err, calendarService.ErrInvalidType):
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid type, use HOLIDAY or WORKDAY"})
	case errors.Is(err, calendarService.ErrInvalidMonth):
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid month format, use YYYY-MM"})
	default:
		log.Printf("[CalendarHandler] error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}
