package http

import (
	"errors"
	"log"
	"net/http"
	"strings"

	userService "hrms/internal/feature/user/service"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *userService.Service
}

func NewHandler(service *userService.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Me(c *gin.Context) {
	authHeader := strings.TrimSpace(c.GetHeader("Authorization"))
	if !strings.HasPrefix(authHeader, "Bearer ") {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid authorization header"})
		return
	}

	accessToken := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
	resp, err := h.service.GetMe(c.Request.Context(), accessToken)
	if err != nil {
		switch {
		case errors.Is(err, userService.ErrInvalidAccessToken):
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		case errors.Is(err, userService.ErrUserNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			log.Printf("[User Profile Handler] Unexpected error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, resp)
}
