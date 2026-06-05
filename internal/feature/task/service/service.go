package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	taskRepository "hrms/internal/feature/task/repository"

	"github.com/google/uuid"
)

const dateLayout = "2006-01-02"

type TaskService interface {
	AssignTask(ctx context.Context, callerUserID uuid.UUID, req AssignTaskRequest) (*AssignTaskResponse, error)
	GetTask(ctx context.Context, callerUserID uuid.UUID, callerRole string, id uuid.UUID) (*TaskResponse, error)
	ListTasks(ctx context.Context, callerUserID uuid.UUID, callerRole string) ([]TaskResponse, error)
	SubmitReport(ctx context.Context, callerUserID uuid.UUID, id uuid.UUID, req SubmitReportRequest) error
	ReviewTask(ctx context.Context, callerUserID uuid.UUID, id uuid.UUID, req ReviewTaskRequest) error
}

type taskService struct {
	repo taskRepository.TaskRepository
}

func NewTaskService(repo taskRepository.TaskRepository) TaskService {
	return &taskService{repo: repo}
}

func (s *taskService) AssignTask(ctx context.Context, callerUserID uuid.UUID, req AssignTaskRequest) (*AssignTaskResponse, error) {
	assignedTo, err := uuid.Parse(req.AssignedTo)
	if err != nil {
		return nil, ErrInvalidEmployeeID
	}

	dueDate, err := time.Parse(dateLayout, req.DueDate)
	if err != nil {
		return nil, ErrInvalidDateFormat
	}

	orgID, err := s.repo.GetOrgIDByUserID(ctx, callerUserID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabaseQueryFailed, err)
	}

	exists, err := s.repo.EmployeeExistsInOrg(ctx, assignedTo, orgID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabaseQueryFailed, err)
	}
	if !exists {
		return nil, ErrEmployeeNotInOrg
	}

	log.Printf("[Task] AssignTask assigned_to=%s title=%q", assignedTo, req.Title)
	id, err := s.repo.CreateTask(ctx, callerUserID, taskRepository.CreateTaskParams{
		AssignedTo:  assignedTo,
		Title:       req.Title,
		Description: req.Description,
		DueDate:     dueDate,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabaseSaveFailed, err)
	}
	return &AssignTaskResponse{TaskID: id.String()}, nil
}

func (s *taskService) GetTask(ctx context.Context, callerUserID uuid.UUID, callerRole string, id uuid.UUID) (*TaskResponse, error) {
	task, err := s.repo.GetTaskByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTaskNotFound
		}
		return nil, fmt.Errorf("%w: %v", ErrDatabaseQueryFailed, err)
	}

	if canManageTasks(callerRole) {
		if err := s.ensureTaskInCallerOrg(ctx, callerUserID, task); err != nil {
			return nil, err
		}
	} else {
		employeeID, err := s.repo.GetEmployeeIDByUserID(ctx, callerUserID)
		if err != nil {
			return nil, ErrEmployeeNotFound
		}
		if task.AssignedTo != employeeID {
			return nil, ErrForbidden
		}
	}

	resp := mapTaskResponse(task)
	return &resp, nil
}

func (s *taskService) ListTasks(ctx context.Context, callerUserID uuid.UUID, callerRole string) ([]TaskResponse, error) {
	if canManageTasks(callerRole) {
		items, err := s.repo.ListTasksByOrg(ctx, callerUserID)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrDatabaseQueryFailed, err)
		}
		return mapTaskResponses(items), nil
	}

	employeeID, err := s.repo.GetEmployeeIDByUserID(ctx, callerUserID)
	if err != nil {
		return nil, ErrEmployeeNotFound
	}
	items, err := s.repo.ListTasksByEmployee(ctx, employeeID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabaseQueryFailed, err)
	}
	return mapTaskResponses(items), nil
}

func (s *taskService) SubmitReport(ctx context.Context, callerUserID uuid.UUID, id uuid.UUID, req SubmitReportRequest) error {
	task, err := s.repo.GetTaskByID(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrTaskNotFound
	}
	if err != nil {
		return fmt.Errorf("%w: %v", ErrDatabaseQueryFailed, err)
	}

	employeeID, err := s.repo.GetEmployeeIDByUserID(ctx, callerUserID)
	if err != nil {
		return ErrEmployeeNotFound
	}
	if task.AssignedTo != employeeID {
		return ErrNotAssignedEmployee
	}

	if task.Status == "SUBMITTED" {
		return ErrTaskAlreadySubmitted
	}
	if task.Status == "APPROVED" || task.Status == "REJECTED" {
		return ErrTaskAlreadyReviewed
	}

	log.Printf("[Task] SubmitReport task=%s employee=%s", id, employeeID)
	if err := s.repo.SubmitReport(ctx, taskRepository.SubmitReportParams{
		ID:                id,
		ReportDescription: req.Description,
		ReportURL:         req.ReportURL,
	}); err != nil {
		return fmt.Errorf("%w: %v", ErrDatabaseSaveFailed, err)
	}
	return nil
}

func (s *taskService) ReviewTask(ctx context.Context, callerUserID uuid.UUID, id uuid.UUID, req ReviewTaskRequest) error {
	task, err := s.repo.GetTaskByID(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrTaskNotFound
	}
	if err != nil {
		return fmt.Errorf("%w: %v", ErrDatabaseQueryFailed, err)
	}
	if err := s.ensureTaskInCallerOrg(ctx, callerUserID, task); err != nil {
		return err
	}

	if task.Status == "PENDING" {
		return ErrTaskNotSubmitted
	}
	if task.Status == "APPROVED" || task.Status == "REJECTED" {
		return ErrTaskAlreadyReviewed
	}

	var status string
	switch req.Action {
	case "approve":
		status = "APPROVED"
	case "reject":
		status = "REJECTED"
	default:
		return ErrInvalidReviewAction
	}

	log.Printf("[Task] ReviewTask id=%s action=%s", id, req.Action)
	if err := s.repo.ReviewTask(ctx, callerUserID, taskRepository.ReviewTaskParams{
		ID:     id,
		Status: status,
	}); err != nil {
		return fmt.Errorf("%w: %v", ErrDatabaseSaveFailed, err)
	}
	return nil
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

func canManageTasks(role string) bool {
	switch role {
	case "SysAdmin", "Admin":
		return true
	default:
		return false
	}
}

func (s *taskService) ensureTaskInCallerOrg(ctx context.Context, callerUserID uuid.UUID, task taskRepository.Task) error {
	orgID, err := s.repo.GetOrgIDByUserID(ctx, callerUserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrForbidden
		}
		return fmt.Errorf("%w: %v", ErrDatabaseQueryFailed, err)
	}
	if task.OrgID != orgID {
		return ErrForbidden
	}
	return nil
}

func mapTaskResponse(t taskRepository.Task) TaskResponse {
	resp := TaskResponse{
		ID:                t.ID.String(),
		CreatedBy:         t.CreatedBy.String(),
		AssignedTo:        t.AssignedTo.String(),
		Title:             t.Title,
		Description:       t.Description,
		DueDate:           t.DueDate.Format(dateLayout),
		Status:            t.Status,
		ReportDescription: t.ReportDescription,
		ReportURL:         t.ReportURL,
		CreatedAt:         t.CreatedAt.Format(time.RFC3339),
	}
	if t.SubmittedAt != nil {
		s := t.SubmittedAt.Format(time.RFC3339)
		resp.SubmittedAt = &s
	}
	if t.ReviewedBy != nil {
		s := t.ReviewedBy.String()
		resp.ReviewedBy = &s
	}
	if t.ReviewedAt != nil {
		s := t.ReviewedAt.Format(time.RFC3339)
		resp.ReviewedAt = &s
	}
	return resp
}

func mapTaskResponses(items []taskRepository.Task) []TaskResponse {
	out := make([]TaskResponse, len(items))
	for i, item := range items {
		out[i] = mapTaskResponse(item)
	}
	return out
}
