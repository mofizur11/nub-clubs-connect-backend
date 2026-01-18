package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/nub-clubs-connect/nub_admin_api/config"
	"github.com/nub-clubs-connect/nub_admin_api/database"
	"github.com/nub-clubs-connect/nub_admin_api/routes"
)

func main() {
	// Load configuration
	if err := config.LoadConfig(); err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database
	if err := database.Init(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	// Set Gin mode
	gin.SetMode(config.AppConfig.GinMode)

	// Create Gin router
	router := gin.Default()

	// Setup routes
	routes.SetupRoutes(router)

	// Start server
	port := config.AppConfig.Port
	fmt.Printf("🚀 Server starting on port %s\n", port)
	fmt.Println("📚 API Documentation: http://localhost:" + port + "/api/docs")

	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
