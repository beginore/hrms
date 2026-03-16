# Local setup

1. Install the migration tool:
   `go install github.com/pressly/goose/v3/cmd/goose@latest`
2. Apply migration:
   `make migrate-up`
   `make seed-fixtures`
3. Run the application:
`go run .\cmd\main.go ".\configs\local\config.toml"`
