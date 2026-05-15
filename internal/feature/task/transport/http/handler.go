package http

import (
	"errors"
	taskService "hrms/internal/feature/task/service"
	"hrms/internal/infrastructure/middleware"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type TaskHandler struct {
	service taskService.TaskService
}

func NewTaskHandler(service taskService.TaskService) *TaskHandler {
	return &TaskHandler{service: service}
}

// AssignTask godoc
// @Summary      Assign task to employee
// @Description  Admin or SysAdmin assigns a task to an employee
// @Tags         Tasks
// @Accept       json
// @Produce      json
// @Param        body  body      taskService.AssignTaskRequest  true  "Task data"
// @Success      201  {object}  taskService.AssignTaskResponse
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      422  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /tasks [post]
func (h *TaskHandler) AssignTask(c *gin.Context) {
	var req taskService.AssignTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	resp, err := h.service.AssignTask(c.Request.Context(), h.callerUserID(c), req)
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

// ListTasks godoc
// @Summary      List tasks
// @Description  Admin/SysAdmin sees all org tasks; Employee sees own assigned tasks
// @Tags         Tasks
// @Produce      json
// @Success      200  {array}   taskService.TaskResponse
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /tasks [get]
func (h *TaskHandler) ListTasks(c *gin.Context) {
	resp, err := h.service.ListTasks(c.Request.Context(), h.callerUserID(c), h.callerRole(c))
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// GetTask godoc
// @Summary      Get task
// @Description  Retrieve a specific task by ID. Employee can only view their own assigned tasks.
// @Tags         Tasks
// @Produce      json
// @Param        id  path      string  true  "Task ID (UUID)"
// @Success      200  {object}  taskService.TaskResponse
// @Failure      400  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /tasks/{id} [get]
func (h *TaskHandler) GetTask(c *gin.Context) {
	id, ok := h.parseUUIDParam(c, "id")
	if !ok {
		return
	}

	resp, err := h.service.GetTask(c.Request.Context(), h.callerUserID(c), h.callerRole(c), id)
	if err != nil {
		h.handleError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// SubmitReport godoc
// @Summary      Submit task report
// @Description  Assigned employee submits their work report for the task. Changes status to SUBMITTED.
// @Tags         Tasks
// @Accept       json
// @Produce      json
// @Param        id    path  string                          true  "Task ID (UUID)"
// @Param        body  body  taskService.SubmitReportRequest  true  "Report data"
// @Success      204
// @Failure      400  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      409  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /tasks/{id}/submit [post]
func (h *TaskHandler) SubmitReport(c *gin.Context) {
	id, ok := h.parseUUIDParam(c, "id")
	if !ok {
		return
	}

	var req taskService.SubmitReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if err := h.service.SubmitReport(c.Request.Context(), h.callerUserID(c), id, req); err != nil {
		h.handleError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ReviewTask godoc
// @Summary      Approve or reject task report
// @Description  Admin or SysAdmin approves or rejects a submitted task report
// @Tags         Tasks
// @Accept       json
// @Param        id    path  string                        true  "Task ID (UUID)"
// @Param        body  body  taskService.ReviewTaskRequest  true  "Action (approve|reject)"
// @Success      204
// @Failure      400  {object}  map[string]string
// @Failure      401  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      409  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /tasks/{id}/review [patch]
func (h *TaskHandler) ReviewTask(c *gin.Context) {
	id, ok := h.parseUUIDParam(c, "id")
	if !ok {
		return
	}

	var req taskService.ReviewTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if err := h.service.ReviewTask(c.Request.Context(), h.callerUserID(c), id, req); err != nil {
		h.handleError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ─── Private helpers ──────────────────────────────────────────────────────────

func (h *TaskHandler) callerUserID(c *gin.Context) uuid.UUID {
	if value, exists := c.Get(middleware.UserIDKey); exists {
		if id, ok := value.(uuid.UUID); ok {
			return id
		}
	}
	return uuid.Nil
}

func (h *TaskHandler) callerRole(c *gin.Context) string {
	if value, exists := c.Get(middleware.RoleKey); exists {
		if role, ok := value.(string); ok {
			return role
		}
	}
	return ""
}

func (h *TaskHandler) parseUUIDParam(c *gin.Context, param string) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param(param))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid " + param})
		return uuid.Nil, false
	}
	return id, true
}

func (h *TaskHandler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, taskService.ErrTaskNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
	case errors.Is(err, taskService.ErrTaskNotSubmitted):
		c.JSON(http.StatusConflict, gin.H{"error": "task report has not been submitted yet"})
	case errors.Is(err, taskService.ErrTaskAlreadyReviewed):
		c.JSON(http.StatusConflict, gin.H{"error": "task has already been reviewed"})
	case errors.Is(err, taskService.ErrTaskAlreadySubmitted):
		c.JSON(http.StatusConflict, gin.H{"error": "task report has already been submitted"})
	case errors.Is(err, taskService.ErrNotAssignedEmployee):
		c.JSON(http.StatusForbidden, gin.H{"error": "you are not the assigned employee for this task"})
	case errors.Is(err, taskService.ErrEmployeeNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "employee not found"})
	case errors.Is(err, taskService.ErrEmployeeNotInOrg):
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "employee does not belong to this organisation"})
	case errors.Is(err, taskService.ErrForbidden):
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
	case errors.Is(err, taskService.ErrInvalidTaskID):
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "invalid task id"})
	case errors.Is(err, taskService.ErrInvalidEmployeeID):
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "invalid employee id"})
	case errors.Is(err, taskService.ErrInvalidDateFormat):
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date format, use YYYY-MM-DD"})
	default:
		log.Printf("[TaskHandler] error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}
