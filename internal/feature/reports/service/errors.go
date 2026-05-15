package service

import "errors"

var (
	ErrForbidden           = errors.New("forbidden")
	ErrInvalidDateFormat   = errors.New("invalid date format")
	ErrInvalidDateRange    = errors.New("invalid date range")
	ErrInvalidReportType   = errors.New("invalid report type")
	ErrInvalidExportFormat = errors.New("invalid export format")
	ErrDatabaseQueryFailed = errors.New("database query failed")
)
