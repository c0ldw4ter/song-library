#!/bin/sh

# Whaiting for PostgreSQL conterines(Ждём, пока PostgreSQL станет доступен)
echo "Waiting for PostgreSQL to start..."
until psql "$DATABASE_URL" -c '\l' > /dev/null 2>&1; do
  echo "PostgreSQL is unavailable - sleeping"
  sleep 1
done
echo "PostgreSQL is ready!"

# Apply all migrations(Выполняем миграции)
echo "Applying database migrations..."
psql "$DATABASE_URL" -f /app/migrations/001_create_songs_table.sql || echo "Migrations skipped or already applied."

# Run app back-end(Запускаем приложение)
echo "Starting backend server..."
exec ./main