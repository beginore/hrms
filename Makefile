# Makefile
# Write needed script shortcuts for the project

# For local scripts that can vary (containers up / down, migrate up / down, psql seed)
# Local Makefile won't be pushed to remote repository,
# '-' char means not panic if file wasn't found
#-include Makefile.local
MIGRATIONS_DIR=internal/infrastructure/storage/postgres/migrations

DB_DSN=postgres://postgres:123@localhost:5432/hrms?sslmode=disable

LINTER_VERSION=1.64.5

migrate-up:
	@echo "Applying migrations..."
	@goose -dir $(MIGRATIONS_DIR) postgres "$(DB_DSN)" up

migrate-down:
	@echo "Rolling back last migration..."
	@goose -dir $(MIGRATIONS_DIR) postgres "$(DB_DSN)" down

migrate-status:
	@echo "Migration status:"
	@goose -dir $(MIGRATIONS_DIR) postgres "$(DB_DSN)" status

migrate-create:
	@echo "Creating new migration file: $(name)"
	@goose -dir $(MIGRATIONS_DIR) create $(name) sql

seed-fixtures:
	@echo "Seeding database..."
	@for f in ./test/fixtures/postgres/*.sql; do \
  		psql -f "$$f" "$(DB_DSN)"; \
	done

migrate-down-to:
	@if [ -z "$(version)" ]; then \
		echo "Error: version required"; \
		echo "Usage: make migrate-down-to version=3"; \
		exit 1; \
	fi
	@echo "⬇Migrating down to version $(version)..."
	@goose -dir $(MIGRATIONS_DIR) postgres "$(DB_DSN)" down-to $(version)

migrate-version:
	@echo "Current version:"
	@goose -dir $(MIGRATIONS_DIR) postgres "$(DB_DSN)" version

migrate-reset:
	@echo "Rolling back ALL migrations..."
	@goose -dir $(MIGRATIONS_DIR) postgres "$(DB_DSN)" down-to 0

lint:
	@echo 'run golangci lint'
	@go run github.com/golangci/golangci-lint/cmd/golangci-lint@v$(LINTER_VERSION) run --out-format=tab

.PHONY: hi

hi:
	@echo "Hello world"

# Sample of Makefile.local:


# .PHONY: up down start stop

# up:
# 	docker compose -f ./configs/local/docker-compose.yaml up --build

# down:
# 	docker compose -f ./configs/local/docker-compose.yaml down -v

# start:
# 	docker compose -f ./configs/local/docker-compose.yaml start

# stop:
# 	docker compose -f ./configs/local/docker-compose.yaml stop
