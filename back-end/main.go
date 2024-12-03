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

// Song represents a song record in the database
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
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	// Connect to the database
	var err error
	db, err = sqlx.Connect("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Initialize Genius API client
	apiToken := os.Getenv("GENIUS_API_TOKEN")
	if apiToken == "" {
		log.Fatal("GENIUS_API_TOKEN is not set in environment variables")
	}
	geniusClient = genius.NewClient(nil, apiToken)

	// Create Gin router
	r := gin.Default()
	r.Static("/static", "../front-end")

	// Routes
	r.GET("/", rootHandler)
	r.GET("/songs", getSongs)
	r.POST("/songs", addOrUpdateSong)
	r.DELETE("/songs/:id", deleteSong)

	// Run server
	r.Run(":8080")
}

// rootHandler serves the index.html file
func rootHandler(c *gin.Context) {
	c.File("../front-end/index.html")
}

// getSongs fetches all songs from the database
func getSongs(c *gin.Context) {
	var songs []Song
	err := db.Select(&songs, "SELECT * FROM songs")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch songs"})
		return
	}
	c.JSON(http.StatusOK, songs)
}

// addOrUpdateSong adds a new song or updates an existing one
func addOrUpdateSong(c *gin.Context) {
	var song Song
	if err := c.BindJSON(&song); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// Call Genius API for additional information
	geniusData, err := fetchSongDetails(song.Group, song.Song)
	if err != nil {
		log.Printf("Error fetching song details: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Could not fetch song details"})
		return
	}

	// Insert or update song in the database
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

// deleteSong removes a song from the database by ID
func deleteSong(c *gin.Context) {
	id := c.Param("id")
	_, err := db.Exec("DELETE FROM songs WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete song"})
		return
	}
	c.Status(http.StatusOK)
}

// fetchSongDetails fetches additional information about a song using Genius API
func fetchSongDetails(group, song string) (*Song, error) {
	query := fmt.Sprintf("%s %s", group, song)

	// Search using Genius API client
	results, err := geniusClient.Search(query)
	if err != nil {
		return nil, fmt.Errorf("error searching Genius API: %v", err)
	}

	// Check if results are available
	if len(results.Response.Hits) == 0 {
		return nil, fmt.Errorf("no results found for %s by %s", song, group)
	}

	// Find the best match (relaxed comparison)
	for _, hit := range results.Response.Hits {
		result := hit.Result
		if strings.Contains(result.FullTitle, song) && strings.Contains(result.PrimaryArtist.Name, group) {
			return &Song{
				ReleaseDate: result.ReleaseDate,
				Text:        "Lyrics unavailable in this implementation",
				Link:        result.URL,
				AlbumCover:  result.HeaderImageURL,
			}, nil
		}
	}
	return nil, fmt.Errorf("no suitable match found for %s by %s", song, group)
}


