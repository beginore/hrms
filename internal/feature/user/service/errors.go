package service

import "errors"

var (
	ErrInvalidAccessToken = errors.New("invalid or expired access token")
	ErrUserNotFound       = errors.New("user not found")
)
