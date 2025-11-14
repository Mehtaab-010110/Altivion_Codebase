package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {

	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	// Get database credentials
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	sslmode := os.Getenv("DB_SSLMODE")

	// Display connection info (hide password)
	fmt.Printf("Host:     %s\n", host)
	fmt.Printf("Port:     %s\n", port)
	fmt.Printf("Database: %s\n", dbname)
	fmt.Printf("User:     %s\n", user)
	fmt.Printf("Password: %s (length: %d)\n\n", maskPassword(password), len(password))

	// Build connection string
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode,
	)

	// Connect to database
	fmt.Println("Connecting to database...")
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("❌ Failed to open connection:", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatal("❌ Failed to ping database:", err)
	}

	// Get PostgreSQL version
	var version string
	err = db.QueryRow("SELECT version()").Scan(&version)
	if err != nil {
		log.Println("Warning: Could not get version")
	} else {
		fmt.Println("📊 Database Info:")
		fmt.Printf("   Version: %s\n\n", version[:50]+"...")
	}

	// List tables
	fmt.Println("📋 Tables in database:")
	rows, err := db.Query(`
		SELECT table_name 
		FROM information_schema.tables 
		WHERE table_schema = 'public' 
		ORDER BY table_name
	`)
	if err != nil {
		log.Println("Warning: Could not list tables:", err)
	} else {
		defer rows.Close()
		tableCount := 0
		for rows.Next() {
			var tableName string
			if err := rows.Scan(&tableName); err != nil {
				continue
			}
			fmt.Printf("   - %s\n", tableName)
			tableCount++
		}
		if tableCount == 0 {
			fmt.Println("   (no tables found)")
		}
		fmt.Printf("\n   Total: %d tables\n", tableCount)
	}

	fmt.Println("\n🎉 Database test complete!")
	fmt.Println("\nNext steps:")
	fmt.Println("  1. Start Gateway:    go run cmd/gateway/main.go")
	fmt.Println("  2. Start Ingestion:  go run cmd/ingestion/main.go")
	fmt.Println("  3. Start API:        go run cmd/api/main.go")
}

// maskPassword hides most of the password for security
func maskPassword(password string) string {
	if len(password) == 0 {
		return "(empty)"
	}
	if len(password) <= 2 {
		return "**"
	}
	return password[:1] + "****" + password[len(password)-1:]
}
