package postgres

import (
	"context"
	"database/sql"
)

//go:generate mockgen -source=interface.go -package=mock -destination=mock/interface_mock.go -mock_names=Database=MockDatabase
type Database interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	SelectContext(ctx context.Context, dest any, query string, args ...any) error
	GetContext(ctx context.Context, dest any, query string, args ...any) error
}
