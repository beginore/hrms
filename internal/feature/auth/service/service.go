package service

import (
	"context"
	"errors"
	"hrms/internal/infrastructure/app/cognito"
)

type AuthService interface {
	Login(ctx context.Context, req LoginRequest) (*TokenResponse, error)
	RefreshToken(ctx context.Context, req RefreshTokenRequest) (*TokenResponse, error)
}

type authService struct {
	cognitoSvc *cognito.Service
}

func NewAuthService(cognitoSvc *cognito.Service) AuthService {
	return &authService{cognitoSvc: cognitoSvc}
}

func (s *authService) Login(ctx context.Context, req LoginRequest) (*TokenResponse, error) {
	if req.Email == "" || req.Password == "" {
		return nil, ErrEmailOrPasswordEmpty
	}
	output, err := s.cognitoSvc.SignIn(ctx, req.Email, req.Password)
	if err != nil {
		if errors.Is(err, cognito.ErrInvalidCredentials) {
			return nil, ErrUserNotConfirmed
		}
		if errors.Is(err, cognito.ErrUserNotConfirmed) {
			return nil, ErrUserNotConfirmed
		}
		return nil, err
	}
	result := output.AuthenticationResult
	return &TokenResponse{
		IdToken:      *result.IdToken,
		AccessToken:  *result.AccessToken,
		RefreshToken: *result.RefreshToken,
		ExpiresIn:    result.ExpiresIn,
		TokenType:    *result.TokenType,
	}, nil
}

func (s *authService) RefreshToken(ctx context.Context, req RefreshTokenRequest) (*TokenResponse, error) {
	if req.RefreshToken == "" {
		return nil, ErrInvalidRefreshToken
	}

	output, err := s.cognitoSvc.RefreshTokens()
}
