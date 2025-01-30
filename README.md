## song-library  using Golang + DB(PostgreSQL) + API Genius to fetch details for each song 

## To run

- Go to the back-end directory
- Create a PostgreSQL database for the postgres caller named songsdb
- run migrations `psql -U postgres -d songsdb -f migrations/001_create_songs_table.sql`
- dowload all depandancies `go mod tidy`
- to run `go run main.go`
