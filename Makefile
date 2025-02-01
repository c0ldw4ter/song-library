# Vars(Переменные)
DB_URL ?= $(shell grep '^DATABASE_URL=' back-end/.env | sed 's/^DATABASE_URL=//')
GO_RUN = go run $(shell pwd)/back-end/main.go
GO_TIDY = go mod tidy
MIGRATION_DIR = $(shell pwd)/back-end/migrations

# Targets(Цели)
.PHONY: run migrate

# Run app and all dependencies(Запуск проекта и всех зависимостей)
run: migrate
	@echo "Starting the backend server..."
	cd $(shell pwd)/back-end && $(GO_TIDY) && $(GO_RUN)

# Apply migrations(Применение миграций)
migrate:
	@echo "Checking if migrations are needed..."
	psql "$(DB_URL)" -f $(MIGRATION_DIR)/001_create_songs_table.sql || echo "Migrations skipped or already applied."
	@echo "Migrations checked successfully."

