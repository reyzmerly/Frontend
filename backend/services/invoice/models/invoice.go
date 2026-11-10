package models

import "time"

type InvoiceStatus string

const (
	InvoiceStatusDraft         InvoiceStatus = "Draft"
	InvoiceStatusIssued        InvoiceStatus = "Issued"
	InvoiceStatusPartiallyPaid InvoiceStatus = "Partially Paid"
	InvoiceStatusPaid          InvoiceStatus = "Paid"
	InvoiceStatusOverdue       InvoiceStatus = "Overdue"
)

type InvoiceRow struct {
	ID            uint          `json:"id"`
	InvoiceNumber string        `json:"invoice_number"`
	Student       StudentInfo   `json:"student"`
	Description   string        `json:"description"`
	TotalAmount   int64         `json:"total_amount"`
	Balance       int64         `json:"balance"`
	DueDate       string        `json:"due_date"`
	Status        InvoiceStatus `json:"status"`
}

type Invoice struct {
	ID            uint                     `json:"id"`
	InvoiceNumber string                   `json:"invoice_number"`
	Student       StudentInfo              `json:"student"`
	Programme     string                   `json:"programme"`
	Semester      string                   `json:"semester"`
	Items         []InvoiceItem            `json:"items"`
	TotalAmount   int64                    `json:"total_amount"`
	PaidAmount    int64                    `json:"paid_amount"`
	Balance       int64                    `json:"balance"`
	DueDate       string                   `json:"due_date"`
	IssuedAt      *time.Time               `json:"issued_at,omitempty"`
	Status        InvoiceStatus            `json:"status"`
	Payments      []map[string]interface{} `json:"payments,omitempty"`
}

type StudentInfo struct {
	ID            uint   `json:"id"`
	FullName      string `json:"full_name"`
	StudentNumber string `json:"student_number"`
}

type InvoiceItem struct {
	FeeTypeID uint   `json:"fee_type_id"`
	FeeName   string `json:"fee_name"`
	Amount    int64  `json:"amount"`
}

type InvoiceInput struct {
	StudentID   uint               `json:"student_id" validate:"required"`
	ProgrammeID uint               `json:"programme_id" validate:"required"`
	SemesterID  uint               `json:"semester_id" validate:"required"`
	DueDate     string             `json:"due_date" validate:"required"`
	Items       []InvoiceItemInput `json:"items" validate:"required,min=1"`
	Action      string             `json:"action" validate:"oneof=draft issue"`
}

type InvoiceItemInput struct {
	FeeTypeID uint  `json:"fee_type_id" validate:"required"`
	Amount    int64 `json:"amount" validate:"required,gt=0"`
}

type IssueInvoiceRequest struct {
	Action string `json:"action" validate:"required,oneof=issue"`
}
