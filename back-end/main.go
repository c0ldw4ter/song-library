package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

// Song represents a song record in the database
type Song struct {
	ID    int    `json:"id" db:"id"`
	Group string `json:"group" db:"group_name"`
	Song  string `json:"song" db:"song_name"`
}

var db *sqlx.DB

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

//func for Root rout
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

	_, err := db.Exec("INSERT INTO songs (group_name, song_name) VALUES ($1, $2) ON CONFLICT (id) DO UPDATE SET group_name = $1, song_name = $2",
		song.Group, song.Song)
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
