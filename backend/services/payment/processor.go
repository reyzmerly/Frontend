package payment

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/kwizera/university-invoices/shared/repositories"
)

func CompleteSuccessfulPayment(db *sql.DB, transactionID, actorType string) error {
	paymentRepo := repositories.NewPaymentRepository(db)
	invoiceRepo := repositories.NewInvoiceRepository(db)
	receiptRepo := repositories.NewReceiptRepository(db)
	auditRepo := repositories.NewAuditRepository(db)

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin successful payment transaction: %w", err)
	}
	defer tx.Rollback()

	payment, err := paymentRepo.GetByTransactionIDInTx(tx, transactionID)
	if err != nil {
		return fmt.Errorf("load payment for successful transaction: %w", err)
	}
	if payment["status"].(string) != "Pending" {
		return nil
	}

	now := time.Now()
	paymentID := payment["id"].(uint)
	invoiceID := payment["invoice_id"].(uint)
	amount := payment["amount"].(int64)
	reference := payment["mopay_reference"].(string)

	if err := paymentRepo.UpdateStatusInTx(tx, paymentID, "Paid", &now); err != nil {
		return fmt.Errorf("mark payment as paid: %w", err)
	}
	if err := invoiceRepo.UpdateBalanceInTx(tx, invoiceID, amount); err != nil {
		return fmt.Errorf("update invoice balance: %w", err)
	}

	invoice, err := invoiceRepo.GetByIDInTx(tx, invoiceID)
	if err != nil {
		return fmt.Errorf("load updated invoice: %w", err)
	}
	if _, err := receiptRepo.CreateInTx(tx, paymentID, invoiceID, amount, invoice["balance"].(int64)); err != nil {
		return fmt.Errorf("create payment receipt: %w", err)
	}

	if err := auditRepo.LogInTx(tx, actorType, nil, "Payment status updated", "payment", reference, "Pending", "Paid"); err != nil {
		return fmt.Errorf("log payment status update: %w", err)
	}
	if err := auditRepo.LogInTx(tx, "system", nil, "Receipt generated", "receipt", "", "", ""); err != nil {
		return fmt.Errorf("log receipt generation: %w", err)
	}
	if err := auditRepo.LogInTx(tx, "system", nil, "Invoice balance updated", "invoice", invoice["invoice_number"].(string), "", ""); err != nil {
		return fmt.Errorf("log invoice balance update: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit successful payment transaction: %w", err)
	}
	return nil
}
