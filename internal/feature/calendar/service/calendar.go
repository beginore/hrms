package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	calendarRepository "hrms/internal/feature/calendar/repository"
	notificationService "hrms/internal/feature/notification/service"

	"github.com/google/uuid"
)

const dateLayout = "2006-01-02"

type CalendarService interface {
	AddDay(ctx context.Context, callerUserID uuid.UUID, req AddCalendarDayRequest) (*AddCalendarDayResponse, error)
	DeleteDay(ctx context.Context, callerUserID uuid.UUID, id uuid.UUID) error
	GetMonthSummary(ctx context.Context, callerUserID uuid.UUID, yearMonth string) (*MonthSummaryResponse, error)
}

type calendarService struct {
	repo      calendarRepository.CalendarRepository
	notifySvc *notificationService.Service
}

func NewCalendarService(repo calendarRepository.CalendarRepository, notifySvc *notificationService.Service) CalendarService {
	return &calendarService{repo: repo, notifySvc: notifySvc}
}

func (s *calendarService) AddDay(ctx context.Context, callerUserID uuid.UUID, req AddCalendarDayRequest) (*AddCalendarDayResponse, error) {
	if req.Type != "HOLIDAY" && req.Type != "WORKDAY" {
		return nil, ErrInvalidType
	}

	date, err := time.Parse(dateLayout, req.Date)
	if err != nil {
		return nil, ErrInvalidDate
	}

	log.Printf("[Calendar] AddDay date=%s type=%s", req.Date, req.Type)
	id, err := s.repo.AddDay(ctx, callerUserID, calendarRepository.AddCalendarDayParams{
		Date: date,
		Type: req.Type,
		Name: req.Name,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabaseSaveFailed, err)
	}
	s.notifyCalendarChanged(ctx, callerUserID, "Calendar day added", fmt.Sprintf("%s was marked as %s.", req.Date, req.Type), calendarMetadata(id, req.Date, req.Type, req.Name, "added"))
	return &AddCalendarDayResponse{ID: id.String()}, nil
}

func (s *calendarService) DeleteDay(ctx context.Context, callerUserID uuid.UUID, id uuid.UUID) error {
	exists, err := s.repo.DayExists(ctx, id)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrDatabaseQueryFailed, err)
	}
	if !exists {
		return ErrCalendarDayNotFound
	}

	log.Printf("[Calendar] DeleteDay id=%s", id)
	if err := s.repo.DeleteDay(ctx, callerUserID, id); err != nil {
		return fmt.Errorf("%w: %v", ErrDatabaseSaveFailed, err)
	}
	s.notifyCalendarChanged(ctx, callerUserID, "Calendar day removed", "A calendar override was removed.", calendarMetadata(id, "", "", "", "deleted"))
	return nil
}

func (s *calendarService) GetMonthSummary(ctx context.Context, callerUserID uuid.UUID, yearMonth string) (*MonthSummaryResponse, error) {
	var year, month int
	if _, err := fmt.Sscanf(yearMonth, "%d-%d", &year, &month); err != nil || month < 1 || month > 12 {
		return nil, ErrInvalidMonth
	}

	days, err := s.repo.GetByMonth(ctx, callerUserID, year, month)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabaseQueryFailed, err)
	}

	// index explicit entries by date string for fast lookup
	overrides := make(map[string]calendarRepository.CalendarDay, len(days))
	for _, d := range days {
		overrides[d.Date.Format(dateLayout)] = d
	}

	orgID, err := s.repo.GetOrgIDByUserID(ctx, callerUserID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDatabaseQueryFailed, err)
	}

	from := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	totalDays := from.AddDate(0, 1, 0).AddDate(0, 0, -1).Day()

	summary := &MonthSummaryResponse{
		Year:  year,
		Month: month,
		Days:  make([]DaySummary, 0, totalDays),
	}

	for day := 1; day <= totalDays; day++ {
		date := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
		dateStr := date.Format(dateLayout)

		ds := DaySummary{
			Date:    dateStr,
			Weekday: date.Weekday().String(),
		}

		if entry, ok := overrides[dateStr]; ok {
			ds.Name = entry.Name
			if entry.Type == "HOLIDAY" {
				ds.Type = "HOLIDAY"
				ds.IsWorking = false
				summary.Holidays++
			} else {
				ds.Type = "WORKDAY"
				ds.IsWorking = true
				summary.WorkdayOverrides++
			}
		} else if date.Weekday() == time.Saturday || date.Weekday() == time.Sunday {
			ds.Type = "WEEKEND"
			ds.IsWorking = false
			summary.Weekends++
		} else {
			ds.Type = "REGULAR"
			ds.IsWorking = true
		}

		if ds.IsWorking {
			summary.WorkingDays++
		}
		summary.Days = append(summary.Days, ds)
	}

	summary.TotalDays = totalDays
	_ = orgID // used implicitly via repo.GetByMonth
	return summary, nil
}

func (s *calendarService) notifyCalendarChanged(ctx context.Context, callerUserID uuid.UUID, title, message string, metadata json.RawMessage) {
	if s.notifySvc == nil {
		return
	}

	orgID, err := s.repo.GetOrgIDByUserID(ctx, callerUserID)
	if err != nil {
		log.Printf("[Calendar] resolve org for calendar notification failed: %v", err)
		return
	}

	orgIDRaw := orgID.String()
	for _, role := range []string{notificationService.RoleAdmin, notificationService.RoleEmployee} {
		if _, err := s.notifySvc.NotifySystemToRole(ctx, notificationService.NotifyRoleRequest{
			OrgID:    &orgIDRaw,
			Role:     role,
			Title:    title,
			Message:  message,
			Metadata: metadata,
		}); err != nil {
			log.Printf("[Calendar] notify %s about calendar change failed: %v", role, err)
		}
	}
}

func calendarMetadata(id uuid.UUID, date, dayType, name, action string) json.RawMessage {
	data := map[string]string{
		"calendarDayId": id.String(),
		"action":        action,
	}
	if date != "" {
		data["date"] = date
	}
	if dayType != "" {
		data["type"] = dayType
	}
	if name != "" {
		data["name"] = name
	}

	raw, err := json.Marshal(data)
	if err != nil {
		return json.RawMessage(`{}`)
	}
	return raw
}
