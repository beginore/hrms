package postgres

import (
	"context"
	"database/sql"
	"log"
	"sync"
	"time"

	"hrms/internal/infrastructure/config"

	_ "github.com/jackc/pgx/v5/stdlib"
)

var (
	DB   *sql.DB
	once sync.Once
)

func InitDB(cfg *config.Config) {
	once.Do(func() {
		var err error

		DB, err = sql.Open("pgx", cfg.Database.DSN)
		if err != nil {
			log.Fatalf("Failed to open PostgreSQL connection: %v", err)
		}

		DB.SetMaxOpenConns(25)
		DB.SetMaxIdleConns(25)
		DB.SetConnMaxLifetime(5 * time.Minute)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err = DB.PingContext(ctx); err != nil {
			log.Fatalf("PostgreSQL ping failed: %v", err)
		}

		log.Println("PostgreSQL connection pool initialized successfully")
	})
}
