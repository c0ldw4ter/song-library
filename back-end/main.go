package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/broxgit/genius"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

// Song представляет запись песни в базе данных
type Song struct {
	ID          int    `json:"id" db:"id"`
	Group       string `json:"group" db:"group_name"`
	Song        string `json:"song" db:"song_name"`
	ReleaseDate string `json:"release_date" db:"release_date"`
	Text        string `json:"text" db:"text"`
	Link        string `json:"link" db:"link"`
	AlbumCover  string `json:"album_cover_url" db:"album_cover_url"`
}

var db *sqlx.DB
var geniusClient *genius.Client

func main() {
	// Загружаем переменные окружения
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	// Подключаемся к базе данных
	var err error
	db, err = sqlx.Connect("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Инициализируем клиент Genius API
	apiToken := os.Getenv("GENIUS_API_TOKEN")
	if apiToken == "" {
		log.Fatal("GENIUS_API_TOKEN is not set in environment variables")
	}
	geniusClient = genius.NewClient(nil, apiToken)

	// Создаем маршрутизатор Gin
	r := gin.Default()

	// Определяем маршруты
	r.GET("/songs", getSongs)
	r.POST("/songs", addOrUpdateSong)
	r.DELETE("/songs/:id", deleteSong)

	// Запускаем сервер
	r.Run(":8080")
}

// getSongs получает все песни из базы данных с пагинацией и фильтрацией
func getSongs(c *gin.Context) {
	var songs []Song
	group := c.Query("group")
	song := c.Query("song")
	limit := c.Query("limit")
	offset := c.Query("offset")

	query := "SELECT * FROM songs WHERE TRUE"
	args := []interface{}{}
	if group != "" {
		query += " AND group_name ILIKE $1"
		args = append(args, "%"+group+"%")
	}
	if song != "" {
		query += " AND song_name ILIKE $2"
		args = append(args, "%"+song+"%")
	}
	if limit != "" {
		query += " LIMIT $3"
		args = append(args, limit)
	}
	if offset != "" {
		query += " OFFSET $4"
		args = append(args, offset)
	}

	err := db.Select(&songs, query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch songs"})
		return
	}
	c.JSON(http.StatusOK, songs)
}

// addOrUpdateSong добавляет новую песню или обновляет существующую
func addOrUpdateSong(c *gin.Context) {
	var song Song
	if err := c.BindJSON(&song); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// Запрашиваем дополнительную информацию о песне из Genius API
	geniusData, err := fetchSongDetails(song.Group, song.Song)
	if err != nil {
		log.Printf("Error fetching song details: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Could not fetch song details"})
		return
	}

	// Вставляем или обновляем песню в базе данных
	_, err = db.Exec(
		`INSERT INTO songs (group_name, song_name, release_date, text, link, album_cover_url)
       VALUES ($1, $2, $3, $4, $5, $6)
       ON CONFLICT (group_name, song_name) DO UPDATE SET
       release_date = $3, text = $4, link = $5, album_cover_url = $6`,
		song.Group, song.Song, geniusData.ReleaseDate, geniusData.Text, geniusData.Link, geniusData.AlbumCover,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save song"})
		return
	}
	c.Status(http.StatusOK)
}

// deleteSong удаляет песню из базы данных по ID
func deleteSong(c *gin.Context) {
	id := c.Param("id")
	_, err := db.Exec("DELETE FROM songs WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete song"})
		return
	}
	c.Status(http.StatusOK)
}

// fetchSongDetails запрашивает дополнительную информацию о песне с использованием Genius API
func fetchSongDetails(group, song string) (*Song, error) {
	query := fmt.Sprintf("%s %s", group, song)

	// Используем клиент Genius для поиска
	results, err := geniusClient.Search(query)
	if err != nil {
		return nil, fmt.Errorf("error searching Genius API: %v", err)
	}

	// Проверяем, есть ли результаты
	if len(results.Response.Hits) == 0 {
		return nil, fmt.Errorf("no results found for %s by %s", song, group)
	}

	// Берем первый результат
	result := results.Response.Hits[0].Result
	return &Song{
		ReleaseDate: result.ReleaseDate,
		Text:        "Lyrics unavailable in this implementation", // Здесь можно реализовать получение текста песни
		Link:        result.URL,
		AlbumCover:  result.HeaderImageURL,
	}, nil
}










