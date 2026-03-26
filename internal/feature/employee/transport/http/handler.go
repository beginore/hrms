package http

import (
	"errors"
	employeeService "hrms/internal/feature/employee/service"
	"hrms/internal/infrastructure/app/cognito"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type EmployeeHandler struct {
	service    employeeService.EmployeeService
	cognitoSvc *cognito.Service
}

func NewEmployeeHandler(service employeeService.EmployeeService, cognitoSvc *cognito.Service) *EmployeeHandler {
	return &EmployeeHandler{
		service:    service,
		cognitoSvc: cognitoSvc,
	}
}

// CreateEmployee POST /employees
func (h *EmployeeHandler) CreateEmployee(c *gin.Context) {
	callerUserID, ok := h.extractCallerUserID(c)
	if !ok {
		return
	}

	var req employeeService.CreateEmployeeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	resp, err := h.service.CreateEmployee(c.Request.Context(), callerUserID, req)
	if err != nil {
		switch {
		case errors.Is(err, employeeService.ErrInvalidUserID):
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "invalid user id"})
		case errors.Is(err, employeeService.ErrInvalidDepartmentID):
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "invalid department id"})
		case errors.Is(err, employeeService.ErrInvalidPositionID):
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "invalid position id"})
		default:
			log.Printf("[Handler] CreateEmployee error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// GetEmployee GET /employees/:id
func (h *EmployeeHandler) GetEmployee(c *gin.Context) {
	id, ok := h.parseEmployeeID(c)
	if !ok {
		return
	}

	resp, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, employeeService.ErrEmployeeNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "employee not found"})
		default:
			log.Printf("[Handler] GetEmployee error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ListEmployees GET /employees
func (h *EmployeeHandler) ListEmployees(c *gin.Context) {
	callerUserID, ok := h.extractCallerUserID(c)
	if !ok {
		return
	}

	resp, err := h.service.GetByOrgID(c.Request.Context(), callerUserID)
	if err != nil {
		switch {
		case errors.Is(err, employeeService.ErrInvalidUserID):
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		default:
			log.Printf("[Handler] ListEmployees error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, resp)
}

// UpdateRole PATCH /employees/:id/role
func (h *EmployeeHandler) UpdateRole(c *gin.Context) {
	id, ok := h.parseEmployeeID(c)
	if !ok {
		return
	}

	var req employeeService.UpdateEmployeeRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if err := h.service.UpdateRole(c.Request.Context(), id, req); err != nil {
		h.handleMutationError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// UpdateSalary PATCH /employees/:id/salary
func (h *EmployeeHandler) UpdateSalary(c *gin.Context) {
	id, ok := h.parseEmployeeID(c)
	if !ok {
		return
	}

	var req employeeService.UpdateEmployeeSalaryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if err := h.service.UpdateSalary(c.Request.Context(), id, req); err != nil {
		h.handleMutationError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// UpdateStatus PATCH /employees/:id/status
func (h *EmployeeHandler) UpdateStatus(c *gin.Context) {
	id, ok := h.parseEmployeeID(c)
	if !ok {
		return
	}

	var req employeeService.UpdateEmployeeStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if err := h.service.UpdateStatus(c.Request.Context(), id, req); err != nil {
		h.handleMutationError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// UpdateDepartment PATCH /employees/:id/department
func (h *EmployeeHandler) UpdateDepartment(c *gin.Context) {
	id, ok := h.parseEmployeeID(c)
	if !ok {
		return
	}

	var req employeeService.UpdateEmployeeDepartmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if err := h.service.UpdateDepartment(c.Request.Context(), id, req); err != nil {
		switch {
		case errors.Is(err, employeeService.ErrInvalidDepartmentID):
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "invalid department id"})
		default:
			h.handleMutationError(c, err)
		}
		return
	}

	c.Status(http.StatusNoContent)
}

// UpdatePosition PATCH /employees/:id/position
func (h *EmployeeHandler) UpdatePosition(c *gin.Context) {
	id, ok := h.parseEmployeeID(c)
	if !ok {
		return
	}

	var req employeeService.UpdateEmployeePositionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if err := h.service.UpdatePosition(c.Request.Context(), id, req); err != nil {
		switch {
		case errors.Is(err, employeeService.ErrInvalidPositionID):
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "invalid position id"})
		default:
			h.handleMutationError(c, err)
		}
		return
	}

	c.Status(http.StatusNoContent)
}

// DeleteEmployee DELETE /employees/:id
func (h *EmployeeHandler) DeleteEmployee(c *gin.Context) {
	id, ok := h.parseEmployeeID(c)
	if !ok {
		return
	}

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		h.handleMutationError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// extractCallerUserID parses the Bearer token and returns the userID from the sub claim.
func (h *EmployeeHandler) extractCallerUserID(c *gin.Context) (uuid.UUID, bool) {
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
		log.Printf("[Handler] ParseTokenClaims failed: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
		return uuid.Nil, false
	}

	return userID, true
}

// parseEmployeeID extracts and validates the :id path param.
func (h *EmployeeHandler) parseEmployeeID(c *gin.Context) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid employee id"})
		return uuid.Nil, false
	}
	return id, true
}

// handleMutationError handles errors common to all mutating operations.
func (h *EmployeeHandler) handleMutationError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, employeeService.ErrEmployeeNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "employee not found"})
	default:
		log.Printf("[Handler] mutation error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}
