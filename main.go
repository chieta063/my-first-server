package main

import (
	"cmp"
	"database/sql"
	"log/slog"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	sloggin "github.com/samber/slog-gin"
)

type Message struct {
	ID        int       `json:"id"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	gin.SetMode(gin.ReleaseMode)

	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		logger.Error("Failed to connect to database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer db.Close()

	r := gin.New()
	r.Use(sloggin.New(logger))
	r.Use(gin.Recovery())

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Hello world!",
		})
	})

	r.GET("/hello/:name", func(c *gin.Context) {
		name := c.Param("name")
		c.JSON(200, gin.H{
			"message": "Hello, " + name + "!",
		})
	})

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	r.GET("/db-check", func(c *gin.Context) {
		err := db.Ping()
		if err != nil {
			c.JSON(500, gin.H{"error": "DB connection failed"})
			return
		}
		c.JSON(200, gin.H{"message": "DB connected!"})
	})

	r.POST("/messages", func(c *gin.Context) {
		var input struct {
			Body string `json:"body"`
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(400, gin.H{"error": "Invalid request"})
			return
		}
		var message Message
		err := db.QueryRow(
			"INSERT INTO messages (body) VALUES ($1) RETURNING id, body, created_at",
			input.Body,
		).Scan(&message.ID, &message.Body, &message.CreatedAt)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to save message"})
			return
		}
		c.JSON(201, message)
	})

	r.GET("/messages", func(c *gin.Context) {
		rows, err := db.Query("SELECT id, body, created_at FROM messages ORDER BY created_at DESC")
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to get messages"})
			return
		}
		defer rows.Close()

		messages := []Message{}
		for rows.Next() {
			var message Message
			if err := rows.Scan(&message.ID, &message.Body, &message.CreatedAt); err != nil {
				c.JSON(500, gin.H{"error": "Failed to scan message"})
				return
			}
			messages = append(messages, message)
		}
		c.JSON(200, messages)
	})

	port := cmp.Or(os.Getenv("PORT"), "8080")
	logger.Info("Server starting", slog.String("port", port))
	r.Run((":" + port))
}

