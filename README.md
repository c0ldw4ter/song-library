# song-library сдалеанный при помощи Golang + БД(PostgreSQL

branch feature

## Для запуска

- Перейти в директорию back-end
- Создаём базу данных на PostgreSQL для позователя postgres с название songsdb
- psql -U postgres -d songsdb -f migrations/001_create_songs_table.sql
- go mod tidy
- go run main.go
