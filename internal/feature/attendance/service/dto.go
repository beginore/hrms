package service

// ─── Work Schedule DTOs ───────────────────────────────────────────────────────

type SetWorkScheduleRequest struct {
	EmployeeID           string `json:"employee_id"           binding:"required"`
	WorkStart            string `json:"work_start"            binding:"required"` // "09:00"
	WorkEnd              string `json:"work_end"              binding:"required"` // "18:00"
	LateThresholdMinutes int32  `json:"late_threshold_minutes"`                   // default 15
}

type WorkScheduleResponse struct {
	EmployeeID           string `json:"employee_id"`
	WorkStart            string `json:"work_start"`
	WorkEnd              string `json:"work_end"`
	LateThresholdMinutes int32  `json:"late_threshold_minutes"`
}

// ─── SKUD DTOs ────────────────────────────────────────────────────────────────

// CreateSkudEventRequest is the payload received from the СКУД system (or simulator).
type CreateSkudEventRequest struct {
	EmployeeID string `json:"employee_id" binding:"required"`
	EventType  string `json:"event_type"  binding:"required"` // ENTER | EXIT
	DeviceID   string `json:"device_id"`
	OccurredAt string `json:"occurred_at"` // RFC3339; empty = now
}

type CreateSkudEventResponse struct {
	Message string `json:"message"`
}

type CheckInRequest struct {
	WorkType string `json:"work_type"`
	Note     string `json:"note"`
}

type CheckInResponse struct {
	Status  string `json:"status"`
	CheckIn string `json:"check_in"`
}

type CheckOutResponse struct {
	CheckOut string `json:"check_out"`
}

// ─── Leave Request DTOs ───────────────────────────────────────────────────────

// CreateLeaveRequest is submitted by the caller for themselves.
type CreateLeaveRequest struct {
	EmployeeID  string `json:"employee_id,omitempty"`
	Type        string `json:"type"         binding:"required"` // SICK_LEAVE | VACATION | REMOTE | BUSINESS_TRIP | UNPAID_LEAVE
	StartDate   string `json:"start_date"   binding:"required"` // YYYY-MM-DD
	EndDate     string `json:"end_date"     binding:"required"` // YYYY-MM-DD
	Reason      string `json:"reason"`
	DocumentURL string `json:"document_url"`
}

type CreateLeaveResponse struct {
	LeaveRequestID string `json:"leave_request_id"`
}

type LeaveRequestResponse struct {
	ID          string  `json:"id"`
	EmployeeID  string  `json:"employee_id"`
	Type        string  `json:"type"`
	StartDate   string  `json:"start_date"`
	EndDate     string  `json:"end_date"`
	Reason      string  `json:"reason,omitempty"`
	DocumentURL string  `json:"document_url,omitempty"`
	Status      string  `json:"status"`
	ReviewedBy  *string `json:"reviewed_by,omitempty"`
	ReviewedAt  *string `json:"reviewed_at,omitempty"`
	CreatedAt   string  `json:"created_at"`
}

type ReviewLeaveRequest struct {
	Action string `json:"action" binding:"required"` // approve | reject
}

// ─── Attendance DTOs ──────────────────────────────────────────────────────────

type AttendanceResponse struct {
	ID         string  `json:"id"`
	EmployeeID string  `json:"employee_id"`
	Date       string  `json:"date"`
	Type       string  `json:"type"`
	Source     string  `json:"source"`
	Status     string  `json:"status"`
	CheckIn    *string `json:"check_in,omitempty"`
	CheckOut   *string `json:"check_out,omitempty"`
	Note       string  `json:"note,omitempty"`
}
