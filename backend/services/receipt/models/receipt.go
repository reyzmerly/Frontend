package models

import "time"

type Receipt struct {
	ID             uint      `json:"id"`
	ReceiptNumber  string    `json:"receipt_number"`
	PaymentID      uint      `json:"payment_id"`
	InvoiceID      uint      `json:"invoice_id"`
	InvoiceNumber  string    `json:"invoice_number"`
	AmountPaid     int64     `json:"amount_paid"`
	NewBalance     int64     `json:"new_balance"`
	IssuedAt       time.Time `json:"issued_at"`
	StudentName    string    `json:"student_name"`
	Description    string    `json:"description"`
}
