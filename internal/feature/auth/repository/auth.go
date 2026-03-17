package repository

import (
	"context"
	"database/sql"
	"errors"
	"hrms/internal/feature/auth/repository/postgres"
	"time"

	"github.com/google/uuid"
)

type authRepository struct {
	queries postgres.Querier
}

var ErrSessionNotFound = errors.New("session not found or expired")

func NewAuthRepository(conn *sql.DB) AuthRepository {
	return &authRepository{queries: postgres.New(conn)}
}

func (r *authRepository) InsertUserSession(ctx context.Context, arg postgres.InsertUserSessionParams) error {
	return r.queries.InsertUserSession(ctx, postgres.InsertUserSessionParams{
		ID:              arg.ID,
		UserID:          arg.UserID,
		CognitoUsername: arg.CognitoUsername,
		RefreshToken:    arg.RefreshToken,
		ExpiresAt:       arg.ExpiresAt,
	})
}

func (r *authRepository) GetSessionByRefreshToken(ctx context.Context, refreshToken string) (postgres.GetSessionByRefreshTokenRow, error) {
	row, err := r.queries.GetSessionByRefreshToken(ctx, refreshToken)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return postgres.GetSessionByRefreshTokenRow{}, ErrSessionNotFound
		}
		return postgres.GetSessionByRefreshTokenRow{}, err
	}
	return row, nil
}

func (r *authRepository) DeleteSessionByRefreshToken(ctx context.Context, refreshToken string) error {
	return r.queries.DeleteSessionByRefreshToken(ctx, refreshToken)
}

func (r *authRepository) DeleteSessionsByUserID(ctx context.Context, userID uuid.UUID) error {
	return r.queries.DeleteSessionsByUserID(ctx, userID)
}

const sessionExpiresIn = 30 * 24 * time.Hour

func SessionExpiresAt() time.Time {
	return time.Now().Add(sessionExpiresIn)
}
