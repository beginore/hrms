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

// CreateEmployee godoc
// @Summary      Create employee
// @Description  Add a new employee record for the caller's organisation
// @Tags         Employees
// @Accept       json
// @Produce      json
// @Param        body  body      employeeService.CreateEmployeeRequest   true  "Employee data"
// @Success      201   {object}  employeeService.CreateEmployeeResponse
// @Failure      400   {object}  map[string]string
// @Failure      401   {object}  map[string]string
// @Failure      422   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Security     BearerAuth
// @Router       /employees [post]
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

// GetEmployee godoc
// @Summary      Get employee
// @Description  Retrieve a single employee by their ID
// @Tags         Employees
// @Produce      json
// @Param        id   path      string  true  "Employee ID (UUID)"
// @Success      200  {object}  employeeService.EmployeeResponse
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /employees/{id} [get]
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

// ListEmployees godoc
// @Summary      List employees
// @Description  Retrieve all employees belonging to the caller's organisation
// @Tags         Employees
// @Produce      json
// @Success      200  {array}   employeeService.EmployeeResponse
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /employees [get]
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

// ListDepartments GET /departments
func (h *EmployeeHandler) ListDepartments(c *gin.Context) {
	callerUserID, ok := h.extractCallerUserID(c)
	if !ok {
		return
	}

	resp, err := h.service.ListDepartments(c.Request.Context(), callerUserID)
	if err != nil {
		switch {
		case errors.Is(err, employeeService.ErrInvalidUserID):
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		default:
			log.Printf("[Handler] ListDepartments error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ListPositions GET /positions
func (h *EmployeeHandler) ListPositions(c *gin.Context) {
	callerUserID, ok := h.extractCallerUserID(c)
	if !ok {
		return
	}

	resp, err := h.service.ListPositions(c.Request.Context(), callerUserID)
	if err != nil {
		switch {
		case errors.Is(err, employeeService.ErrInvalidUserID):
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		default:
			log.Printf("[Handler] ListPositions error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		return
	}

	c.JSON(http.StatusOK, resp)
}

// UpdateRole godoc
// @Summary      Update employee role
// @Description  Change the role of an employee
// @Tags         Employees
// @Accept       json
// @Param        id    path  string                              true  "Employee ID (UUID)"
// @Param        body  body  employeeService.UpdateEmployeeRoleRequest  true  "New role"
// @Success      204
// @Failure      400   {object}  map[string]string
// @Failure      404   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Security     BearerAuth
// @Router       /employees/{id}/role [patch]
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

// UpdateSalary godoc
// @Summary      Update employee salary
// @Description  Change the salary rate of an employee
// @Tags         Employees
// @Accept       json
// @Param        id    path  string                                true  "Employee ID (UUID)"
// @Param        body  body  employeeService.UpdateEmployeeSalaryRequest  true  "New salary rate"
// @Success      204
// @Failure      400   {object}  map[string]string
// @Failure      404   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Security     BearerAuth
// @Router       /employees/{id}/salary [patch]
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

// UpdateStatus godoc
// @Summary      Update employee status
// @Description  Change the employment status of an employee
// @Tags         Employees
// @Accept       json
// @Param        id    path  string                                true  "Employee ID (UUID)"
// @Param        body  body  employeeService.UpdateEmployeeStatusRequest  true  "New status"
// @Success      204
// @Failure      400   {object}  map[string]string
// @Failure      404   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Security     BearerAuth
// @Router       /employees/{id}/status [patch]
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

// UpdateDepartment godoc
// @Summary      Update employee department
// @Description  Move an employee to a different department
// @Tags         Employees
// @Accept       json
// @Param        id    path  string                                    true  "Employee ID (UUID)"
// @Param        body  body  employeeService.UpdateEmployeeDepartmentRequest  true  "New department"
// @Success      204
// @Failure      400   {object}  map[string]string
// @Failure      404   {object}  map[string]string
// @Failure      422   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Security     BearerAuth
// @Router       /employees/{id}/department [patch]
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

// UpdatePosition godoc
// @Summary      Update employee position
// @Description  Change the position of an employee
// @Tags         Employees
// @Accept       json
// @Param        id    path  string                                   true  "Employee ID (UUID)"
// @Param        body  body  employeeService.UpdateEmployeePositionRequest  true  "New position"
// @Success      204
// @Failure      400   {object}  map[string]string
// @Failure      404   {object}  map[string]string
// @Failure      422   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Security     BearerAuth
// @Router       /employees/{id}/position [patch]
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

// DeleteEmployee godoc
// @Summary      Delete employee
// @Description  Permanently remove an employee record
// @Tags         Employees
// @Param        id  path  string  true  "Employee ID (UUID)"
// @Success      204
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /employees/{id} [delete]
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

func (h *EmployeeHandler) parseEmployeeID(c *gin.Context) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid employee id"})
		return uuid.Nil, false
	}
	return id, true
}

func (h *EmployeeHandler) handleMutationError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, employeeService.ErrEmployeeNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "employee not found"})
	default:
		log.Printf("[Handler] mutation error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}
