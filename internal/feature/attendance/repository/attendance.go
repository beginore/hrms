package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	attendancePostgres "hrms/internal/feature/attendance/repository/postgres"
	calendarRepository "hrms/internal/feature/calendar/repository"

	"github.com/google/uuid"
)

type attendanceRepository struct {
	queries  attendancePostgres.Querier
	calendar calendarRepository.CalendarRepository
}

func NewRepository(conn *sql.DB, calRepo calendarRepository.CalendarRepository) AttendanceRepository {
	return &attendanceRepository{queries: attendancePostgres.New(conn), calendar: calRepo}
}

// ─── Work Schedules ───────────────────────────────────────────────────────────

func (r *attendanceRepository) SetWorkSchedule(ctx context.Context, callerUserID uuid.UUID, params SetWorkScheduleParams) error {
	orgID, err := r.GetOrgIDByUserID(ctx, callerUserID)
	if err != nil {
		return fmt.Errorf("resolve org: %w", err)
	}
	return r.queries.UpsertWorkSchedule(ctx, attendancePostgres.UpsertWorkScheduleParams{
		ID:                   uuid.New(),
		OrgID:                orgID,
		EmployeeID:           params.EmployeeID,
		WorkStart:            params.WorkStart,
		WorkEnd:              params.WorkEnd,
		LateThresholdMinutes: params.LateThresholdMinutes,
	})
}

func (r *attendanceRepository) GetWorkSchedule(ctx context.Context, employeeID uuid.UUID) (WorkSchedule, error) {
	ws, err := r.queries.GetWorkScheduleByEmployeeID(ctx, employeeID)
	if err != nil {
		return WorkSchedule{}, err
	}
	return WorkSchedule{
		ID:                   ws.ID,
		OrgID:                ws.OrgID,
		EmployeeID:           ws.EmployeeID,
		WorkStart:            ws.WorkStart,
		WorkEnd:              ws.WorkEnd,
		LateThresholdMinutes: ws.LateThresholdMinutes,
	}, nil
}

func (r *attendanceRepository) GetWorkScheduleOrDefault(ctx context.Context, employeeID uuid.UUID) WorkSchedule {
	ws, err := r.GetWorkSchedule(ctx, employeeID)
	if err != nil {
		return WorkSchedule{
			EmployeeID:           employeeID,
			WorkStart:            "09:00:00",
			WorkEnd:              "18:00:00",
			LateThresholdMinutes: 15,
		}
	}
	return ws
}

// ─── SKUD Events ──────────────────────────────────────────────────────────────

func (r *attendanceRepository) CreateSkudEvent(ctx context.Context, callerUserID uuid.UUID, params CreateSkudEventParams) error {
	orgID, err := r.GetOrgIDByEmployeeID(ctx, params.EmployeeID)
	if err != nil {
		return fmt.Errorf("resolve org: %w", err)
	}
	if err := r.queries.InsertSkudEvent(ctx, attendancePostgres.InsertSkudEventParams{
		ID:         uuid.New(),
		OrgID:      orgID,
		EmployeeID: params.EmployeeID,
		EventType:  params.EventType,
		DeviceID:   params.DeviceID,
		OccurredAt: params.OccurredAt,
	}); err != nil {
		return err
	}
	return r.ProcessSkudEvents(ctx, orgID)
}

func (r *attendanceRepository) ProcessSkudEvents(ctx context.Context, orgID uuid.UUID) error {
	events, err := r.queries.GetUnprocessedSkudEvents(ctx, orgID)
	if err != nil {
		return err
	}

	for _, ev := range events {
		if err := r.processOneSkudEvent(ctx, ev); err != nil {
			continue
		}
		_ = r.queries.MarkSkudEventProcessed(ctx, ev.ID)
	}
	return nil
}

func (r *attendanceRepository) processOneSkudEvent(ctx context.Context, ev attendancePostgres.SkudEvent) error {
	date := ev.OccurredAt.Truncate(24 * time.Hour)

	if r.calendar != nil {
		nonWorking, err := r.calendar.IsNonWorkingDay(ctx, ev.OrgID, date)
		if err != nil {
			return err
		}
		if nonWorking {
			return r.queries.UpsertAttendanceRecord(ctx, attendancePostgres.UpsertAttendanceRecordParams{
				ID:         uuid.New(),
				OrgID:      ev.OrgID,
				EmployeeID: ev.EmployeeID,
				Date:       date,
				Type:       "HOLIDAY",
				Source:     "SYSTEM",
				Status:     "ON_LEAVE",
				Note:       "Non-working day",
			})
		}
	}

	leave, hasLeave, err := r.GetApprovedLeaveForDate(ctx, ev.EmployeeID, date)
	if err != nil {
		return err
	}
	if hasLeave {
		return r.queries.UpsertAttendanceRecord(ctx, attendancePostgres.UpsertAttendanceRecordParams{
			ID:         uuid.New(),
			OrgID:      ev.OrgID,
			EmployeeID: ev.EmployeeID,
			Date:       date,
			Type:       leave.Type,
			Source:     "SYSTEM",
			Status:     "ON_LEAVE",
		})
	}

	existing, existErr := r.queries.GetAttendanceByEmployeeAndDate(ctx, attendancePostgres.GetAttendanceByEmployeeAndDateParams{
		EmployeeID: ev.EmployeeID,
		Date:       date,
	})

	if ev.EventType == "EXIT" {
		if existErr == nil {
			return r.queries.UpdateCheckOut(ctx, ev.EmployeeID, date, ev.OccurredAt)
		}
		return r.queries.UpsertAttendanceRecord(ctx, attendancePostgres.UpsertAttendanceRecordParams{
			ID:         uuid.New(),
			OrgID:      ev.OrgID,
			EmployeeID: ev.EmployeeID,
			Date:       date,
			Type:       "OFFICE",
			Source:     "SKUD",
			Status:     "PRESENT",
			CheckOut:   &ev.OccurredAt,
		})
	}

	ws := r.GetWorkScheduleOrDefault(ctx, ev.EmployeeID)
	status, attendanceType := r.resolveEnterStatus(ev, ws)

	params := attendancePostgres.UpsertAttendanceRecordParams{
		ID:         uuid.New(),
		OrgID:      ev.OrgID,
		EmployeeID: ev.EmployeeID,
		Date:       date,
		Type:       attendanceType,
		Source:     "SKUD",
		Status:     status,
		CheckIn:    &ev.OccurredAt,
	}

	if existErr == nil && existing.CheckOut.Valid {
		params.CheckOut = &existing.CheckOut.Time
	}

	return r.queries.UpsertAttendanceRecord(ctx, params)
}

func (r *attendanceRepository) resolveEnterStatus(ev attendancePostgres.SkudEvent, ws WorkSchedule) (string, string) {
	workStartTime, err := time.Parse("15:04:05", ws.WorkStart)
	if err != nil {
		return "PRESENT", "OFFICE"
	}

	loc := ev.OccurredAt.Location()
	deadline := time.Date(
		ev.OccurredAt.Year(), ev.OccurredAt.Month(), ev.OccurredAt.Day(),
		workStartTime.Hour(), workStartTime.Minute(), workStartTime.Second(), 0, loc,
	).Add(time.Duration(ws.LateThresholdMinutes) * time.Minute)

	if ev.OccurredAt.After(deadline) {
		return "LATE", "OFFICE"
	}
	return "PRESENT", "OFFICE"
}

// ─── Leave Requests ───────────────────────────────────────────────────────────

func (r *attendanceRepository) ManualCheckIn(ctx context.Context, params ManualCheckInParams) error {
	return r.queries.UpsertAttendanceRecord(ctx, attendancePostgres.UpsertAttendanceRecordParams{
		ID:         uuid.New(),
		OrgID:      params.OrgID,
		EmployeeID: params.EmployeeID,
		Date:       params.CheckIn.Truncate(24 * time.Hour),
		Type:       params.WorkType,
		Source:     "MANUAL",
		Status:     params.Status,
		CheckIn:    &params.CheckIn,
		Note:       params.Note,
	})
}

func (r *attendanceRepository) ManualCheckOut(ctx context.Context, employeeID uuid.UUID) error {
	now := time.Now()
	date := now.Truncate(24 * time.Hour)
	return r.queries.UpdateCheckOut(ctx, employeeID, date, now)
}

func (r *attendanceRepository) GetTodayAttendance(ctx context.Context, employeeID uuid.UUID) (AttendanceRecord, error) {
	today := time.Now().Truncate(24 * time.Hour)
	record, err := r.queries.GetAttendanceByEmployeeAndDate(ctx, attendancePostgres.GetAttendanceByEmployeeAndDateParams{
		EmployeeID: employeeID,
		Date:       today,
	})
	if err != nil {
		return AttendanceRecord{}, err
	}
	return mapAttendanceRecord(record), nil
}

func (r *attendanceRepository) CreateLeaveRequest(ctx context.Context, callerUserID uuid.UUID, params CreateLeaveRequestParams) (uuid.UUID, error) {
	orgID, err := r.GetOrgIDByUserID(ctx, callerUserID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("resolve org: %w", err)
	}
	id := uuid.New()
	if err := r.queries.InsertLeaveRequest(ctx, attendancePostgres.InsertLeaveRequestParams{
		ID:          id,
		OrgID:       orgID,
		EmployeeID:  params.EmployeeID,
		Type:        params.Type,
		StartDate:   params.StartDate,
		EndDate:     params.EndDate,
		Reason:      params.Reason,
		DocumentURL: params.DocumentURL,
	}); err != nil {
		return uuid.Nil, err
	}
	return id, nil
}

func (r *attendanceRepository) GetLeaveRequestByID(ctx context.Context, id uuid.UUID) (LeaveRequest, error) {
	lr, err := r.queries.GetLeaveRequestByID(ctx, id)
	if err != nil {
		return LeaveRequest{}, err
	}
	return mapLeaveRequest(lr), nil
}

func (r *attendanceRepository) GetLeaveRequestsByOrg(ctx context.Context, callerUserID uuid.UUID) ([]LeaveRequest, error) {
	orgID, err := r.GetOrgIDByUserID(ctx, callerUserID)
	if err != nil {
		return nil, err
	}
	rows, err := r.queries.GetLeaveRequestsByOrgID(ctx, orgID)
	if err != nil {
		return nil, err
	}
	return mapLeaveRequests(rows), nil
}

func (r *attendanceRepository) GetLeaveRequestsByEmployee(ctx context.Context, employeeID uuid.UUID) ([]LeaveRequest, error) {
	rows, err := r.queries.GetLeaveRequestsByEmployeeID(ctx, employeeID)
	if err != nil {
		return nil, err
	}
	return mapLeaveRequests(rows), nil
}

func (r *attendanceRepository) ReviewLeaveRequest(ctx context.Context, callerUserID uuid.UUID, params ReviewLeaveRequestParams) error {
	return r.queries.UpdateLeaveRequestStatus(ctx, attendancePostgres.UpdateLeaveRequestStatusParams{
		ID:         params.ID,
		Status:     params.Status,
		ReviewedBy: callerUserID,
	})
}

// ─── Attendance Records ───────────────────────────────────────────────────────

func (r *attendanceRepository) UpsertAttendance(ctx context.Context, callerUserID uuid.UUID, params UpsertAttendanceParams) error {
	orgID, err := r.GetOrgIDByUserID(ctx, callerUserID)
	if err != nil {
		return fmt.Errorf("resolve org: %w", err)
	}
	return r.queries.UpsertAttendanceRecord(ctx, attendancePostgres.UpsertAttendanceRecordParams{
		ID:         uuid.New(),
		OrgID:      orgID,
		EmployeeID: params.EmployeeID,
		Date:       params.Date,
		Type:       params.Type,
		Source:     params.Source,
		Status:     params.Status,
		CheckIn:    params.CheckIn,
		CheckOut:   params.CheckOut,
		Note:       params.Note,
	})
}

func (r *attendanceRepository) GetAttendanceByOrg(ctx context.Context, callerUserID uuid.UUID, period GetByPeriodParams) ([]AttendanceRecord, error) {
	orgID, err := r.GetOrgIDByUserID(ctx, callerUserID)
	if err != nil {
		return nil, err
	}
	rows, err := r.queries.GetAttendanceByOrgAndPeriod(ctx, attendancePostgres.GetAttendanceByOrgAndPeriodParams{
		OrgID:     orgID,
		StartDate: period.StartDate,
		EndDate:   period.EndDate,
	})
	if err != nil {
		return nil, err
	}
	return mapAttendanceRecords(rows), nil
}

func (r *attendanceRepository) GetAttendanceByEmployee(ctx context.Context, employeeID uuid.UUID, period GetByPeriodParams) ([]AttendanceRecord, error) {
	rows, err := r.queries.GetAttendanceByEmployeeAndPeriod(ctx, attendancePostgres.GetAttendanceByEmployeeAndPeriodParams{
		EmployeeID: employeeID,
		StartDate:  period.StartDate,
		EndDate:    period.EndDate,
	})
	if err != nil {
		return nil, err
	}
	return mapAttendanceRecords(rows), nil
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

func (r *attendanceRepository) GetOrgIDByUserID(ctx context.Context, userID uuid.UUID) (uuid.UUID, error) {
	return r.queries.GetOrgIDByUserID(ctx, userID)
}

func (r *attendanceRepository) GetEmployeeIDByUserID(ctx context.Context, userID uuid.UUID) (uuid.UUID, error) {
	return r.queries.GetEmployeeIDByUserID(ctx, userID)
}

func (r *attendanceRepository) GetOrgIDByEmployeeID(ctx context.Context, employeeID uuid.UUID) (uuid.UUID, error) {
	return r.queries.GetOrgIDByEmployeeID(ctx, employeeID)
}

func (r *attendanceRepository) EmployeeExists(ctx context.Context, employeeID uuid.UUID) (bool, error) {
	count, err := r.queries.CheckEmployeeExists(ctx, employeeID)
	return count > 0, err
}

func (r *attendanceRepository) LeaveRequestExists(ctx context.Context, id uuid.UUID) (bool, error) {
	count, err := r.queries.CheckLeaveRequestExists(ctx, id)
	return count > 0, err
}

func (r *attendanceRepository) GetApprovedLeaveForDate(ctx context.Context, employeeID uuid.UUID, date time.Time) (LeaveRequest, bool, error) {
	lr, err := r.queries.GetApprovedLeaveForDate(ctx, attendancePostgres.GetApprovedLeaveForDateParams{
		EmployeeID: employeeID,
		Date:       date,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return LeaveRequest{}, false, nil
	}
	if err != nil {
		return LeaveRequest{}, false, err
	}
	return mapLeaveRequest(lr), true, nil
}

// ─── Mappers ──────────────────────────────────────────────────────────────────

func mapLeaveRequest(lr attendancePostgres.LeaveRequest) LeaveRequest {
	out := LeaveRequest{
		ID:         lr.ID,
		OrgID:      lr.OrgID,
		EmployeeID: lr.EmployeeID,
		Type:       lr.Type,
		StartDate:  lr.StartDate,
		EndDate:    lr.EndDate,
		Status:     lr.Status,
		CreatedAt:  lr.CreatedAt,
	}
	if lr.Reason.Valid {
		out.Reason = lr.Reason.String
	}
	if lr.DocumentURL.Valid {
		out.DocumentURL = lr.DocumentURL.String
	}
	if lr.ReviewedBy.Valid {
		id := lr.ReviewedBy.UUID
		out.ReviewedBy = &id
	}
	if lr.ReviewedAt.Valid {
		t := lr.ReviewedAt.Time
		out.ReviewedAt = &t
	}
	return out
}

func mapLeaveRequests(rows []attendancePostgres.LeaveRequest) []LeaveRequest {
	out := make([]LeaveRequest, len(rows))
	for i, r := range rows {
		out[i] = mapLeaveRequest(r)
	}
	return out
}

func mapAttendanceRecord(ar attendancePostgres.AttendanceRecord) AttendanceRecord {
	out := AttendanceRecord{
		ID:         ar.ID,
		OrgID:      ar.OrgID,
		EmployeeID: ar.EmployeeID,
		Date:       ar.Date,
		Type:       ar.Type,
		Source:     ar.Source,
		Status:     ar.Status,
	}
	if ar.CheckIn.Valid {
		t := ar.CheckIn.Time
		out.CheckIn = &t
	}
	if ar.CheckOut.Valid {
		t := ar.CheckOut.Time
		out.CheckOut = &t
	}
	if ar.Note.Valid {
		out.Note = ar.Note.String
	}
	return out
}

func mapAttendanceRecords(rows []attendancePostgres.AttendanceRecord) []AttendanceRecord {
	out := make([]AttendanceRecord, len(rows))
	for i, r := range rows {
		out[i] = mapAttendanceRecord(r)
	}
	return out
}
