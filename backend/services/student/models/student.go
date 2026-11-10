package models

import "time"

type Student struct {
	ID            uint   `json:"id"`
	UserID        uint   `json:"user_id"`
	StudentNumber string `json:"student_number"`
	ProgrammeID   uint   `json:"programme_id"`
	FullName      string `json:"full_name"`
	Email         string `json:"email"`
	Role          string `json:"role"`
	AccessStatus  string `json:"access_status"`
}

type StudentDashboard struct {
	Semester           string         `json:"semester"`
	OutstandingBalance int64          `json:"outstanding_balance"`
	UnpaidInvoiceCount int            `json:"unpaid_invoice_count"`
	TotalInvoiced      int64          `json:"total_invoiced"`
	InvoiceCount       int            `json:"invoice_count"`
	TotalPaid          int64          `json:"total_paid"`
	NextDueDate        string         `json:"next_due_date"`
	LatestPayment      *PaymentDetail `json:"latest_payment,omitempty"`
	RecentInvoices     []InvoiceInfo  `json:"recent_invoices"`
}

type InvoiceInfo struct {
	ID            uint   `json:"id"`
	InvoiceNumber string `json:"invoice_number"`
	Description   string `json:"description"`
	TotalAmount   int64  `json:"total_amount"`
	Balance       int64  `json:"balance"`
	DueDate       string `json:"due_date"`
	Status        string `json:"status"`
}

type PaymentDetail struct {
	ID                 uint      `json:"id"`
	InvoiceID          uint      `json:"invoice_id"`
	InvoiceNumber      string    `json:"invoice_number"`
	StudentName        string    `json:"student_name"`
	Description        string    `json:"description"`
	Amount             int64     `json:"amount"`
	Status             string    `json:"status"`
	MoPayReference     string    `json:"mopay_reference"`
	MobileNumberMasked string    `json:"mobile_number_masked"`
	CreatedAt          time.Time `json:"created_at"`
	StatusUpdatedAt    time.Time `json:"status_updated_at"`
	ExpiresAt          time.Time `json:"expires_at"`
	Stuck              bool      `json:"stuck"`
	ReceiptID          *uint     `json:"receipt_id,omitempty"`
	NewInvoiceBalance  int64     `json:"new_invoice_balance,omitempty"`
}
