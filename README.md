# Local setup

1. Install the migration tool:

   `go install github.com/pressly/goose/v3/cmd/goose@latest`
2. Apply migration:

   `goose -dir internal/infrastructure/storage/postgres/migrations postgres "postgres://postgres:postgres@localhost:5432/iam?sslmode=disable" up` or `make migrate-up`
3. Run the application:
4.
`go run .\cmd\main.go ".\configs\local\config.toml"`

## Invite SMTP

Invites use SMTP configuration from the `[smtp]` section.

For Gmail SMTP use:

`host = "smtp.gmail.com"`

`port = 587`

`username = "your-email@gmail.com"`

`password = "your-gmail-app-password"`

`sender_email = "your-email@gmail.com"`

Gmail requires an App Password when 2-Step Verification is enabled.
