package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	// Get database connection from environment
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "3306")
	dbUser := getEnv("DB_USER", "root")
	dbPassword := getEnv("DB_PASSWORD", "")
	dbName := getEnv("DB_NAME", "university_invoices")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		dbUser, dbPassword, dbHost, dbPort, dbName)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	log.Println("Connected to database successfully")

	// Create migrations table if it doesn't exist
	if err := createMigrationsTable(db); err != nil {
		log.Fatalf("Failed to create migrations table: %v", err)
	}

	// Get migration files
	migrationFiles, err := getMigrationFiles()
	if err != nil {
		log.Fatalf("Failed to get migration files: %v", err)
	}

	// Run migrations
	for _, file := range migrationFiles {
		if err := runMigration(db, file); err != nil {
			log.Fatalf("Failed to run migration %s: %v", file, err)
		}
	}

	log.Println("All migrations completed successfully")
}

func createMigrationsTable(db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			id INT AUTO_INCREMENT PRIMARY KEY,
			version VARCHAR(255) NOT NULL UNIQUE,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
	`
	_, err := db.Exec(query)
	return err
}

func getMigrationFiles() ([]string, error) {
	files, err := ioutil.ReadDir(".")
	if err != nil {
		return nil, err
	}

	var migrationFiles []string
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".sql") && file.Name() != "migrate.go" {
			migrationFiles = append(migrationFiles, file.Name())
		}
	}

	sort.Strings(migrationFiles)
	return migrationFiles, nil
}

func runMigration(db *sql.DB, filename string) error {
	// Check if migration has already been applied
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE version = ?", filename).Scan(&count)
	if err != nil {
		return err
	}

	if count > 0 {
		log.Printf("Migration %s already applied, skipping", filename)
		return nil
	}

	// Read migration file
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	// Execute migration
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	if _, err := tx.Exec(string(content)); err != nil {
		tx.Rollback()
		return err
	}

	// Record migration
	if _, err := tx.Exec("INSERT INTO schema_migrations (version) VALUES (?)", filename); err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	log.Printf("Applied migration: %s", filename)
	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
