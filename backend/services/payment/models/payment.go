package models

import "time"

type PaymentStatus string

const (
	PaymentStatusPending   PaymentStatus = "Pending"
	PaymentStatusPaid      PaymentStatus = "Paid"
	PaymentStatusFailed    PaymentStatus = "Failed"
	PaymentStatusCancelled PaymentStatus = "Cancelled"
	PaymentStatusExpired   PaymentStatus = "Expired"
)

type PaymentRow struct {
	ID                 uint                   `json:"id"`
	InvoiceID          uint                   `json:"invoice_id"`
	InvoiceNumber      string                 `json:"invoice_number"`
	Student            map[string]interface{} `json:"student"`
	Description        string                 `json:"description"`
	Amount             int64                  `json:"amount"`
	Status             PaymentStatus          `json:"status"`
	MoPayReference     string                 `json:"mopay_reference"`
	MobileNumberMasked string                 `json:"mobile_number_masked"`
	CreatedAt          time.Time              `json:"created_at"`
	StatusUpdatedAt    time.Time              `json:"status_updated_at"`
	ExpiresAt          time.Time              `json:"expires_at"`
	Stuck              bool                   `json:"stuck"`
}

type PaymentDetail struct {
	ID                 uint          `json:"id"`
	InvoiceID          uint          `json:"invoice_id"`
	InvoiceNumber      string        `json:"invoice_number"`
	StudentName        string        `json:"student_name"`
	Description        string        `json:"description"`
	Amount             int64         `json:"amount"`
	Status             PaymentStatus `json:"status"`
	MoPayReference     string        `json:"mopay_reference"`
	MobileNumberMasked string        `json:"mobile_number_masked"`
	CreatedAt          time.Time     `json:"created_at"`
	StatusUpdatedAt    time.Time     `json:"status_updated_at"`
	ExpiresAt          time.Time     `json:"expires_at"`
	Stuck              bool          `json:"stuck"`
	ReceiptID          *uint         `json:"receipt_id,omitempty"`
	NewInvoiceBalance  int64         `json:"new_invoice_balance,omitempty"`
}

type CreatePaymentRequest struct {
	Amount       int64  `json:"amount" validate:"required,gt=0"`
	MobileNumber string `json:"mobile_number" validate:"required"`
}

type PaymentNote struct {
	ID         uint      `json:"id"`
	PaymentID  uint      `json:"payment_id"`
	AuthorID   uint      `json:"author_id"`
	AuthorName string    `json:"author_name"`
	Note       string    `json:"note"`
	CreatedAt  time.Time `json:"created_at"`
}

type CreatePaymentNoteRequest struct {
	Note string `json:"note" validate:"required"`
}
