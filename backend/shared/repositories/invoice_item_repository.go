package repositories

import (
	"database/sql"
)

type InvoiceItemRepository struct {
	db *sql.DB
}

func NewInvoiceItemRepository(db *sql.DB) *InvoiceItemRepository {
	return &InvoiceItemRepository{db: db}
}

func (r *InvoiceItemRepository) Create(invoiceID, feeTypeID uint, description string, amount int64) error {
	query := `
		INSERT INTO invoice_items (invoice_id, fee_type_id, description, amount)
		VALUES (?, ?, ?, ?)
	`
	_, err := r.db.Exec(query, invoiceID, feeTypeID, description, amount)
	return err
}

func (r *InvoiceItemRepository) GetByInvoiceID(invoiceID uint) ([]map[string]interface{}, error) {
	query := `
		SELECT ii.id, ii.fee_type_id, ii.description, ii.amount, ft.name as fee_name
		FROM invoice_items ii
		JOIN fee_types ft ON ii.fee_type_id = ft.id
		WHERE ii.invoice_id = ?
	`
	rows, err := r.db.Query(query, invoiceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []map[string]interface{}
	for rows.Next() {
		var id, feeTypeID uint
		var description, feeName string
		var amount int64
		if err := rows.Scan(&id, &feeTypeID, &description, &amount, &feeName); err != nil {
			return nil, err
		}
		items = append(items, map[string]interface{}{
			"fee_type_id": feeTypeID,
			"fee_name":    feeName,
			"amount":      amount,
		})
	}

	return items, nil
}

func (r *InvoiceItemRepository) DeleteByInvoiceID(invoiceID uint) error {
	query := `DELETE FROM invoice_items WHERE invoice_id = ?`
	_, err := r.db.Exec(query, invoiceID)
	return err
}
