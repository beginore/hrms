package http

import (
	"errors"
	"log"
	"net/http"

	"hrms/internal/feature/invite/service"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *service.Service
}

func NewHandler(service *service.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) GenerateInvite(c *gin.Context) {
	var req service.GenerateInviteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	resp, err := h.service.GenerateInvite(c.Request.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrOrganizationIDRequired),
			errors.Is(err, service.ErrInvalidOrganizationID),
			errors.Is(err, service.ErrFirstNameRequired),
			errors.Is(err, service.ErrLastNameRequired),
			errors.Is(err, service.ErrEmailRequired),
			errors.Is(err, service.ErrInvalidEmail):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, service.ErrOrganizationNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case errors.Is(err, service.ErrGenerateInvite):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		case errors.Is(err, service.ErrInviteEmailUnavailable):
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		default:
			log.Printf("[Invite Generate] Unexpected error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	c.JSON(http.StatusCreated, resp)
}

func (h *Handler) VerifyInvite(c *gin.Context) {
	var req service.VerifyInviteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	resp, err := h.service.VerifyInvite(c.Request.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInviteCodeRequired):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, service.ErrInviteNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case errors.Is(err, service.ErrInviteExpired),
			errors.Is(err, service.ErrInviteAlreadyUsed):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			log.Printf("[Invite Verify] Unexpected error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *Handler) CompleteRegistration(c *gin.Context) {
	var req service.CompleteRegistrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	resp, err := h.service.CompleteRegistration(c.Request.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInviteCodeRequired),
			errors.Is(err, service.ErrPasswordRequired),
			errors.Is(err, service.ErrPhoneNumberRequired):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, service.ErrInviteNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case errors.Is(err, service.ErrInviteExpired),
			errors.Is(err, service.ErrInviteAlreadyUsed),
			errors.Is(err, service.ErrUserAlreadyExists):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			log.Printf("[Invite CompleteRegistration] Unexpected error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	c.JSON(http.StatusCreated, resp)
}
