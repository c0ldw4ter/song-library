# Переменные (vars)
DB_URL ?= $(shell grep '^DATABASE_URL=' back-end/.env | sed 's/^DATABASE_URL=//')
GO_RUN = go run $(shell pwd)/back-end/main.go
MIGRATION_DIR = $(shell pwd)/back-end/migrations

# Цели (targets)
.PHONY: run migrate

# Запуск проекта (run app)
run: migrate
	@echo "Starting the backend server..."
	cd $(shell pwd)/back-end && $(GO_RUN)

# Применение миграций (apply migrations)
migrate:
	@echo "Checking if migrations are needed..."
	psql "$(DB_URL)" -f $(MIGRATION_DIR)/001_create_songs_table.sql || echo "Migrations skipped or already applied."
	@echo "Migrations checked successfully."

