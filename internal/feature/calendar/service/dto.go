package service

type AddCalendarDayRequest struct {
	Date string `json:"date" binding:"required"` // YYYY-MM-DD
	Type string `json:"type" binding:"required"` // HOLIDAY | WORKDAY
	Name string `json:"name"`                    // "Наурыз", "День Независимости", etc.
}

type AddCalendarDayResponse struct {
	ID string `json:"id"`
}

type CalendarDayResponse struct {
	ID   string `json:"id"`
	Date string `json:"date"`
	Type string `json:"type"`
	Name string `json:"name,omitempty"`
}

// MonthSummaryResponse gives a full picture of a calendar month.
type MonthSummaryResponse struct {
	Year         int                   `json:"year"`
	Month        int                   `json:"month"`
	TotalDays    int                   `json:"total_days"`
	WorkingDays  int                   `json:"working_days"`
	Weekends     int                   `json:"weekends"`
	Holidays     int                   `json:"holidays"`
	WorkdayOverrides int               `json:"workday_overrides"`
	Days         []DaySummary          `json:"days"`
}

type DaySummary struct {
	Date      string `json:"date"`
	Weekday   string `json:"weekday"`
	IsWorking bool   `json:"is_working"`
	Type      string `json:"type,omitempty"` // HOLIDAY | WORKDAY | WEEKEND | REGULAR
	Name      string `json:"name,omitempty"`
}
