package repositories

import (
	"database/sql"
	"errors"
	"time"
)

type PaymentRepository struct {
	db *sql.DB
}

func NewPaymentRepository(db *sql.DB) *PaymentRepository {
	return &PaymentRepository{db: db}
}

func (r *PaymentRepository) List(studentID uint, status string, page, pageSize int) ([]map[string]interface{}, int, error) {
	offset := (page - 1) * pageSize

	query := `
		SELECT p.id, p.invoice_id, p.student_id, p.amount, p.status, p.mopay_reference,
		       p.mobile_number, p.created_at, p.status_updated_at, p.expires_at,
		       i.invoice_number, u.full_name, s.student_number
		FROM payments p
		JOIN invoices i ON p.invoice_id = i.id
		JOIN students s ON p.student_id = s.id
		JOIN users u ON s.user_id = u.id
		WHERE 1=1
	`
	var args []interface{}

	if studentID > 0 {
		query += " AND p.student_id = ?"
		args = append(args, studentID)
	}

	if status != "" {
		query += " AND p.status = ?"
		args = append(args, status)
	}

	countQuery := "SELECT COUNT(*) FROM payments p JOIN invoices i ON p.invoice_id = i.id WHERE 1=1"
	var total int
	countArgs := args
	if studentID > 0 {
		countQuery += " AND p.student_id = ?"
	}
	if status != "" {
		countQuery += " AND p.status = ?"
	}
	if err := r.db.QueryRow(countQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query += " ORDER BY p.created_at DESC LIMIT ? OFFSET ?"
	args = append(args, pageSize, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var payments []map[string]interface{}
	for rows.Next() {
		var id, invoiceID, studentID uint
		var amount int64
		var status, mopayReference, mobileNumber, invoiceNumber, fullName, studentNumber string
		var createdAt, statusUpdatedAt, expiresAt time.Time

		if err := rows.Scan(&id, &invoiceID, &studentID, &amount, &status, &mopayReference,
			&mobileNumber, &createdAt, &statusUpdatedAt, &expiresAt,
			&invoiceNumber, &fullName, &studentNumber); err != nil {
			return nil, 0, err
		}

		mobileNumberMasked := maskMobileNumber(mobileNumber)
		stuck := isPaymentStuck(status, expiresAt)

		payments = append(payments, map[string]interface{}{
			"id":             id,
			"invoice_id":     invoiceID,
			"invoice_number": invoiceNumber,
			"student": map[string]interface{}{
				"id":             studentID,
				"full_name":      fullName,
				"student_number": studentNumber,
			},
			"description":          "Payment",
			"amount":               amount,
			"status":               status,
			"mopay_reference":      mopayReference,
			"mobile_number_masked": mobileNumberMasked,
			"created_at":           createdAt,
			"status_updated_at":    statusUpdatedAt,
			"expires_at":           expiresAt,
			"stuck":                stuck,
		})
	}

	return payments, total, nil
}

func (r *PaymentRepository) GetByID(id uint) (map[string]interface{}, error) {
	query := `
		SELECT p.id, p.invoice_id, p.student_id, p.amount, p.status, p.mopay_reference,
		       p.mobile_number, p.created_at, p.status_updated_at, p.expires_at, p.paid_at,
		       i.invoice_number, u.full_name
		FROM payments p
		JOIN invoices i ON p.invoice_id = i.id
		JOIN students s ON p.student_id = s.id
		JOIN users u ON s.user_id = u.id
		WHERE p.id = ?
	`
	var paymentID uint
	var invoiceID, studentID uint
	var amount int64
	var status, mopayReference, mobileNumber, invoiceNumber, fullName string
	var createdAt, statusUpdatedAt, expiresAt time.Time
	var paidAt sql.NullTime

	err := r.db.QueryRow(query, id).Scan(&paymentID, &invoiceID, &studentID, &amount, &status, &mopayReference,
		&mobileNumber, &createdAt, &statusUpdatedAt, &expiresAt, &paidAt,
		&invoiceNumber, &fullName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("payment not found")
		}
		return nil, err
	}

	mobileNumberMasked := maskMobileNumber(mobileNumber)
	stuck := isPaymentStuck(status, expiresAt)

	var receiptID *uint
	var newBalance int64
	if status == "Paid" {
		var rid uint
		var nb int64
		queryR := `SELECT id, new_balance FROM receipts WHERE payment_id = ?`
		r.db.QueryRow(queryR, id).Scan(&rid, &nb)
		receiptID = &rid
		newBalance = nb
	}

	return map[string]interface{}{
		"id":                   paymentID,
		"invoice_id":           invoiceID,
		"invoice_number":       invoiceNumber,
		"student_id":           studentID,
		"student_name":         fullName,
		"description":          "Payment",
		"amount":               amount,
		"status":               status,
		"mopay_reference":      mopayReference,
		"mobile_number_masked": mobileNumberMasked,
		"created_at":           createdAt,
		"status_updated_at":    statusUpdatedAt,
		"expires_at":           expiresAt,
		"stuck":                stuck,
		"receipt_id":           receiptID,
		"new_invoice_balance":  newBalance,
	}, nil
}

func (r *PaymentRepository) Create(invoiceID, studentID uint, amount int64, mobileNumber, mopayReference, mopayTransactionID string, expiresAt time.Time) (uint, error) {
	query := `
		INSERT INTO payments (invoice_id, student_id, amount, status, mobile_number, mopay_reference, mopay_transaction_id, expires_at)
		VALUES (?, ?, ?, 'Pending', ?, ?, ?, ?)
	`
	result, err := r.db.Exec(query, invoiceID, studentID, amount, mobileNumber, mopayReference, mopayTransactionID, expiresAt)
	if err != nil {
		return 0, err
	}
	id, _ := result.LastInsertId()
	return uint(id), nil
}

func (r *PaymentRepository) UpdateStatus(id uint, status string, paidAt *time.Time) error {
	query := `
		UPDATE payments
		SET status = ?, status_updated_at = NOW(), paid_at = ?
		WHERE id = ?
	`
	_, err := r.db.Exec(query, status, paidAt, id)
	return err
}

func (r *PaymentRepository) Cancel(id uint) error {
	query := `
		UPDATE payments
		SET status = 'Cancelled', status_updated_at = NOW()
		WHERE id = ? AND status = 'Pending'
	`
	result, err := r.db.Exec(query, id)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.New("payment not found or not in pending status")
	}
	return nil
}

func (r *PaymentRepository) HasPendingPayment(invoiceID uint) (bool, error) {
	var count int
	query := `SELECT COUNT(*) FROM payments WHERE invoice_id = ? AND status = 'Pending'`
	err := r.db.QueryRow(query, invoiceID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func maskMobileNumber(number string) string {
	if len(number) < 4 {
		return number
	}
	return number[:4] + "••• " + number[len(number)-4:]
}

func isPaymentStuck(status string, expiresAt time.Time) bool {
	return status == "Pending" && time.Now().After(expiresAt)
}

func (r *PaymentRepository) BeginTx() (*sql.Tx, error) {
	return r.db.Begin()
}

func (r *PaymentRepository) GetDB() *sql.DB {
	return r.db
}

func (r *PaymentRepository) GetByTransactionIDInTx(tx *sql.Tx, transactionID string) (map[string]interface{}, error) {
	query := `
		SELECT p.id, p.invoice_id, p.student_id, p.amount, p.status, p.mopay_reference,
		       p.mobile_number, p.created_at, p.status_updated_at, p.expires_at, p.paid_at,
		       i.invoice_number, u.full_name
		FROM payments p
		JOIN invoices i ON p.invoice_id = i.id
		JOIN students s ON p.student_id = s.id
		JOIN users u ON s.user_id = u.id
		WHERE p.mopay_transaction_id = ?
		FOR UPDATE
	`
	var id uint
	var invoiceID, studentID uint
	var amount int64
	var status, mopayReference, mobileNumber, invoiceNumber, fullName string
	var createdAt, statusUpdatedAt, expiresAt time.Time
	var paidAt sql.NullTime

	err := tx.QueryRow(query, transactionID).Scan(&id, &invoiceID, &studentID, &amount, &status, &mopayReference,
		&mobileNumber, &createdAt, &statusUpdatedAt, &expiresAt, &paidAt,
		&invoiceNumber, &fullName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("payment not found")
		}
		return nil, err
	}

	return map[string]interface{}{
		"id":              id,
		"invoice_id":      invoiceID,
		"student_id":      studentID,
		"amount":          amount,
		"status":          status,
		"mopay_reference": mopayReference,
		"mobile_number":   mobileNumber,
		"created_at":      createdAt,
		"expires_at":      expiresAt,
	}, nil
}

func (r *PaymentRepository) UpdateStatusInTx(tx *sql.Tx, id uint, status string, paidAt *time.Time) error {
	query := `
		UPDATE payments
		SET status = ?, status_updated_at = NOW(), paid_at = ?
		WHERE id = ?
	`
	_, err := tx.Exec(query, status, paidAt, id)
	return err
}
