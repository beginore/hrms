package service

import "errors"

var (
	ErrTaskNotFound         = errors.New("task not found")
	ErrTaskAlreadyReviewed  = errors.New("task has already been reviewed")
	ErrTaskNotSubmitted     = errors.New("task report has not been submitted yet")
	ErrTaskAlreadySubmitted = errors.New("task report has already been submitted")
	ErrNotAssignedEmployee  = errors.New("you are not the assigned employee for this task")
	ErrEmployeeNotFound     = errors.New("employee not found")
	ErrEmployeeNotInOrg     = errors.New("employee does not belong to this organisation")
	ErrForbidden            = errors.New("forbidden")
	ErrInvalidTaskID        = errors.New("invalid task id")
	ErrInvalidEmployeeID    = errors.New("invalid employee id")
	ErrInvalidDateFormat    = errors.New("invalid date format, use YYYY-MM-DD")
	ErrInvalidReviewAction  = errors.New("invalid action: use approve or reject")
	ErrDatabaseQueryFailed  = errors.New("failed to query database")
	ErrDatabaseSaveFailed   = errors.New("failed to save data")
)
