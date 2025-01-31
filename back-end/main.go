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

// Song represents a record in the database (Song представляет запись песни в базе данных)
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
    // Load environment variables (Загружаем переменные окружения)
    if err := godotenv.Load(); err != nil {
        log.Fatal("Error loading .env file (Ошибка загрузки .env файла)")
    }

    // Connect to the database (Подключаемся к базе данных)
    var err error
    db, err = sqlx.Connect("postgres", os.Getenv("DATABASE_URL"))
    if err != nil {
        log.Fatal("Failed to connect to database (Не удалось подключиться к базе данных):", err)
    }

    // Initialize Genius API client (Инициализируем клиент Genius API)
    apiToken := os.Getenv("GENIUS_API_TOKEN")
    if apiToken == "" {
        log.Fatal("GENIUS_API_TOKEN is not set in environment variables (GENIUS_API_TOKEN не установлен в переменных окружения)")
    }
    geniusClient = genius.NewClient(nil, apiToken)

    // Create Gin router (Создаем маршрутизатор Gin)
    r := gin.Default()

    // Define routes (Определяем маршруты)
    r.GET("/", rootHandler)                              // Root route (Корневой маршрут)
    r.GET("/songs", getSongs)                            // List of songs (Список песен)
    r.GET("/songs/:id", getSongDetails)                  // Song details by ID (Детали песни по ID)
    r.POST("/songs", addOrUpdateSong)                    // Add or update song (Добавление или обновление песни)
    r.DELETE("/songs/:id", deleteSong)                   // Delete song by ID (Удаление песни по ID)
    r.GET("/songs/:id/verses", getSongVerses)            // Get song verses (Получение куплетов песни)

    // Start the server (Запускаем сервер)
    r.Run(":8080")
}

// Handle requests to the root route (Обрабатывает запросы к корневому маршруту)
func rootHandler(c *gin.Context) {
    c.File("../front-end/index.html") // Send index.html file (Отправляем файл index.html)
}

// Get all songs from the database (Получает все песни из базы данных)
func getSongs(c *gin.Context) {
    var songs []Song
    err := db.Select(&songs, "SELECT id, group_name, song_name FROM songs")
    if err != nil {
        log.Printf("[ERROR] Failed to fetch songs (Не удалось получить песни): %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch songs (Не удалось получить песни)"})
        return
    }
    c.JSON(http.StatusOK, songs) // Return list of songs (Возвращаем список песен)
}

// Return full song details by ID (Возвращает полную информацию о песне по ID)
func getSongDetails(c *gin.Context) {
    id := c.Param("id")
    var song Song
    err := db.Get(&song, "SELECT * FROM songs WHERE id = $1", id)
    if err != nil {
        log.Printf("[ERROR] Failed to fetch song with ID %s (Не удалось найти песню с ID %s): %v", id, id, err)
        c.JSON(http.StatusNotFound, gin.H{"error": "Song not found (Песня не найдена)"})
        return
    }
    c.JSON(http.StatusOK, song) // Return song details (Возвращаем детали песни)
}

// Add new song or update existing one (Добавляет новую песню или обновляет существующую)
func addOrUpdateSong(c *gin.Context) {
    var song Song
    if err := c.BindJSON(&song); err != nil {
        log.Printf("[ERROR] Invalid input (Некорректный ввод): %v", err)
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input (Некорректный ввод)"})
        return
    }

    // Fetch additional song details from Genius API (Запрашиваем дополнительную информацию о песне из Genius API)
    geniusData, err := fetchSongDetails(song.Group, song.Song)
    if err != nil {
        log.Printf("[ERROR] Error fetching song details (Ошибка при получении данных о песне): %v", err)
        c.JSON(http.StatusNotFound, gin.H{"error": "No results found for the given song and group (Данные для данной песни и группы не найдены)"})
        return
    }

    // Insert or update song in the database (Вставляем или обновляем песню в базе данных)
    _, err = db.Exec(
        `INSERT INTO songs (group_name, song_name, release_date, text, link, album_cover_url)
       VALUES ($1, $2, $3, $4, $5, $6)
       ON CONFLICT (group_name, song_name) DO UPDATE SET
       release_date = $3, text = $4, link = $5, album_cover_url = $6`,
        song.Group, song.Song, geniusData.ReleaseDate, geniusData.Text, geniusData.Link, geniusData.AlbumCover,
    )
    if err != nil {
        log.Printf("[ERROR] Failed to save song (Не удалось сохранить песню): %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save song (Не удалось сохранить песню)"})
        return
    }
    c.Status(http.StatusOK) // Return success status (Возвращаем статус успеха)
}

// Delete song from the database by ID (Удаляет песню из базы данных по ID)
func deleteSong(c *gin.Context) {
    id := c.Param("id")
    _, err := db.Exec("DELETE FROM songs WHERE id = $1", id)
    if err != nil {
        log.Printf("[ERROR] Failed to delete song with ID %s (Не удалось удалить песню с ID %s): %v", id, id, err)
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete song (Не удалось удалить песню)"})
        return
    }
    c.Status(http.StatusOK) // Return success status (Возвращаем статус успеха)
}

// Return song verses with pagination (Возвращает куплеты песни с пагинацией)
func getSongVerses(c *gin.Context) {
    id := c.Param("id")
    limit := c.Query("limit")
    offset := c.Query("offset")

    var song Song
    err := db.Get(&song, "SELECT text FROM songs WHERE id = $1", id)
    if err != nil {
        log.Printf("[ERROR] Failed to fetch song with ID %s (Не удалось найти песню с ID %s): %v", id, id, err)
        c.JSON(http.StatusNotFound, gin.H{"error": "Song not found (Песня не найдена)"})
        return
    }

    verses := splitVerses(song.Text)
    start, end := calculatePagination(len(verses), limit, offset)
    paginatedVerses := verses[start:end]

    c.JSON(http.StatusOK, paginatedVerses) // Return paginated verses (Возвращаем пагинированные куплеты)
}

// Fetch additional song details using Genius API (Запрашивает дополнительную информацию о песне с использованием Genius API)
func fetchSongDetails(group, song string) (*Song, error) {
    query := fmt.Sprintf("%s %s", group, song)
    results, err := geniusClient.Search(query)
    if err != nil {
        return nil, fmt.Errorf("Error searching Genius API (Ошибка при поиске в Genius API): %v", err)
    }
    if len(results.Response.Hits) == 0 {
        return nil, fmt.Errorf("No results found for %s by %s (Результаты не найдены для %s от %s)", song, group, song, group)
    }

    result := results.Response.Hits[0].Result
    fullSong, err := geniusClient.GetSong(result.ID)
    if err != nil {
        return nil, fmt.Errorf("Error fetching full song details (Ошибка при получении полных данных о песне): %v", err)
    }

    // Extract release date from date components (Извлекаем дату релиза из компонентов даты)
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

// Split song text into verses (Разделяет текст песни на куплеты)
func splitVerses(text string) []string {
    return strings.Split(text, "\n\n")
}

// Calculate start and end indices for pagination (Вычисляет начальный и конечный индексы для пагинации)
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

// Parse pagination parameter into integer (Преобразует строковый параметр в целое число)
func parsePaginationParam(param string, defaultValue int) int {
    value := defaultValue
    if param != "" {
        fmt.Sscanf(param, "%d", &value)
    }
    return value
}