package jobs

import (
	"database/sql"
	"log"
	"strings"
	"time"

	"github.com/kwizera/university-invoices/internal/config"
	"github.com/kwizera/university-invoices/services/hdev"
	paymentservice "github.com/kwizera/university-invoices/services/payment"
)

type JobScheduler struct {
	db          *sql.DB
	hdevService *hdev.Service
	stopChan    chan struct{}
}

func NewJobScheduler(db *sql.DB, cfg *config.Config) *JobScheduler {
	return &JobScheduler{
		db:          db,
		hdevService: hdev.NewService(cfg.HDevAPIID, cfg.HDevAPIKey, cfg.PaymentCallbackURL, cfg.PaymentTimeoutMS),
		stopChan:    make(chan struct{}),
	}
}

func (js *JobScheduler) Start() {
	log.Println("Starting background job scheduler...")

	go js.runPaymentStatusCheck()
	go js.runOverdueInvoiceCheck()
}

func (js *JobScheduler) Stop() {
	close(js.stopChan)
	log.Println("Background job scheduler stopped")
}

func (js *JobScheduler) runPaymentStatusCheck() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			js.checkPendingPayments()
		case <-js.stopChan:
			return
		}
	}
}

func (js *JobScheduler) checkPendingPayments() {
	log.Println("Checking pending payments...")

	query := `
		SELECT id, mopay_transaction_id
		FROM payments
		WHERE status = 'Pending'
	`

	rows, err := js.db.Query(query)
	if err != nil {
		log.Printf("Error checking pending payments: %v", err)
		return
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var id uint
		var hdevTransactionID string
		if err := rows.Scan(&id, &hdevTransactionID); err != nil {
			continue
		}

		// Check status with HDEV API
		statusResp, err := js.hdevService.CheckTransactionStatus(hdevTransactionID)
		if err != nil {
			log.Printf("Error checking status for payment %d: %v", id, err)
			continue
		}

		// Update payment status based on HDEV response
		var newStatus string
		if strings.EqualFold(strings.TrimSpace(statusResp.Status), "success") || strings.EqualFold(strings.TrimSpace(statusResp.Status), "paid") {
			newStatus = "Paid"
		} else if strings.EqualFold(strings.TrimSpace(statusResp.Status), "failed") || strings.EqualFold(strings.TrimSpace(statusResp.Status), "cancelled") {
			newStatus = "Failed"
		} else {
			// Still pending
			continue
		}

		if newStatus == "Paid" {
			if err := paymentservice.CompleteSuccessfulPayment(js.db, hdevTransactionID, "system"); err != nil {
				log.Printf("Error completing successful payment %d: %v", id, err)
				continue
			}
		} else {
			updateQuery := `UPDATE payments SET status = ?, status_updated_at = NOW() WHERE id = ? AND status = 'Pending'`
			result, err := js.db.Exec(updateQuery, newStatus, id)
			if err != nil {
				log.Printf("Error updating payment %d status: %v", id, err)
				continue
			}
			updated, err := result.RowsAffected()
			if err != nil {
				log.Printf("Error checking payment %d status update: %v", id, err)
				continue
			}
			if updated == 0 {
				continue
			}
		}

		count++
		log.Printf("Updated payment %d to %s", id, newStatus)
	}

	if count > 0 {
		log.Printf("Updated %d payment statuses", count)
	}
}

func (js *JobScheduler) runOverdueInvoiceCheck() {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			js.checkOverdueInvoices()
		case <-js.stopChan:
			return
		}
	}
}

func (js *JobScheduler) checkOverdueInvoices() {
	log.Println("Checking overdue invoices...")

	query := `
		UPDATE invoices
		SET status = 'Overdue'
		WHERE status IN ('Issued', 'Partially Paid')
		  AND due_date < CURDATE()
		  AND total_amount > paid_amount
	`

	result, err := js.db.Exec(query)
	if err != nil {
		log.Printf("Error checking overdue invoices: %v", err)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		log.Printf("Marked %d invoices as overdue", rowsAffected)
	}
}
