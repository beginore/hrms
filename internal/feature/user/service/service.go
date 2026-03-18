package service

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"strings"
	"time"

	userRepository "hrms/internal/feature/user/repository"
	"hrms/internal/infrastructure/app/cognito"
)

type Service struct {
	cognitoSvc *cognito.Service
	repo       *userRepository.Repository
}

func NewService(cognitoSvc *cognito.Service, repo *userRepository.Repository) *Service {
	return &Service{
		cognitoSvc: cognitoSvc,
		repo:       repo,
	}
}

func (s *Service) GetMe(ctx context.Context, accessToken string) (*UserProfileResponse, error) {
	if strings.TrimSpace(accessToken) == "" {
		return nil, ErrInvalidAccessToken
	}

	userID, _, err := s.cognitoSvc.ParseTokenClaims(accessToken)
	if err != nil {
		log.Printf("[User Profile] Failed to parse access token claims: %v", err)
		return nil, ErrInvalidAccessToken
	}

	profile, err := s.repo.GetUserProfileByID(ctx, userID)
	if err != nil {
		if errors.Is(err, userRepository.ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		log.Printf("[User Profile] Failed to load profile for userID=%s: %v", userID, err)
		return nil, err
	}

	joinedDate := time.Time{}
	if profile.JoinedDate.Valid {
		joinedDate = profile.JoinedDate.Time
	}

	return &UserProfileResponse{
		ID:                 profile.ID.String(),
		OrganizationID:     profile.OrganizationID.String(),
		OrganizationName:   profile.OrganizationName,
		Email:              profile.Email,
		Firstname:          profile.Firstname,
		Lastname:           profile.Lastname,
		FullName:           strings.TrimSpace(profile.Firstname + " " + profile.Lastname),
		Role:               profile.Role,
		Phone:              profile.PhoneNumber,
		PhoneNumber:        profile.PhoneNumber,
		VerificationStatus: profile.VerificationStatus,
		JoinedDate:         joinedDate,
		Department:         nullStringValue(profile.Department),
		DepartmentID:       nullStringValue(profile.DepartmentID),
		Position:           nullStringValue(profile.Position),
		Salary:             nullStringValue(profile.Salary),
		Location:           nullStringValue(profile.Location),
	}, nil
}

func nullStringValue(value sql.NullString) string {
	if value.Valid {
		return value.String
	}

	return ""
}
