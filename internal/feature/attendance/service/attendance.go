package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	attendanceRepository "hrms/internal/feature/attendance/repository"

	"github.com/google/uuid"
)

const dateLayout = "2006-01-02"

var validLeaveTypes = map[string]bool{
	"SICK_LEAVE": true, "VACATION": true, "REMOTE": true,
	"BUSINESS_TRIP": true, "UNPAID_LEAVE": true,
}

type AttendanceService interface {
	// work schedules
	SetWorkSchedule(ctx context.Context, callerUserID uuid.UUID, req SetWorkScheduleRequest) error
	GetWorkSchedule(ctx context.Context, employeeID uuid.UUID) (*WorkScheduleResponse, error)

	// skud
	CreateSkudEvent(ctx context.Context, req CreateSkudEventRequest) (*CreateSkudEventResponse, error)

	// leave requests
	CreateLeaveRequest(ctx context.Context, callerUserID uuid.UUID, req CreateLeaveRequest) (*CreateLeaveResponse, error)
	GetLeaveRequestByID(ctx context.Context, id uuid.UUID) (*LeaveRequestResponse, error)
	ListLeaveRequestsByOrg(ctx context.Context, callerUserID uuid.UUID) ([]LeaveRequestResponse, error)
	ListLeaveRequestsByEmployee(ctx context.Context, employeeID uuid.UUID) ([]LeaveRequestResponse, error)
	ReviewLeaveRequest(ctx context.Context, callerUserID uuid.UUID, id uuid.UUID, req ReviewLeaveRequest) error

	// attendance records
	ListAttendanceByOrg(ctx context.Context, callerUserID uuid.UUID, startDate, endDate string) ([]AttendanceResponse, error)
	ListAttendanceByEmployee(ctx context.Context, employeeID uuid.UUID, startDate, endDate string) ([]AttendanceResponse, error)
}

type attendanceService struct {
	repo attendanceRepository.AttendanceRepository
}

func NewAttendanceService(repo attendanceRepository.AttendanceRepository) AttendanceService {
	return &attendanceService{repo: repo}
}

// ─── Work Schedule ────────────────────────────────────────────────────────────

func (s *attendanceService) SetWorkSchedule(ctx context.Context, callerUserID uuid.UUID, req SetWorkScheduleRequest) error {
	employeeID, err := uuid.Parse(req.EmployeeID)
	if err != nil {
		return ErrInvalidEmployeeID
	}

	exists, err := s.repo.EmployeeExists(ctx, employeeID)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrDatabaseQueryFailed, err)
	}
	if !exists {
		return ErrEmployeeNotFound
	}

	threshold := req.LateThresholdMinutes
	if threshold == 0 {
		threshold = 15
	}

	log.Printf("[Attendance] SetWorkSchedule employee=%s", employeeID)
	if err := s.repo.SetWorkSchedule(ctx, callerUserID, attendanceRepository.SetWorkScheduleParams{
		EmployeeID:           employeeID,
		WorkStart:            req.WorkStart,
		WorkEnd:              req.WorkEnd,
		LateThresholdMinutes: threshold,
	}); err != nil {
		return fmt.Errorf("%w: %v", ErrDatabaseSaveFailed, err)
	}
	return nil
}

func (s *attendanceService) GetWorkSchedule(ctx context.Context, employeeID uuid.UUID) (*WorkScheduleResponse, error) {
	ws := s.repo.GetWorkScheduleOrDefault(ctx, employeeID)
	return &WorkScheduleResponse{
		EmployeeID:           ws.EmployeeID.String(),
		WorkStart:            ws.WorkStart,
		WorkEnd:              ws.WorkEnd,
		LateThresholdMinutes: ws.LateThresholdMinutes,
	}, nil
}

// ─── SKUD ─────────────────────────────────────────────────────────────────────

func (s *attendanceService) CreateSkudEvent(ctx context.Context, req CreateSkudEventRequest) (*CreateSkudEventResponse, error) {
	employeeID, err := uuid.Parse(req.EmployeeID)
	if err != nil {
		return nil, ErrInvalidEmployeeID
	}

	if req.EventType != "ENTER" && req.EventType != "EXIT" {
		return nil, ErrInvalidSkudEventType
	}

	var occurredAt time.Time
	if req.OccurredAt == "" {
		occurredAt = time.Now()
	} else {
		occurredAt, err = time.Parse(time.RFC3339, req.OccurredAt)
		if err != nil {
			return nil, fmt.Errorf("invalid occurred_at: use RFC3339 format")
		}
	}

	log.Printf("[Attendance] SkudEvent employee=%s type=%s", employeeID, req.EventType)
	if err := s.repo.CreateSkudEvent(ctx, uuid.Nil, attendanceRepository.CreateSkudEventParams{
		EmployeeID: employeeID,
		EventType:  req.EventType,
		DeviceID:   req.DeviceID,
		OccurredAt: occurredAt,
	}); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabaseSaveFailed, err)
	}
	return &CreateSkudEventResponse{Message: "event recorded"}, nil
}

// ─── Leave Requests ───────────────────────────────────────────────────────────

func (s *attendanceService) CreateLeaveRequest(ctx context.Context, callerUserID uuid.UUID, req CreateLeaveRequest) (*CreateLeaveResponse, error) {
	employeeID, err := uuid.Parse(req.EmployeeID)
	if err != nil {
		return nil, ErrInvalidEmployeeID
	}

	if !validLeaveTypes[req.Type] {
		return nil, ErrInvalidLeaveType
	}

	startDate, err := time.Parse(dateLayout, req.StartDate)
	if err != nil {
		return nil, ErrInvalidDateFormat
	}
	endDate, err := time.Parse(dateLayout, req.EndDate)
	if err != nil {
		return nil, ErrInvalidDateFormat
	}
	if endDate.Before(startDate) {
		return nil, ErrInvalidDateRange
	}

	log.Printf("[Attendance] CreateLeaveRequest employee=%s type=%s", employeeID, req.Type)
	id, err := s.repo.CreateLeaveRequest(ctx, callerUserID, attendanceRepository.CreateLeaveRequestParams{
		EmployeeID:  employeeID,
		Type:        req.Type,
		StartDate:   startDate,
		EndDate:     endDate,
		Reason:      req.Reason,
		DocumentURL: req.DocumentURL,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabaseSaveFailed, err)
	}
	return &CreateLeaveResponse{LeaveRequestID: id.String()}, nil
}

func (s *attendanceService) GetLeaveRequestByID(ctx context.Context, id uuid.UUID) (*LeaveRequestResponse, error) {
	lr, err := s.repo.GetLeaveRequestByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrLeaveRequestNotFound
		}
		return nil, fmt.Errorf("%w: %v", ErrDatabaseQueryFailed, err)
	}
	resp := mapLeaveResponse(lr)
	return &resp, nil
}

func (s *attendanceService) ListLeaveRequestsByOrg(ctx context.Context, callerUserID uuid.UUID) ([]LeaveRequestResponse, error) {
	items, err := s.repo.GetLeaveRequestsByOrg(ctx, callerUserID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabaseQueryFailed, err)
	}
	return mapLeaveResponses(items), nil
}

func (s *attendanceService) ListLeaveRequestsByEmployee(ctx context.Context, employeeID uuid.UUID) ([]LeaveRequestResponse, error) {
	items, err := s.repo.GetLeaveRequestsByEmployee(ctx, employeeID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabaseQueryFailed, err)
	}
	return mapLeaveResponses(items), nil
}

func (s *attendanceService) ReviewLeaveRequest(ctx context.Context, callerUserID uuid.UUID, id uuid.UUID, req ReviewLeaveRequest) error {
	exists, err := s.repo.LeaveRequestExists(ctx, id)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrDatabaseQueryFailed, err)
	}
	if !exists {
		return ErrLeaveRequestNotFound
	}

	var status string
	switch req.Action {
	case "approve":
		status = "APPROVED"
	case "reject":
		status = "REJECTED"
	default:
		return fmt.Errorf("invalid action: use approve or reject")
	}

	log.Printf("[Attendance] ReviewLeaveRequest id=%s action=%s", id, req.Action)
	if err := s.repo.ReviewLeaveRequest(ctx, callerUserID, attendanceRepository.ReviewLeaveRequestParams{
		ID:         id,
		Status:     status,
		ReviewedBy: callerUserID,
	}); err != nil {
		return fmt.Errorf("%w: %v", ErrDatabaseSaveFailed, err)
	}
	return nil
}

// ─── Attendance Records ───────────────────────────────────────────────────────

func (s *attendanceService) ListAttendanceByOrg(ctx context.Context, callerUserID uuid.UUID, startDate, endDate string) ([]AttendanceResponse, error) {
	start, end, err := parsePeriod(startDate, endDate)
	if err != nil {
		return nil, err
	}
	records, err := s.repo.GetAttendanceByOrg(ctx, callerUserID, attendanceRepository.GetByPeriodParams{
		StartDate: start,
		EndDate:   end,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabaseQueryFailed, err)
	}
	return mapAttendanceResponses(records), nil
}

func (s *attendanceService) ListAttendanceByEmployee(ctx context.Context, employeeID uuid.UUID, startDate, endDate string) ([]AttendanceResponse, error) {
	start, end, err := parsePeriod(startDate, endDate)
	if err != nil {
		return nil, err
	}
	records, err := s.repo.GetAttendanceByEmployee(ctx, employeeID, attendanceRepository.GetByPeriodParams{
		StartDate: start,
		EndDate:   end,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabaseQueryFailed, err)
	}
	return mapAttendanceResponses(records), nil
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

func parsePeriod(startDate, endDate string) (time.Time, time.Time, error) {
	if startDate == "" {
		now := time.Now()
		startDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC).Format(dateLayout)
	}
	if endDate == "" {
		endDate = time.Now().Format(dateLayout)
	}
	start, err := time.Parse(dateLayout, startDate)
	if err != nil {
		return time.Time{}, time.Time{}, ErrInvalidDateFormat
	}
	end, err := time.Parse(dateLayout, endDate)
	if err != nil {
		return time.Time{}, time.Time{}, ErrInvalidDateFormat
	}
	if end.Before(start) {
		return time.Time{}, time.Time{}, ErrInvalidDateRange
	}
	return start, end, nil
}

func mapLeaveResponse(lr attendanceRepository.LeaveRequest) LeaveRequestResponse {
	resp := LeaveRequestResponse{
		ID:          lr.ID.String(),
		EmployeeID:  lr.EmployeeID.String(),
		Type:        lr.Type,
		StartDate:   lr.StartDate.Format(dateLayout),
		EndDate:     lr.EndDate.Format(dateLayout),
		Reason:      lr.Reason,
		DocumentURL: lr.DocumentURL,
		Status:      lr.Status,
		CreatedAt:   lr.CreatedAt.Format(time.RFC3339),
	}
	if lr.ReviewedBy != nil {
		s := lr.ReviewedBy.String()
		resp.ReviewedBy = &s
	}
	if lr.ReviewedAt != nil {
		s := lr.ReviewedAt.Format(time.RFC3339)
		resp.ReviewedAt = &s
	}
	return resp
}

func mapLeaveResponses(items []attendanceRepository.LeaveRequest) []LeaveRequestResponse {
	out := make([]LeaveRequestResponse, len(items))
	for i, item := range items {
		out[i] = mapLeaveResponse(item)
	}
	return out
}

func mapAttendanceResponse(ar attendanceRepository.AttendanceRecord) AttendanceResponse {
	resp := AttendanceResponse{
		ID:         ar.ID.String(),
		EmployeeID: ar.EmployeeID.String(),
		Date:       ar.Date.Format(dateLayout),
		Type:       ar.Type,
		Source:     ar.Source,
		Status:     ar.Status,
		Note:       ar.Note,
	}
	if ar.CheckIn != nil {
		s := ar.CheckIn.Format(time.RFC3339)
		resp.CheckIn = &s
	}
	if ar.CheckOut != nil {
		s := ar.CheckOut.Format(time.RFC3339)
		resp.CheckOut = &s
	}
	return resp
}

func mapAttendanceResponses(items []attendanceRepository.AttendanceRecord) []AttendanceResponse {
	out := make([]AttendanceResponse, len(items))
	for i, item := range items {
		out[i] = mapAttendanceResponse(item)
	}
	return out
}
