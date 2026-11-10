package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/kwizera/university-invoices/internal/config"
	"github.com/kwizera/university-invoices/internal/jobs"
	"github.com/kwizera/university-invoices/internal/routes"
)

func main() {
	log.Println("Starting University of Kigali Invoices API Server...")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log.Printf("Configuration loaded successfully")
	log.Printf("Database: %s@%s:%s/%s", cfg.DBUser, cfg.DBHost, cfg.DBPort, cfg.DBName)

	// Initialize database connection
	db, err := config.InitDB(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Setup routes
	router := routes.SetupRouter(db, cfg)

	// Start background jobs
	jobScheduler := jobs.NewJobScheduler(db, cfg)
	jobScheduler.Start()
	defer jobScheduler.Stop()

	// Start server
	addr := ":" + cfg.ServerPort
	log.Printf("Server starting on %s", addr)
	log.Println("Available endpoints:")
	log.Println("  GET    /health")
	log.Println("  POST   /api/v1/auth/login")
	log.Println("  POST   /api/v1/auth/logout")
	log.Println("  POST   /api/v1/auth/forgot-password")
	log.Println("  GET    /api/v1/auth/me")
	log.Println("  GET    /api/v1/student/dashboard")
	log.Println("  GET    /api/v1/invoices")
	log.Println("  POST   /api/v1/invoices")
	log.Println("  GET    /api/v1/payments")
	log.Println("  POST   /api/v1/invoices/{invoiceId}/payments")
	log.Println("  GET    /api/v1/receipts/{receiptId}")
	log.Println("  GET    /api/v1/finance/dashboard")
	log.Println("  GET    /api/v1/fee-types")
	log.Println("  GET    /api/v1/admin/users")

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		log.Println("Shutting down server...")
		jobScheduler.Stop()
		db.Close()
		os.Exit(0)
	}()

	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
