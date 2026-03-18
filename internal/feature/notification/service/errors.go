package service

import "errors"

var (
	ErrUserIDRequired         = errors.New("user_id is required")
	ErrInvalidUserID          = errors.New("invalid user_id")
	ErrInvalidOrgID           = errors.New("invalid org_id")
	ErrRoleRequired           = errors.New("role is required")
	ErrTitleRequired          = errors.New("title is required")
	ErrMessageRequired        = errors.New("message is required")
	ErrInvalidNotificationID  = errors.New("invalid notification_id")
	ErrNotificationNotFound   = errors.New("notification not found")
	ErrInvalidPaginationLimit = errors.New("limit must be between 1 and 100")
	ErrInvalidPaginationShift = errors.New("offset must be zero or greater")
)
