package service

import "errors"

var (
	ErrCalendarDayNotFound = errors.New("calendar day not found")
	ErrInvalidDate         = errors.New("invalid date format, use YYYY-MM-DD")
	ErrInvalidType         = errors.New("invalid type, use HOLIDAY or WORKDAY")
	ErrInvalidMonth        = errors.New("invalid month format, use YYYY-MM")
	ErrDatabaseQueryFailed = errors.New("failed to query database")
	ErrDatabaseSaveFailed  = errors.New("failed to save data")
)
