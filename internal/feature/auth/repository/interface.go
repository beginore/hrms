package repository

import (
	"context"
	"hrms/internal/feature/auth/repository/postgres"

	"github.com/google/uuid"
)

type AuthRepository interface {
	DeleteSessionByRefreshToken(ctx context.Context, refreshToken string) error
	DeleteSessionsByUserID(ctx context.Context, userID uuid.UUID) error
	GetSessionByRefreshToken(ctx context.Context, refreshToken string) (postgres.GetSessionByRefreshTokenRow, error)
	InsertUserSession(ctx context.Context, arg postgres.InsertUserSessionParams) error
}
