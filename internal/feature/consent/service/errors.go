package service

import "errors"

var (
	ErrNoActiveDocuments     = errors.New("no active documents found")
	ErrInvalidOrganizationID = errors.New("invalid organization id")
	ErrInvalidUserID         = errors.New("invalid user id")
)
