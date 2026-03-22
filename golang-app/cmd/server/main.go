package main

import (
	"fmt"
	"golang-app/internal/config"
	"golang-app/internal/api"
	"log"
	"golang-app/pkg/db"
)

// @title           AuctionHub API
// @version         1.0
// @description     This is a sample auction server.
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer <your_token>" to authenticate.
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
	sqlDB, err := database.DB()
    if err != nil {
        log.Fatalf("Failed to get sql.DB from gorm: %v", err)
    }
    defer sqlDB.Close()

	
	// Set Gin mode
	server := api.NewServer(database, cfg)
    
	// Setup Router
    r := api.NewRouter(server)

	// Server runs on configured port
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("Server is running at %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}