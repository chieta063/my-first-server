package main

import (
	"cmp"
	"database/sql"
	"log/slog"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"

	sloggin "github.com/samber/slog-gin"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	gin.SetMode(gin.ReleaseMode)

	// DB接続
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

	// DBの接続確認エンドポイント
	r.GET("/db-check", func(c *gin.Context) {
		err := db.Ping()
		if err != nil {
			c.JSON(500, gin.H{"error": "DB connection failed"})
			return
		}
		c.JSON(200, gin.H{"message": "DB connected!"})
	})

	port := cmp.Or(os.Getenv("PORT"), "8080")

	logger.Info("Server starting", slog.String("port", port))

	r.Run((":" + port))
}
