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

// GenerateInvite godoc
// @Summary      Generate invite
// @Description  Create a time-limited invite link for a new employee and send it by email
// @Tags         Invites
// @Accept       json
// @Produce      json
// @Param        body  body      service.GenerateInviteRequest   true  "Invite details"
// @Success      201   {object}  service.GenerateInviteResponse
// @Failure      400   {object}  map[string]string
// @Failure      404   {object}  map[string]string
// @Failure      409   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /invites/generate [post]
func (h *Handler) GenerateInvite(c *gin.Context) {
	var req service.GenerateInviteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[Invite Generate] Invalid request body: %v", err)
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
			errors.Is(err, service.ErrInvalidEmail),
			errors.Is(err, service.ErrDepartmentIDRequired),
			errors.Is(err, service.ErrInvalidDepartmentID),
			errors.Is(err, service.ErrPositionIDRequired),
			errors.Is(err, service.ErrInvalidPositionID),
			errors.Is(err, service.ErrSalaryRateRequired),
			errors.Is(err, service.ErrInvalidSalaryRate),
			errors.Is(err, service.ErrStatusRequired):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, service.ErrOrganizationNotFound),
			errors.Is(err, service.ErrDepartmentNotFound),
			errors.Is(err, service.ErrPositionNotFound):
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

// VerifyInvite godoc
// @Summary      Verify invite code
// @Description  Validate an invite code and return the invitation details
// @Tags         Invites
// @Accept       json
// @Produce      json
// @Param        body  body      service.VerifyInviteRequest   true  "Invite code"
// @Success      200   {object}  service.VerifyInviteResponse
// @Failure      400   {object}  map[string]string
// @Failure      404   {object}  map[string]string
// @Failure      409   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /invites/verify [post]
func (h *Handler) VerifyInvite(c *gin.Context) {
	var req service.VerifyInviteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[Invite Verify] Invalid request body: %v", err)
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

// CompleteRegistration godoc
// @Summary      Complete registration
// @Description  Finalize a new employee's account using their invite code
// @Tags         Invites
// @Accept       json
// @Produce      json
// @Param        body  body      service.CompleteRegistrationRequest   true  "Registration data"
// @Success      201   {object}  service.CompleteRegistrationResponse
// @Failure      400   {object}  map[string]string
// @Failure      404   {object}  map[string]string
// @Failure      409   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /invites/complete-registration [post]
func (h *Handler) CompleteRegistration(c *gin.Context) {
	var req service.CompleteRegistrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[Invite CompleteRegistration] Invalid request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	resp, err := h.service.CompleteRegistration(c.Request.Context(), req)
	if err != nil {
		var passwordPolicyErr *service.PasswordPolicyError

		switch {
		case errors.Is(err, service.ErrInviteCodeRequired),
			errors.Is(err, service.ErrPasswordRequired),
			errors.Is(err, service.ErrPhoneNumberRequired),
			errors.Is(err, service.ErrInvalidPhoneNumber),
			errors.Is(err, service.ErrDepartmentIDRequired),
			errors.Is(err, service.ErrInvalidDepartmentID),
			errors.Is(err, service.ErrPositionIDRequired),
			errors.Is(err, service.ErrInvalidPositionID),
			errors.Is(err, service.ErrSalaryRateRequired),
			errors.Is(err, service.ErrInvalidSalaryRate),
			errors.Is(err, service.ErrStatusRequired):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.As(err, &passwordPolicyErr):
			c.JSON(http.StatusBadRequest, gin.H{"error": passwordPolicyErr.Error()})
		case errors.Is(err, service.ErrInviteNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case errors.Is(err, service.ErrInviteExpired),
			errors.Is(err, service.ErrInviteAlreadyUsed),
			errors.Is(err, service.ErrPhoneNumberExists),
			errors.Is(err, service.ErrEmailAlreadyExists),
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
