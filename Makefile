# Makefile
# Write needed script shortcuts for the project

# For local scripts that can vary (containers up / down, migrate up / down, psql seed)
# Local Makefile won't be pushed to remote repository,
# '-' char means not panic if file wasn't found
-include Makefile.local

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

