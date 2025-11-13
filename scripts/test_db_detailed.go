package main

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	godotenv.Load()

	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	fmt.Println("🔍 Testing database connection with details:\n")
	fmt.Printf("Host: %s\n", host)
	fmt.Printf("Port: %s\n", port)
	fmt.Printf("User: %s\n", user)
	fmt.Printf("Password length: %d characters\n", len(password))
	fmt.Printf("Password starts with: %s...\n", password[:min(3, len(password))])
	fmt.Printf("Database: %s\n\n", dbname)

	// Try connection 1: With dbname
	fmt.Println("Attempt 1: Connecting to specified database...")
	connStr1 := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db1, err := sql.Open("postgres", connStr1)
	if err == nil {
		err = db1.Ping()
		if err == nil {
			fmt.Println("✅ SUCCESS! Connected to", dbname)
			db1.Close()
			return
		}
	}
	fmt.Printf("❌ Failed: %v\n\n", err)

	// Try connection 2: Default postgres database
	fmt.Println("Attempt 2: Connecting to default 'postgres' database...")
	connStr2 := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=postgres sslmode=disable",
		host, port, user, password)

	db2, err := sql.Open("postgres", connStr2)
	if err == nil {
		err = db2.Ping()
		if err == nil {
			fmt.Println("✅ SUCCESS! Connected to postgres database")

			// List available databases
			fmt.Println("\n📊 Available databases:")
			rows, _ := db2.Query("SELECT datname FROM pg_database WHERE datistemplate = false")
			for rows.Next() {
				var name string
				rows.Scan(&name)
				fmt.Printf("  - %s\n", name)
			}
			db2.Close()
			return
		}
	}
	fmt.Printf("❌ Failed: %v\n\n", err)

	// Try connection 3: With SSL required
	fmt.Println("Attempt 3: Connecting with SSL required...")
	connStr3 := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=require",
		host, port, user, password, dbname)

	db3, err := sql.Open("postgres", connStr3)
	if err == nil {
		err = db3.Ping()
		if err == nil {
			fmt.Println("✅ SUCCESS! Connected with SSL")
			db3.Close()
			return
		}
	}
	fmt.Printf("❌ Failed: %v\n\n", err)

	fmt.Println("\n❌ All connection attempts failed.")
	fmt.Println("\n💡 Next steps:")
	fmt.Println("1. Verify the password is correct (no extra spaces)")
	fmt.Println("2. Ask database engineer to confirm:")
	fmt.Println("   - Exact username (case-sensitive)")
	fmt.Println("   - Exact password")
	fmt.Println("   - Database name")
	fmt.Println("   - If your IP needs to be whitelisted")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
