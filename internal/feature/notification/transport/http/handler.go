package http

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"hrms/internal/feature/notification/service"
	"hrms/internal/infrastructure/app/cognito"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	service *service.Service
}

func NewHandler(service *service.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) ListNotifications(c *gin.Context) {
	var req service.ListNotificationsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid query params"})
		return
	}

	userID, err := currentUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.service.ListNotifications(c.Request.Context(), userID, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUserIDRequired),
			errors.Is(err, service.ErrInvalidUserID),
			errors.Is(err, service.ErrInvalidPaginationLimit),
			errors.Is(err, service.ErrInvalidPaginationShift):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			log.Printf("[Notification List] Unexpected error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *Handler) MarkAsRead(c *gin.Context) {
	userID, err := currentUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	err = h.service.MarkAsRead(c.Request.Context(), userID, c.Param("id"))
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUserIDRequired),
			errors.Is(err, service.ErrInvalidUserID),
			errors.Is(err, service.ErrInvalidNotificationID):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, service.ErrNotificationNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			log.Printf("[Notification MarkAsRead] Unexpected error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) MarkAllAsRead(c *gin.Context) {
	userID, err := currentUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.service.MarkAllAsRead(c.Request.Context(), userID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUserIDRequired),
			errors.Is(err, service.ErrInvalidUserID):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			log.Printf("[Notification MarkAllAsRead] Unexpected error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, resp)
}

func currentUserID(c *gin.Context) (string, error) {
	value, exists := c.Get(cognito.ContextUserIDKey)
	if !exists {
		return "", fmt.Errorf("missing authenticated user")
	}

	userID, ok := value.(uuid.UUID)
	if !ok {
		return "", fmt.Errorf("invalid authenticated user")
	}

	return userID.String(), nil
}
