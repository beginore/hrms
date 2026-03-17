package service

import "errors"

var (
	ErrOrganizationIDRequired = errors.New("organization id is required")
	ErrInvalidOrganizationID  = errors.New("invalid organization id")
	ErrOrganizationNotFound   = errors.New("organization not found")
	ErrFirstNameRequired      = errors.New("first name is required")
	ErrLastNameRequired       = errors.New("last name is required")
	ErrEmailRequired          = errors.New("email is required")
	ErrInvalidEmail           = errors.New("email format is invalid")
	ErrInviteCodeRequired     = errors.New("invite code is required")
	ErrInviteNotFound         = errors.New("invite not found")
	ErrInviteExpired          = errors.New("invite has expired")
	ErrInviteAlreadyUsed      = errors.New("invite has already been used")
	ErrGenerateInvite         = errors.New("failed to generate unique invite code")
	ErrInviteEmailUnavailable = errors.New("invite email delivery is not configured")
	ErrPasswordRequired       = errors.New("password must be at least 8 characters and contain a special character")
	ErrPhoneNumberRequired    = errors.New("phone number is required")
	ErrInvalidPhoneNumber     = errors.New("phone number must be in international format, for example +77001234567")
	ErrPhoneNumberExists      = errors.New("phone number already exists")
	ErrEmailAlreadyExists     = errors.New("email already exists")
	ErrUserAlreadyExists      = errors.New("user already exists")
)

type PasswordPolicyError struct {
	Message string
}

func (e *PasswordPolicyError) Error() string {
	return e.Message
}
