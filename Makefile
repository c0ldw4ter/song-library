# Vars(Переменные)
COMPOSE = docker-compose
DB_URL ?= $(shell grep '^DATABASE_URL=' back-end/.env | sed 's/^DATABASE_URL=//')
MIGRATION_DIR = back-end/migrations

# Targets(Цели)
.PHONY: run migrate down

# Run all docker-contaners(Запуск всех контейнеров)
run:
	@echo "Starting all containers..."
	$(COMPOSE) up --build

# Manual apply migrations(Применение миграций (вручную))
migrate:
	@echo "Applying database migrations..."
	docker exec -it backend_app psql "$(DB_URL)" -f $(MIGRATION_DIR)/001_create_songs_table.sql
	@echo "Migrations applied successfully."

# Stop all docker-containers(Остановка всех контейнеров)
down:
	@echo "Stopping all containers..."
	$(COMPOSE) down

