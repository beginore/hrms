package http

import (
	"errors"
	consentService "hrms/internal/feature/consent/service"
	"log"
	"net/http"

	"hrms/internal/feature/organization/service"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type OrganizationHandler struct {
	signUpService  *service.SignUpService
	consentService consentService.ConsentService
}

func NewOrganizationHandler(signUpService *service.SignUpService, consentService consentService.ConsentService) *OrganizationHandler {
	return &OrganizationHandler{
		signUpService:  signUpService,
		consentService: consentService}
}

func (h *OrganizationHandler) CreateOrganization(c *gin.Context) {
	var req service.CreateOrganizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body",
		})
		return
	}

	resp, err := h.signUpService.CreateOrganization(c.Request.Context(), req)
	var ve = validator.ValidationErrors{}
	var cove = service.CreateOrganizationValidationError{}

	if err != nil {
		switch {
		case errors.Is(err, service.ErrEmailAlreadyExists):
			c.JSON(http.StatusConflict, gin.H{
				"error": "email already exists",
			})
		case errors.Is(err, service.ErrVATAlreadyExists):
			c.JSON(http.StatusConflict, gin.H{
				"error": "VAT already registered",
			})
		case errors.As(err, &ve):
			c.JSON(http.StatusUnprocessableEntity, gin.H{
				"error":   "validation failed",
				"details": ve.Error(),
			})
		case errors.As(err, &cove):
			c.JSON(http.StatusUnprocessableEntity, gin.H{
				"error":   "validation failed",
				"details": cove.Error(),
			})
		case errors.Is(err, service.ErrPoliciesNotAccepted):
			c.JSON(http.StatusUnprocessableEntity, gin.H{
				"error": "must accept privacy policy and terms",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "internal server error",
			})
		}
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (h *OrganizationHandler) SubmitConsents(c *gin.Context) {
	var req consentService.RenewConsentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	if err := h.consentService.SubmitConsents(c.Request.Context(), req); err != nil {
		switch {
		case errors.Is(err, consentService.ErrInvalidOrganizationID),
			errors.Is(err, consentService.ErrInvalidUserID):
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}
	c.Status(http.StatusOK)
}

func (h *OrganizationHandler) ValidateConsents(c *gin.Context) {
	orgID := c.Query("organizationId")
	if orgID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "organizationId is required"})
		return
	}

	resp, err := h.consentService.ValidateConsents(c.Request.Context())
	if err != nil {
		switch {
		case errors.Is(err, consentService.ErrInvalidOrganizationID):
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *OrganizationHandler) GetDocuments(c *gin.Context) {
	docs, err := h.consentService.GetActiveDocuments(c.Request.Context())
	if err != nil {
		log.Printf("[Handler] GetDocuments error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, docs)
}

func (h *OrganizationHandler) VerifyOTP(c *gin.Context) {
	var req service.VerifyOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	err := h.signUpService.VerifyOTP(c.Request.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidOTP):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid or incorrect OTP code",
			})

		case errors.Is(err, service.ErrOTPExpired):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "OTP code has expired",
			})

		case errors.Is(err, service.ErrTooManyOTPAttempts):
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Too many failed attempts – please try again later",
			})

		case errors.Is(err, service.ErrUserAlreadyVerified):
			c.JSON(http.StatusConflict, gin.H{
				"error": "User is already verified",
			})

		case errors.Is(err, service.ErrUserNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error": "No account found for this email",
			})

		default:
			log.Printf("[Handler VerifyOTP] Unexpected error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to verify OTP – please try again or contact support",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "OTP verified successfully.",
	})
}
