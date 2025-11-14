package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func main() {
	fmt.Println("🔍 Checking available databases...\n")

	// Connect to default postgres database
	connStr := "host=192.168.1.232 port=5432 user=postgres password=postgres dbname=postgres sslmode=disable"

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("❌ Failed to connect:", err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatal("❌ Cannot reach database server:", err)
	}

	fmt.Println("✅ Connected to database server!")

	// List all databases
	rows, err := db.Query("SELECT datname FROM pg_database WHERE datistemplate = false ORDER BY datname")
	if err != nil {
		log.Fatal("❌ Failed to list databases:", err)
	}
	defer rows.Close()

	fmt.Println("\n📊 Available databases:")
	for rows.Next() {
		var dbName string
		if err := rows.Scan(&dbName); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("  - %s\n", dbName)
	}

	fmt.Println("\n💡 Ask your database engineer which database name to use from the list above.")
}
