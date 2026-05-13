package service

import "errors"

var (
	ErrEmailOrPasswordEmpty = errors.New("email and password are required")
	ErrInvalidCredentials   = errors.New("invalid email or password")
	ErrInvalidRefreshToken  = errors.New("invalid or expired refresh token")
	ErrUserNotConfirmed     = errors.New("user account is not confirmed")
	ErrRefreshTokenRequired = errors.New("refresh token is required")
	ErrEmailRequired        = errors.New("email is required")
	ErrInvalidResetCode     = errors.New("invalid or incorrect reset code")
	ErrResetCodeExpired     = errors.New("reset code has expired")
	ErrPasswordTooWeak      = errors.New("password does not meet requirements")
)
