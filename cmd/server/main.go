package main

import (
	"context"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/umbranodens/kalabelajar/internal/config"
	"github.com/umbranodens/kalabelajar/internal/database"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	db, err := database.Connect(cfg.Database)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}

	if err := database.Ping(context.Background(), db); err != nil {
		log.Fatalf("database health check failed: %v", err)
	}
	if err := database.Migrate(db); err != nil {
		log.Fatalf("database migration failed: %v", err)
	}

	router := gin.Default()
	router.Static("/static", "./web/static")
	router.GET("/healthz", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	if err := router.Run(":" + cfg.App.Port); err != nil {
		log.Fatalf("start server: %v", err)
	}
}
