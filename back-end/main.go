package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

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
    r.GET("/", rootHandler)
    r.GET("/songs", getSongs)
    r.GET("/songs/:id", getSongDetails)
    r.POST("/songs", addOrUpdateSong)
    r.DELETE("/songs/:id", deleteSong)
    r.GET("/songs/:id/verses", getSongVerses)

    // Запускаем сервер
    r.Run(":8080")
}

// rootHandler обрабатывает запросы к корневому маршруту
func rootHandler(c *gin.Context) {
    c.File("../front-end/index.html")
}

// getSongs получает все песни из базы данных
func getSongs(c *gin.Context) {
    var songs []Song
    err := db.Select(&songs, "SELECT id, group_name, song_name FROM songs")
    if err != nil {
        log.Printf("[ERROR] Failed to fetch songs: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch songs"})
        return
    }
    c.JSON(http.StatusOK, songs)
}

// getSongDetails возвращает полную информацию о песне по ID
func getSongDetails(c *gin.Context) {
    id := c.Param("id")
    var song Song
    err := db.Get(&song, "SELECT * FROM songs WHERE id = $1", id)
    if err != nil {
        log.Printf("[ERROR] Failed to fetch song with ID %s: %v", id, err)
        c.JSON(http.StatusNotFound, gin.H{"error": "Song not found"})
        return
    }
    c.JSON(http.StatusOK, song)
}

// addOrUpdateSong добавляет новую песню или обновляет существующую
func addOrUpdateSong(c *gin.Context) {
    var song Song
    if err := c.BindJSON(&song); err != nil {
        log.Printf("[ERROR] Invalid input: %v", err)
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
        return
    }

    // Запрашиваем дополнительную информацию о песне из Genius API
    geniusData, err := fetchSongDetails(song.Group, song.Song)
    if err != nil {
        log.Printf("[ERROR] Error fetching song details: %v", err)
        c.JSON(http.StatusNotFound, gin.H{"error": "No results found for the given song and group"})
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
        log.Printf("[ERROR] Failed to save song: %v", err)
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
        log.Printf("[ERROR] Failed to delete song with ID %s: %v", id, err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete song"})
        return
    }
    c.Status(http.StatusOK)
}

// getSongVerses возвращает куплеты песни с пагинацией
func getSongVerses(c *gin.Context) {
    id := c.Param("id")
    limit := c.Query("limit")
    offset := c.Query("offset")

    var song Song
    err := db.Get(&song, "SELECT text FROM songs WHERE id = $1", id)
    if err != nil {
        log.Printf("[ERROR] Failed to fetch song with ID %s: %v", id, err)
        c.JSON(http.StatusNotFound, gin.H{"error": "Song not found"})
        return
    }

    verses := splitVerses(song.Text)
    start, end := calculatePagination(len(verses), limit, offset)
    paginatedVerses := verses[start:end]

    c.JSON(http.StatusOK, paginatedVerses)
}

// fetchSongDetails запрашивает дополнительную информацию о песне с использованием Genius API
func fetchSongDetails(group, song string) (*Song, error) {
    query := fmt.Sprintf("%s %s", group, song)
    results, err := geniusClient.Search(query)
    if err != nil {
        return nil, fmt.Errorf("error searching Genius API: %v", err)
    }
    if len(results.Response.Hits) == 0 {
        return nil, fmt.Errorf("no results found for %s by %s", song, group)
    }

    result := results.Response.Hits[0].Result
    fullSong, err := geniusClient.GetSong(result.ID)
    if err != nil {
        return nil, fmt.Errorf("error fetching full song details: %v", err)
    }

    // Извлекаем дату релиза из компонентов даты
    releaseDate := ""
    if result.ReleaseDateComponents != nil {
        releaseDate = fmt.Sprintf(
            "%d-%02d-%02d",
            result.ReleaseDateComponents.Year,
            result.ReleaseDateComponents.Month,
            result.ReleaseDateComponents.Day,
        )
    }

    return &Song{
        ReleaseDate: releaseDate,
        Text:        fullSong.Lyrics,
        Link:        result.URL,
        AlbumCover:  result.SongArtImageURL,
    }, nil
}

// splitVerses разделяет текст песни на куплеты
func splitVerses(text string) []string {
    return strings.Split(text, "\n\n")
}

// calculatePagination вычисляет начальный и конечный индексы для пагинации
func calculatePagination(total int, limitStr, offsetStr string) (int, int) {
    limit := parsePaginationParam(limitStr, 10)
    offset := parsePaginationParam(offsetStr, 0)
    start := offset
    end := offset + limit
    if end > total {
        end = total
    }
    return start, end
}

// parsePaginationParam преобразует строковый параметр в целое число
func parsePaginationParam(param string, defaultValue int) int {
    value := defaultValue
    if param != "" {
        fmt.Sscanf(param, "%d", &value)
    }
    return value
}