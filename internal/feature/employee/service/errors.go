package service

import "errors"

var (
	ErrEmployeeNotFound    = errors.New("employee not found")
	ErrInvalidEmployeeID   = errors.New("invalid employee id")
	ErrInvalidUserID       = errors.New("invalid user id")
	ErrInvalidDepartmentID = errors.New("invalid department id")
	ErrInvalidPositionID   = errors.New("invalid position id")
	ErrDatabaseQueryFailed = errors.New("failed to query database")
	ErrDatabaseSaveFailed  = errors.New("failed to save data")
)
