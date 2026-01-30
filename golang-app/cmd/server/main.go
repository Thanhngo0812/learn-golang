package main

import (
	"fmt"
	"golang-app/internal/config"
	"log"
	"net/http"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Set Gin mode
	gin.SetMode(cfg.Server.Mode)

	r := gin.Default()

	r.GET("/hello", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "success",
			"message": fmt.Sprintf("Hello from %s %s!", cfg.App.Name, cfg.App.Version),
		})
	})

	// Server runs on configured port
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("Server is running at %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}