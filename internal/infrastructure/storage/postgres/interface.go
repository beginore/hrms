package postgres

import (
	"context"
	"database/sql"
)

// This file with interface is needed to import it in all repositories that work with Postgres.
// Mocks for the work with Postgres database are also created here and imported in tests of repositories.

//go:generate mockgen -source=interface.go -package=mock -destination=mock/interface_mock.go -mock_names=Database=MockDatabase
type Database interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}
