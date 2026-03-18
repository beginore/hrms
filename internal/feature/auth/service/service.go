package service

import (
	"context"
	"errors"
	"hrms/internal/feature/auth/repository"
	authPostgres "hrms/internal/feature/auth/repository/postgres"
	"hrms/internal/infrastructure/app/cognito"
	"log"

	"github.com/google/uuid"
)

type AuthService interface {
	Login(ctx context.Context, req LoginRequest) (*TokenResponse, error)
	RefreshToken(ctx context.Context, req RefreshTokenRequest) (*TokenResponse, error)
}

type authService struct {
	cognitoSvc *cognito.Service
	repo       repository.AuthRepository
}

func NewAuthService(cognitoSvc *cognito.Service, repo repository.AuthRepository) AuthService {
	return &authService{cognitoSvc: cognitoSvc, repo: repo}
}

func (s *authService) Login(ctx context.Context, req LoginRequest) (*TokenResponse, error) {
	if req.Email == "" || req.Password == "" {
		return nil, ErrEmailOrPasswordEmpty
	}
	output, err := s.cognitoSvc.SignIn(ctx, req.Email, req.Password)
	if err != nil {
		if errors.Is(err, cognito.ErrInvalidCredentials) {
			return nil, ErrInvalidCredentials
		}
		if errors.Is(err, cognito.ErrUserNotConfirmed) {
			return nil, ErrUserNotConfirmed
		}
		return nil, err
	}
	result := output.AuthenticationResult
	userId, cognitoUsername, err := s.cognitoSvc.ParseTokenClaims(*result.AccessToken)
	if err != nil {
		log.Printf("[Auth] Failed to parse token claims %v", err)
		return nil, err
	}
	if err := s.repo.InsertUserSession(ctx, authPostgres.InsertUserSessionParams{
		ID:              uuid.New(),
		UserID:          userId,
		CognitoUsername: cognitoUsername,
		RefreshToken:    *result.RefreshToken,
		ExpiresAt:       repository.SessionExpiresAt(),
	}); err != nil {
		log.Printf("[Auth] Failed to save session %v", err)
	}
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

	session, err := s.repo.GetSessionByRefreshToken(ctx, req.RefreshToken)
	if err != nil {
		log.Printf("[Auth] Failed to get session %v", err)
		return nil, ErrInvalidRefreshToken
	}

	output, err := s.cognitoSvc.RefreshToken(ctx, req.RefreshToken, session.CognitoUsername)
	if err != nil {
		if errors.Is(err, cognito.ErrInvalidRefreshToken) {
			_ = s.repo.DeleteSessionByRefreshToken(ctx, session.RefreshToken)
			return nil, ErrInvalidRefreshToken
		}
		return nil, err
	}
	result := output.AuthenticationResult
	// Since we have disabled refresh token rotation in our cognito, it will just return is nil, so we just use our current refresh token for now ;D
	refreshToken := req.RefreshToken
	if result.RefreshToken != nil {
		refreshToken = *result.RefreshToken
	}
	return &TokenResponse{
		IdToken:      *result.IdToken,
		AccessToken:  *result.AccessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    result.ExpiresIn,
		TokenType:    *result.TokenType,
	}, nil
}
