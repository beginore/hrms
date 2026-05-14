package http

import (
	"errors"
	"log"
	"net/http"

	eventsService "hrms/internal/feature/events/service"
	"hrms/internal/infrastructure/app/cognito"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	service *eventsService.Service
}

func NewHandler(service *eventsService.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Upcoming(c *gin.Context) {
	userID, err := currentUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.service.ListUpcoming(c.Request.Context(), userID)
	if err != nil {
		log.Printf("[Events Upcoming] error for user=%s: %v", userID, err)
		handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) My(c *gin.Context) {
	userID, err := currentUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.service.ListMy(c.Request.Context(), userID)
	if err != nil {
		log.Printf("[Events My] error for user=%s: %v", userID, err)
		handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) Create(c *gin.Context) {
	userID, err := currentUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	var req eventsService.CreateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	resp, err := h.service.Create(c.Request.Context(), userID, req)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *Handler) Update(c *gin.Context) {
	userID, err := currentUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	var req eventsService.UpdateEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	resp, err := h.service.Update(c.Request.Context(), userID, c.Param("id"), req)
	if err != nil {
		handleServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) Delete(c *gin.Context) {
	userID, err := currentUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	if err := h.service.Delete(c.Request.Context(), userID, c.Param("id")); err != nil {
		handleServiceError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func handleServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, eventsService.ErrInvalidEventID),
		errors.Is(err, eventsService.ErrTitleRequired),
		errors.Is(err, eventsService.ErrDescriptionRequired),
		errors.Is(err, eventsService.ErrScopeRequired),
		errors.Is(err, eventsService.ErrInvalidScope),
		errors.Is(err, eventsService.ErrDepartmentRequired),
		errors.Is(err, eventsService.ErrDepartmentForbidden),
		errors.Is(err, eventsService.ErrInvalidDepartmentID),
		errors.Is(err, eventsService.ErrStartsAtRequired),
		errors.Is(err, eventsService.ErrEndsAtRequired),
		errors.Is(err, eventsService.ErrInvalidStartsAt),
		errors.Is(err, eventsService.ErrInvalidEndsAt),
		errors.Is(err, eventsService.ErrInvalidEventTime),
		errors.Is(err, eventsService.ErrDepartmentNotBound):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, eventsService.ErrPermissionDenied):
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	case errors.Is(err, eventsService.ErrEventNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}

func currentUserID(c *gin.Context) (string, error) {
	value, exists := c.Get(cognito.ContextUserIDKey)
	if !exists {
		return "", errors.New("missing authenticated user")
	}
	userID, ok := value.(uuid.UUID)
	if !ok {
		return "", errors.New("invalid authenticated user")
	}
	return userID.String(), nil
}
