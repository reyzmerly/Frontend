package models

import "time"

type Receipt struct {
	ID            int       `json:"id"`
	ReceiptNumber string    `json:"receipt_number"`
	PaymentID     int       `json:"payment_id"`
	InvoiceID     int       `json:"invoice_id"`
	StudentID     int       `json:"student_id"`
	Amount        float64   `json:"amount"`
	IssuedDate    time.Time `json:"issued_date"`
	IssuedBy      int       `json:"issued_by"`
	CreatedAt     time.Time `json:"created_at"`
}
