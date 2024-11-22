# song-library

Для запуска

Перейти в директорию back-end
~psql -U postgres -d songsdb -f migrations/001_create_songs_table.sql
~go mod tidy
~go run main.go
