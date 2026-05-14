package service

import "errors"

var (
	ErrInvalidEventID      = errors.New("invalid event id")
	ErrEventNotFound       = errors.New("event not found")
	ErrTitleRequired       = errors.New("title is required")
	ErrDescriptionRequired = errors.New("description is required")
	ErrScopeRequired       = errors.New("scope is required")
	ErrInvalidScope        = errors.New("scope must be global or department")
	ErrDepartmentRequired  = errors.New("department_id is required for department scope")
	ErrDepartmentForbidden = errors.New("department_id must be empty for global scope")
	ErrInvalidDepartmentID = errors.New("invalid department_id")
	ErrStartsAtRequired    = errors.New("startsAt is required")
	ErrEndsAtRequired      = errors.New("endsAt is required")
	ErrInvalidStartsAt     = errors.New("invalid startsAt")
	ErrInvalidEndsAt       = errors.New("invalid endsAt")
	ErrInvalidEventTime    = errors.New("endsAt must be after startsAt")
	ErrPermissionDenied    = errors.New("permission denied")
	ErrDepartmentNotBound  = errors.New("user is not assigned to a department")
	ErrRoleUnavailable     = errors.New("user role is not available")
)
