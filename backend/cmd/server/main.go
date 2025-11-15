package main

import (
	"fmt"
	"log"
	"os"

	"talytics/internal/api"
	"talytics/internal/api/handlers"
	"talytics/internal/auth"
	"talytics/internal/database"
	"talytics/internal/services"
)

func main() {
	// Configuration
	port := getEnv("PORT", "8080")
	dbPath := getEnv("DB_PATH", "./data/talytics.db")
	claudeAPIKey := getEnv("CLAUDE_API_KEY", "")
	jwtSecret := getEnv("JWT_SECRET", "default-secret-change-in-production")

	// Initialize database
	log.Println("Initializing database...")
	if err := database.InitDB(dbPath); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.CloseDB()

	log.Println("Database initialized successfully")

	// Initialize JWT
	log.Println("Initializing JWT authentication...")
	auth.InitJWT(jwtSecret)
	log.Println("JWT authentication initialized")

	// Initialize services
	log.Println("Initializing services...")
	claudeService := services.NewClaudeService(claudeAPIKey)
	pdfService := services.NewPDFService()
	anomalyService := services.NewAnomalyService()

	// Check PDF tool availability
	if tool, err := pdfService.CheckPDFToolAvailability(); err != nil {
		log.Printf("WARNING: No PDF extraction tool available: %v", err)
		log.Printf("PDF upload feature will not work. Install pdftotext (poppler-utils) or pdfcpu.")
	} else {
		log.Printf("PDF extraction tool available: %s", tool)
	}

	// Initialize handlers with services
	handlers.InitServices(claudeService, pdfService)
	handlers.InitAnomalyService(anomalyService)
	log.Println("Anomaly detection service initialized")

	if claudeAPIKey == "" {
		log.Println("WARNING: CLAUDE_API_KEY not set. Claude-powered features will not work.")
	} else {
		log.Println("Claude API key configured")
	}

	// Setup router
	router := api.SetupRouter()

	// Start server
	addr := fmt.Sprintf(":%s", port)
	log.Printf("Starting TAlytics server on %s", addr)
	log.Printf("Health check: http://localhost%s/health", addr)

	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// getEnv retrieves environment variable with fallback default
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
