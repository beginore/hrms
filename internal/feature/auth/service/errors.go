package service

import "errors"

var (
	ErrEmailOrPasswordEmpty = errors.New("email and password are required")
	ErrInvalidCredentials   = errors.New("invalid email or password")
	ErrInvalidRefreshToken  = errors.New("invalid or expired refresh token")
	ErrUserNotConfirmed     = errors.New("user account is not confirmed")
)
