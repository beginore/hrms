package domain

import (
	"hrms/internal/infrastructure/errs"
)

// The struct already contains aggregated data from different User models that represent tabel entity
// No JSON, nor Validation, nor DB tags should be there.

type User struct {
	// From users table
	ID    uint64
	Email string
	// From user_infos table
	FirstName string
	LastName  string
	// From additional computations
	PhotoURL *string
}

func NewUserAlreadyExistsError() *errs.Error {
	return errs.New(errs.CodeAlreadyExists, "User already exists")
}
