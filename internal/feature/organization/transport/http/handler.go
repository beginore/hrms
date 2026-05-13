package http

import (
	"errors"
	consentService "hrms/internal/feature/consent/service"
	"log"
	"net/http"

	"hrms/internal/feature/organization/service"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
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

// CreateOrganization godoc
// @Summary      Create organisation
// @Description  Register a new organisation along with its admin user
// @Tags         Organizations
// @Accept       json
// @Produce      json
// @Param        body  body      service.CreateOrganizationRequest   true  "Organisation details"
// @Success      201   {object}  service.CreateOrganizationResponse
// @Failure      400   {object}  map[string]string
// @Failure      409   {object}  map[string]string
// @Failure      422   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /organizations [post]
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
			c.JSON(http.StatusConflict, gin.H{"error": "email already exists"})
		case errors.Is(err, service.ErrVATAlreadyExists):
			c.JSON(http.StatusConflict, gin.H{"error": "VAT already registered"})
		case errors.As(err, &ve):
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "validation failed", "details": ve.Error()})
		case errors.As(err, &cove):
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "validation failed", "details": cove.Error()})
		case errors.Is(err, service.ErrPoliciesNotAccepted):
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "must accept privacy policy and terms"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}
	c.JSON(http.StatusCreated, resp)
}

// VerifyOTP godoc
// @Summary      Verify OTP
// @Description  Confirm organisation admin's email with the OTP sent during registration
// @Tags         Organizations
// @Accept       json
// @Produce      json
// @Param        body  body  service.VerifyOTPRequest  true  "OTP details"
// @Success      200   {object}  map[string]string
// @Failure      400   {object}  map[string]string
// @Failure      404   {object}  map[string]string
// @Failure      409   {object}  map[string]string
// @Failure      429   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /organizations/verify-otp [post]
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
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or incorrect OTP code"})
		case errors.Is(err, service.ErrOTPExpired):
			c.JSON(http.StatusBadRequest, gin.H{"error": "OTP code has expired"})
		case errors.Is(err, service.ErrTooManyOTPAttempts):
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Too many failed attempts – please try again later"})
		case errors.Is(err, service.ErrUserAlreadyVerified):
			c.JSON(http.StatusConflict, gin.H{"error": "User is already verified"})
		case errors.Is(err, service.ErrUserNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "No account found for this email"})
		default:
			log.Printf("[Handler VerifyOTP] Unexpected error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify OTP – please try again or contact support"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "OTP verified successfully."})
}

// SubmitConsents godoc
// @Summary      Submit consents
// @Description  Record the organisation's acceptance of policy documents
// @Tags         Organizations
// @Accept       json
// @Produce      json
// @Param        body  body  consentService.RenewConsentRequest  true  "Consent flags"
// @Success      200
// @Failure      400   {object}  map[string]string
// @Failure      422   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Security     BearerAuth
// @Router       /organizations/consents [post]
func (h *OrganizationHandler) SubmitConsents(c *gin.Context) {
	orgID := mustOrgID(c)
	if orgID == uuid.Nil {
		return
	}
	var req consentService.RenewConsentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	if err := h.consentService.SubmitConsents(c.Request.Context(), orgID, req); err != nil {
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

// ValidateConsents godoc
// @Summary      Validate consents
// @Description  Check whether the organisation's consents are up to date
// @Tags         Organizations
// @Produce      json
// @Param        organizationId  query     string  true  "Organisation ID"
// @Success      200             {object}  consentService.ConsentValidationResponse
// @Failure      400             {object}  map[string]string
// @Failure      422             {object}  map[string]string
// @Failure      500             {object}  map[string]string
// @Security     BearerAuth
// @Router       /organizations/consents/validate [get]
func (h *OrganizationHandler) ValidateConsents(c *gin.Context) {
	orgID := mustOrgID(c)
	if orgID == uuid.Nil {
		return
	}

	resp, err := h.consentService.ValidateConsents(c.Request.Context(), orgID)
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

// GetDocuments godoc
// @Summary      Get legal documents
// @Description  Return the list of active legal documents (privacy policy, terms, etc.)
// @Tags         Legal
// @Produce      json
// @Success      200  {array}   consentService.ConsentItemResponse
// @Failure      500  {object}  map[string]string
// @Router       /legal/documents [get]
func (h *OrganizationHandler) GetDocuments(c *gin.Context) {
	docs, err := h.consentService.GetActiveDocuments(c.Request.Context())
	if err != nil {
		log.Printf("[Handler] GetDocuments error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, docs)
}

// CreateDepartment godoc
// @Summary      Create department
// @Description  Add a new department to the caller's organisation
// @Tags         Departments
// @Accept       json
// @Produce      json
// @Param        body  body      service.CreateDepartmentRequest   true  "Department name"
// @Success      201   {object}  service.CreateDepartmentResponse
// @Failure      400   {object}  map[string]string
// @Failure      401   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Security     BearerAuth
// @Router       /organizations/departments [post]
func (h *OrganizationHandler) CreateDepartment(c *gin.Context) {
	orgID := mustOrgID(c)
	if orgID == uuid.Nil {
		return
	}
	var req service.CreateDepartmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	req.OrgID = orgID
	resp, err := h.signUpService.CreateDepartment(c.Request.Context(), req)
	if err != nil {
		log.Printf("[Handler] CreateDepartment error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusCreated, resp)
}

// ListDepartments godoc
// @Summary      List departments
// @Description  Return all departments belonging to the caller's organisation
// @Tags         Departments
// @Produce      json
// @Success      200  {array}   service.DepartmentItem
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /organizations/departments [get]
func (h *OrganizationHandler) ListDepartments(c *gin.Context) {
	orgID := mustOrgID(c)
	if orgID == uuid.Nil {
		return
	}
	items, err := h.signUpService.ListDepartments(c.Request.Context(), orgID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, items)
}

// DeleteDepartment godoc
// @Summary      Delete department
// @Description  Remove a department. Fails with 409 if employees are still assigned to it.
// @Tags         Departments
// @Param        id  path  string  true  "Department ID (UUID)"
// @Success      204
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      409  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /organizations/departments/{id} [delete]
func (h *OrganizationHandler) DeleteDepartment(c *gin.Context) {
	orgID := mustOrgID(c)
	if orgID == uuid.Nil {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid department id"})
		return
	}
	if err := h.signUpService.DeleteDepartment(c.Request.Context(), id, orgID); err != nil {
		switch {
		case errors.Is(err, service.ErrDepartmentHasEmployees):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			log.Printf("[Handler] DeleteDepartment error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}
	c.Status(http.StatusNoContent)
}

// CreatePosition godoc
// @Summary      Create position
// @Description  Add a new position to the caller's organisation
// @Tags         Positions
// @Accept       json
// @Produce      json
// @Param        body  body      service.CreatePositionRequest   true  "Position name"
// @Success      201   {object}  service.CreatePositionResponse
// @Failure      400   {object}  map[string]string
// @Failure      401   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Security     BearerAuth
// @Router       /organizations/positions [post]
func (h *OrganizationHandler) CreatePosition(c *gin.Context) {
	orgID := mustOrgID(c)
	if orgID == uuid.Nil {
		return
	}
	var req service.CreatePositionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	req.OrgID = orgID
	resp, err := h.signUpService.CreatePosition(c.Request.Context(), req)
	if err != nil {
		log.Printf("[Handler] CreatePosition error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusCreated, resp)
}

// ListPositions godoc
// @Summary      List positions
// @Description  Return all positions belonging to the caller's organisation
// @Tags         Positions
// @Produce      json
// @Success      200  {array}   service.PositionItem
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /organizations/positions [get]
func (h *OrganizationHandler) ListPositions(c *gin.Context) {
	orgID := mustOrgID(c)
	if orgID == uuid.Nil {
		return
	}
	items, err := h.signUpService.ListPositions(c.Request.Context(), orgID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, items)
}

// DeletePosition godoc
// @Summary      Delete position
// @Description  Remove a position. Fails with 409 if employees are still assigned to it.
// @Tags         Positions
// @Param        id  path  string  true  "Position ID (UUID)"
// @Success      204
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      409  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /organizations/positions/{id} [delete]
func (h *OrganizationHandler) DeletePosition(c *gin.Context) {
	orgID := mustOrgID(c)
	if orgID == uuid.Nil {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid position id"})
		return
	}
	if err := h.signUpService.DeletePosition(c.Request.Context(), id, orgID); err != nil {
		switch {
		case errors.Is(err, service.ErrPositionHasEmployees):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			log.Printf("[Handler] DeletePosition error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}
	c.Status(http.StatusNoContent)
}

func mustOrgID(c *gin.Context) uuid.UUID {
	val, exists := c.Get("orgID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return uuid.Nil
	}
	id, ok := val.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid orgID in context"})
		return uuid.Nil
	}
	return id
}
