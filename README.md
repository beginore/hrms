# Local setup

1. Install the migration tool:

   `go install github.com/pressly/goose/v3/cmd/goose@latest`
2. Apply migration:

   `goose -dir internal/infrastructure/storage/postgres/migrations postgres "postgres://postgres:postgres@localhost:5432/iam?sslmode=disable" up` or `make migrate-up`
3. Run the application:
4.
`go run .\cmd\main.go ".\configs\local\config.toml"`
