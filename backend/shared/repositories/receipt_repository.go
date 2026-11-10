package repositories

import (
	"database/sql"
	"errors"
	"time"

	"github.com/kwizera/university-invoices/shared/utils"
)

type ReceiptRepository struct {
	db *sql.DB
}

func NewReceiptRepository(db *sql.DB) *ReceiptRepository {
	return &ReceiptRepository{db: db}
}

func (r *ReceiptRepository) GetByID(id uint) (map[string]interface{}, error) {
	query := `
		SELECT r.id, r.receipt_number, r.payment_id, r.invoice_id, r.amount_paid, r.new_balance, r.issued_at,
		       i.invoice_number, u.full_name
		FROM receipts r
		JOIN invoices i ON r.invoice_id = i.id
		JOIN students s ON i.student_id = s.id
		JOIN users u ON s.user_id = u.id
		WHERE r.id = ?
	`
	var receiptID uint
	var receiptNumber string
	var paymentID, invoiceID uint
	var amountPaid, newBalance int64
	var issuedAt time.Time
	var invoiceNumber, fullName string

	err := r.db.QueryRow(query, id).Scan(&receiptID, &receiptNumber, &paymentID, &invoiceID, &amountPaid, &newBalance, &issuedAt,
		&invoiceNumber, &fullName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("receipt not found")
		}
		return nil, err
	}

	return map[string]interface{}{
		"id":             receiptID,
		"receipt_number": receiptNumber,
		"payment_id":     paymentID,
		"invoice_id":     invoiceID,
		"invoice_number": invoiceNumber,
		"amount_paid":    amountPaid,
		"new_balance":    newBalance,
		"issued_at":      issuedAt,
		"student_name":   fullName,
		"description":    "Payment Receipt",
	}, nil
}

func (r *ReceiptRepository) Create(paymentID, invoiceID uint, amountPaid, newBalance int64) (uint, error) {
	receiptNumber := generateReceiptNumber()

	query := `
		INSERT INTO receipts (receipt_number, payment_id, invoice_id, amount_paid, new_balance)
		VALUES (?, ?, ?, ?, ?)
	`
	result, err := r.db.Exec(query, receiptNumber, paymentID, invoiceID, amountPaid, newBalance)
	if err != nil {
		return 0, err
	}
	id, _ := result.LastInsertId()
	return uint(id), nil
}

func generateReceiptNumber() string {
	return "RCT-" + time.Now().Format("2006") + "-" + utils.RandomString(4)
}

func (r *ReceiptRepository) GetByPaymentID(paymentID uint) (map[string]interface{}, error) {
	query := `
		SELECT r.id, r.receipt_number, r.payment_id, r.invoice_id, r.amount_paid, r.new_balance, r.issued_at,
		       i.invoice_number, u.full_name
		FROM receipts r
		JOIN invoices i ON r.invoice_id = i.id
		JOIN students s ON i.student_id = s.id
		JOIN users u ON s.user_id = u.id
		WHERE r.payment_id = ?
	`
	var receiptID uint
	var receiptNumber string
	var pid uint
	var invoiceID uint
	var amountPaid, newBalance int64
	var issuedAt time.Time
	var invoiceNumber, fullName string

	err := r.db.QueryRow(query, paymentID).Scan(&receiptID, &receiptNumber, &pid, &invoiceID, &amountPaid, &newBalance, &issuedAt,
		&invoiceNumber, &fullName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("receipt not found")
		}
		return nil, err
	}

	return map[string]interface{}{
		"id":             receiptID,
		"receipt_number": receiptNumber,
		"payment_id":     pid,
		"invoice_id":     invoiceID,
		"invoice_number": invoiceNumber,
		"amount_paid":    amountPaid,
		"new_balance":    newBalance,
		"issued_at":      issuedAt,
		"student_name":   fullName,
		"description":    "Payment Receipt",
	}, nil
}

func (r *ReceiptRepository) CreateInTx(tx *sql.Tx, paymentID, invoiceID uint, amountPaid, newBalance int64) (uint, error) {
	receiptNumber := generateReceiptNumber()

	query := `
		INSERT INTO receipts (receipt_number, payment_id, invoice_id, amount_paid, new_balance)
		VALUES (?, ?, ?, ?, ?)
	`
	result, err := tx.Exec(query, receiptNumber, paymentID, invoiceID, amountPaid, newBalance)
	if err != nil {
		return 0, err
	}
	id, _ := result.LastInsertId()
	return uint(id), nil
}
