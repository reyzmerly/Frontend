package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	// Database connection
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "3306")
	dbUser := getEnv("DB_USER", "root")
	dbPassword := getEnv("DB_PASSWORD", "")
	dbName := getEnv("DB_NAME", "university_invoices")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		dbUser, dbPassword, dbHost, dbPort, dbName)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	log.Println("Connected to database successfully")

	// Run seed files in order
	seedFiles := []string{
		"seed_users.sql",
		"seed_programmes.sql",
		"seed_semesters.sql",
		"seed_students.sql",
	}

	for _, file := range seedFiles {
		content, err := os.ReadFile(file)
		if err != nil {
			log.Printf("Warning: Could not read %s: %v", file, err)
			continue
		}

		_, err = db.Exec(string(content))
		if err != nil {
			log.Printf("Skipped %s (already exists): %v", file, err)
			continue
		}

		log.Printf("Seeded: %s", file)
	}

	log.Println("Seeding completed")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
