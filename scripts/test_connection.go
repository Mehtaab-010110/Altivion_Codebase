package main

import (
	"fmt"
	"log"

	"silentraven/internal/database"
	"silentraven/pkg/config"
)

func main() {
	fmt.Println("🧪 Testing Database Connection\n")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("❌ Failed to load config:", err)
	}

	fmt.Printf("Connecting to: %s:%s/%s\n", cfg.DBHost, cfg.DBPort, cfg.DBName)
	fmt.Printf("User: %s\n\n", cfg.DBUser)

	// Connect to database
	db, err := database.New(cfg)
	if err != nil {
		log.Fatal("❌ Database connection failed:", err)
	}
	defer db.Close()

	fmt.Println("✅ Database connection successful!")
	fmt.Println("\n🎉 Ready to build services!")
}
