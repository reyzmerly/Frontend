package repositories

import (
	"database/sql"
	"time"
)

type PaymentNoteRepository struct {
	db *sql.DB
}

func NewPaymentNoteRepository(db *sql.DB) *PaymentNoteRepository {
	return &PaymentNoteRepository{db: db}
}

func (r *PaymentNoteRepository) List(paymentID uint) ([]map[string]interface{}, error) {
	query := `
		SELECT pn.id, pn.payment_id, pn.author_id, pn.note, pn.created_at, u.full_name
		FROM payment_notes pn
		JOIN users u ON pn.author_id = u.id
		WHERE pn.payment_id = ?
		ORDER BY pn.created_at ASC
	`
	rows, err := r.db.Query(query, paymentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notes []map[string]interface{}
	for rows.Next() {
		var id, paymentID, authorID uint
		var note, authorName string
		var createdAt time.Time
		if err := rows.Scan(&id, &paymentID, &authorID, &note, &createdAt, &authorName); err != nil {
			return nil, err
		}
		notes = append(notes, map[string]interface{}{
			"id":          id,
			"payment_id":  paymentID,
			"author_id":   authorID,
			"author_name": authorName,
			"note":        note,
			"created_at":  createdAt,
		})
	}

	return notes, nil
}

func (r *PaymentNoteRepository) Create(paymentID, authorID uint, note string) error {
	query := `
		INSERT INTO payment_notes (payment_id, author_id, note)
		VALUES (?, ?, ?)
	`
	_, err := r.db.Exec(query, paymentID, authorID, note)
	return err
}
