package main

import (
	"fmt"
	"golang-app/internal/config"
	"log"
	"golang-app/pkg/db"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
// 2. Database Connection (Sử dụng module pkg/db vừa tách)
	database, err := db.NewPostgresDB(cfg)
	if err != nil {
		log.Fatalf("Failed to init database: %v", err)
	}
	defer database.Close()
	// Set Gin mode
	server := config.NewServer(database)

    // Setup Router
    r := config.NewRouter(server)

	// Server runs on configured port
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("Server is running at %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}