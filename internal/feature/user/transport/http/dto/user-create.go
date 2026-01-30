package dto

import domain "hrms/internal/domain/user"

// dto.go can be changed to dto directory to store DTOs for different handlers with their validation in different files

// JSON tags and validations are handled within DTOs
// If you need very strong validation, you can create Validate method on DTO struct and use some external package for validation
// Example tag is for Swagger docs, also may be optional based on what is used for documenting API

type UserCreateRequest struct {
	FirstName string `example:"Jansaya" json:"firstName"   validate:"required"`
	LastName  string `example:"Adel" json:"lastName"   validate:"required"`
	Email     string `example:"strings" json:"email"  validate:"required,email"`
}

func (r *UserCreateRequest) ToDomain() domain.User {
	return domain.User{
		FirstName: r.FirstName,
		LastName:  r.LastName,
		Email:     r.Email,
	}
}

type UserCreateResponse struct {
	Message string `example:"User was created successfully" json:"message"`
}
