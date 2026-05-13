package service

import "errors"

var (
	ErrEmployeeNotFound      = errors.New("employee not found")
	ErrLeaveRequestNotFound  = errors.New("leave request not found")
	ErrInvalidEmployeeID     = errors.New("invalid employee id")
	ErrInvalidLeaveRequestID = errors.New("invalid leave request id")
	ErrInvalidDateFormat     = errors.New("invalid date format, use YYYY-MM-DD")
	ErrInvalidDateRange      = errors.New("end date must be after or equal to start date")
	ErrInvalidLeaveType      = errors.New("invalid leave type")
	ErrInvalidSkudEventType  = errors.New("invalid skud event type, use ENTER or EXIT")
	ErrDatabaseQueryFailed   = errors.New("failed to query database")
	ErrDatabaseSaveFailed    = errors.New("failed to save data")
)
