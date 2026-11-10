package repositories

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/kwizera/university-invoices/shared/utils"
)

type InvoiceRepository struct {
	db *sql.DB
}

func NewInvoiceRepository(db *sql.DB) *InvoiceRepository {
	return &InvoiceRepository{db: db}
}

func (r *InvoiceRepository) List(studentID uint, status, semesterID, programmeID, queryStr, from, to string, page, pageSize int) ([]map[string]interface{}, int, error) {
	offset := (page - 1) * pageSize

	query := `
		SELECT i.id, i.invoice_number, i.student_id, i.programme_id, i.semester_id,
		       i.total_amount, i.paid_amount, i.due_date, i.status, i.issued_at,
		       u.full_name, s.student_number, p.name as programme, sem.name as semester
		FROM invoices i
		JOIN students s ON i.student_id = s.id
		JOIN users u ON s.user_id = u.id
		JOIN programmes p ON i.programme_id = p.id
		JOIN semesters sem ON i.semester_id = sem.id
		WHERE 1=1
	`
	var args []interface{}
	argCount := 0

	if studentID > 0 {
		argCount++
		query += " AND i.student_id = ?"
		args = append(args, studentID)
	}

	if status != "" {
		argCount++
		// Handle multiple statuses separated by comma
		if contains(status, ",") {
			query += " AND i.status IN ("
			statuses := splitByComma(status)
			for i, s := range statuses {
				if i > 0 {
					query += ","
				}
				query += "?"
				args = append(args, s)
			}
			query += ")"
		} else {
			query += " AND i.status = ?"
			args = append(args, status)
		}
	}

	if semesterID != "" {
		argCount++
		query += " AND i.semester_id = ?"
		args = append(args, semesterID)
	}

	if programmeID != "" {
		argCount++
		query += " AND i.programme_id = ?"
		args = append(args, programmeID)
	}

	if queryStr != "" {
		argCount++
		query += " AND (s.student_number LIKE ? OR u.full_name LIKE ? OR i.invoice_number LIKE ?)"
		args = append(args, "%"+queryStr+"%", "%"+queryStr+"%", "%"+queryStr+"%")
	}

	if from != "" {
		argCount++
		query += " AND i.due_date >= ?"
		args = append(args, from)
	}

	if to != "" {
		argCount++
		query += " AND i.due_date <= ?"
		args = append(args, to)
	}

	countQuery := `
		SELECT COUNT(*) FROM invoices i
		JOIN students s ON i.student_id = s.id
		JOIN users u ON s.user_id = u.id
		WHERE 1=1
	`
	var countArgs []interface{}
	if studentID > 0 {
		countQuery += " AND i.student_id = ?"
		countArgs = append(countArgs, studentID)
	}
	if status != "" {
		// Handle multiple statuses separated by comma
		if contains(status, ",") {
			countQuery += " AND i.status IN ("
			statuses := splitByComma(status)
			for i, s := range statuses {
				if i > 0 {
					countQuery += ","
				}
				countQuery += "?"
				countArgs = append(countArgs, s)
			}
			countQuery += ")"
		} else {
			countQuery += " AND i.status = ?"
			countArgs = append(countArgs, status)
		}
	}
	if semesterID != "" {
		countQuery += " AND i.semester_id = ?"
		countArgs = append(countArgs, semesterID)
	}
	if programmeID != "" {
		countQuery += " AND i.programme_id = ?"
		countArgs = append(countArgs, programmeID)
	}
	if queryStr != "" {
		countQuery += " AND (s.student_number LIKE ? OR u.full_name LIKE ? OR i.invoice_number LIKE ?)"
		countArgs = append(countArgs, "%"+queryStr+"%", "%"+queryStr+"%", "%"+queryStr+"%")
	}
	if from != "" {
		countQuery += " AND i.due_date >= ?"
		countArgs = append(countArgs, from)
	}
	if to != "" {
		countQuery += " AND i.due_date <= ?"
		countArgs = append(countArgs, to)
	}

	var total int
	if err := r.db.QueryRow(countQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query += " ORDER BY i.created_at DESC LIMIT ? OFFSET ?"
	args = append(args, pageSize, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var invoices []map[string]interface{}
	for rows.Next() {
		var id, studentID, programmeID, semesterID uint
		var invoiceNumber string
		var totalAmount, paidAmount int64
		var dueDate string
		var status string
		var issuedAt sql.NullTime
		var fullName, studentNumber, programme, semester string

		if err := rows.Scan(&id, &invoiceNumber, &studentID, &programmeID, &semesterID,
			&totalAmount, &paidAmount, &dueDate, &status, &issuedAt,
			&fullName, &studentNumber, &programme, &semester); err != nil {
			return nil, 0, err
		}

		balance := totalAmount - paidAmount
		description := "Invoice"

		invoices = append(invoices, map[string]interface{}{
			"id":             id,
			"invoice_number": invoiceNumber,
			"student": map[string]interface{}{
				"id":             studentID,
				"full_name":      fullName,
				"student_number": studentNumber,
			},
			"description":  description,
			"total_amount": totalAmount,
			"paid_amount":  paidAmount,
			"balance":      balance,
			"due_date":     dueDate,
			"status":       status,
		})
	}

	return invoices, total, nil
}

func (r *InvoiceRepository) GetByID(id uint) (map[string]interface{}, error) {
	query := `
		SELECT i.id, i.invoice_number, i.student_id, i.programme_id, i.semester_id,
		       i.total_amount, i.paid_amount, i.due_date, i.status, i.issued_at,
		       u.full_name, s.student_number, p.name as programme, sem.name as semester
		FROM invoices i
		JOIN students s ON i.student_id = s.id
		JOIN users u ON s.user_id = u.id
		JOIN programmes p ON i.programme_id = p.id
		JOIN semesters sem ON i.semester_id = sem.id
		WHERE i.id = ?
	`
	var invoiceID uint
	var studentID, programmeID, semesterID uint
	var invoiceNumber string
	var totalAmount, paidAmount int64
	var dueDate string
	var status string
	var issuedAt sql.NullTime
	var fullName, studentNumber, programme, semester string

	err := r.db.QueryRow(query, id).Scan(&invoiceID, &invoiceNumber, &studentID, &programmeID, &semesterID,
		&totalAmount, &paidAmount, &dueDate, &status, &issuedAt,
		&fullName, &studentNumber, &programme, &semester)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("invoice not found")
		}
		return nil, err
	}

	balance := totalAmount - paidAmount

	return map[string]interface{}{
		"id":             invoiceID,
		"invoice_number": invoiceNumber,
		"student_id":     studentID,
		"programme_id":   programmeID,
		"semester_id":    semesterID,
		"total_amount":   totalAmount,
		"paid_amount":    paidAmount,
		"balance":        balance,
		"due_date":       dueDate,
		"status":         status,
		"issued_at":      issuedAt,
		"full_name":      fullName,
		"student_number": studentNumber,
		"programme":      programme,
		"semester":       semester,
	}, nil
}

func (r *InvoiceRepository) Create(studentID, programmeID, semesterID, createdBy uint, dueDate string, totalAmount int64) (uint, error) {
	invoiceNumber := generateInvoiceNumber(dueDate)

	query := `
		INSERT INTO invoices (invoice_number, student_id, programme_id, semester_id, due_date, total_amount, created_by)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	result, err := r.db.Exec(query, invoiceNumber, studentID, programmeID, semesterID, dueDate, totalAmount, createdBy)
	if err != nil {
		return 0, err
	}
	id, _ := result.LastInsertId()
	return uint(id), nil
}

func (r *InvoiceRepository) Update(id uint, dueDate string, totalAmount int64) (bool, error) {
	// Perform the update without status check in WHERE clause
	// The status check will be done in application logic
	query := fmt.Sprintf(`
		UPDATE invoices
		SET due_date = ?, total_amount = ?
		WHERE id = %d
	`, id)
	result, err := r.db.Exec(query, dueDate, totalAmount)
	if err != nil {
		return false, fmt.Errorf("update query failed: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("failed to get rows affected: %w", err)
	}
	// Return true if rows were affected (values changed), false otherwise
	return rows > 0, nil
}

func (r *InvoiceRepository) Issue(id uint) error {
	query := `
		UPDATE invoices
		SET status = 'Issued', issued_at = NOW()
		WHERE id = ? AND status = 'Draft'
	`
	result, err := r.db.Exec(query, id)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.New("invoice not found or not in draft status")
	}
	return nil
}

func (r *InvoiceRepository) UpdateBalance(id uint, paidAmount int64, status string) error {
	query := `
		UPDATE invoices
		SET paid_amount = paid_amount + ?, status = ?
		WHERE id = ?
	`
	_, err := r.db.Exec(query, paidAmount, status, id)
	return err
}

func generateInvoiceNumber(dueDate string) string {
	return "INV-" + dueDate[2:4] + "-" + utils.RandomString(4)
}

func (r *InvoiceRepository) GetNextInvoiceNumber() (string, error) {
	var maxNum int
	query := `SELECT CAST(SUBSTRING(invoice_number, 10) AS UNSIGNED) as num FROM invoices WHERE invoice_number LIKE 'INV-%-%' ORDER BY num DESC LIMIT 1`
	err := r.db.QueryRow(query).Scan(&maxNum)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "INV-26-0001", nil
		}
		return "", err
	}
	nextNum := maxNum + 1
	return fmt.Sprintf("INV-26-%04d", nextNum), nil
}

func (r *InvoiceRepository) UpdateBalanceInTx(tx *sql.Tx, id uint, paidAmount int64) error {
	query := `
		UPDATE invoices
		SET paid_amount = paid_amount + ?,
		    status = CASE
		        WHEN paid_amount + ? >= total_amount THEN 'Paid'
		        WHEN paid_amount + ? > 0 THEN 'Partially Paid'
		        ELSE status
		    END
		WHERE id = ?
	`
	_, err := tx.Exec(query, paidAmount, paidAmount, paidAmount, id)
	return err
}

func (r *InvoiceRepository) GetByIDInTx(tx *sql.Tx, id uint) (map[string]interface{}, error) {
	query := `
		SELECT i.id, i.invoice_number, i.student_id, i.programme_id, i.semester_id,
		       i.total_amount, i.paid_amount, i.due_date, i.status, i.issued_at,
		       u.full_name, s.student_number, p.name as programme, sem.name as semester
		FROM invoices i
		JOIN students s ON i.student_id = s.id
		JOIN users u ON s.user_id = u.id
		JOIN programmes p ON i.programme_id = p.id
		JOIN semesters sem ON i.semester_id = sem.id
		WHERE i.id = ?
	`
	var invoiceID uint
	var studentID, programmeID, semesterID uint
	var invoiceNumber string
	var totalAmount, paidAmount int64
	var dueDate string
	var status string
	var issuedAt sql.NullTime
	var fullName, studentNumber, programme, semester string

	err := tx.QueryRow(query, id).Scan(&invoiceID, &invoiceNumber, &studentID, &programmeID, &semesterID,
		&totalAmount, &paidAmount, &dueDate, &status, &issuedAt,
		&fullName, &studentNumber, &programme, &semester)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("invoice not found")
		}
		return nil, err
	}

	balance := totalAmount - paidAmount

	return map[string]interface{}{
		"id":             invoiceID,
		"invoice_number": invoiceNumber,
		"student_id":     studentID,
		"programme_id":   programmeID,
		"semester_id":    semesterID,
		"total_amount":   totalAmount,
		"paid_amount":    paidAmount,
		"balance":        balance,
		"due_date":       dueDate,
		"status":         status,
		"issued_at":      issuedAt,
		"full_name":      fullName,
		"student_number": studentNumber,
		"programme":      programme,
		"semester":       semester,
	}, nil
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func splitByComma(s string) []string {
	var result []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == ',' {
			result = append(result, s[start:i])
			start = i + 1
		}
	}
	result = append(result, s[start:])
	return result
}
