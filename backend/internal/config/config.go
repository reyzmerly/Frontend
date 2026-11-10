package config

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

type Config struct {
	DBHost             string
	DBPort             string
	DBUser             string
	DBPassword         string
	DBName             string
	ServerPort         string
	JWTSecret          string
	HDevAPIID          string
	HDevAPIKey         string
	PaymentCallbackURL string
	PaymentTimeoutMS   int
}

func Load() (*Config, error) {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found or error loading, using defaults")
	}

	cfg := &Config{
		DBHost:             getEnv("DB_HOST", "localhost"),
		DBPort:             getEnv("DB_PORT", "3306"),
		DBUser:             getEnv("DB_USER", "root"),
		DBPassword:         getEnv("DB_PASSWORD", ""),
		DBName:             getEnv("DB_NAME", "university_invoices"),
		ServerPort:         getEnv("SERVER_PORT", "9107"),
		JWTSecret:          getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
		HDevAPIID:          getEnv("HDEV_API_ID", ""),
		HDevAPIKey:         getEnv("HDEV_API_KEY", ""),
		PaymentCallbackURL: getEnv("PAYMENT_CALLBACK_URL", ""),
		PaymentTimeoutMS:   getEnvInt("PAYMENT_TIMEOUT_MS", 25000),
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		var intValue int
		_, err := fmt.Sscanf(value, "%d", &intValue)
		if err == nil {
			return intValue
		}
	}
	return defaultValue
}

func InitDB(cfg *Config) (*sql.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBName,
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	log.Println("Database connection established successfully")
	return db, nil
}
