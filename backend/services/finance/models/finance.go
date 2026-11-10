package models

type FinanceDashboard struct {
	TotalInvoiced      int64 `json:"total_invoiced"`
	TotalCollected     int64 `json:"total_collected"`
	OutstandingBalance int64 `json:"outstanding_balance"`
	OverdueAmount      int64 `json:"overdue_amount"`
	PendingPayments    int   `json:"pending_payments"`
	FailedPayments     int   `json:"failed_payments"`
	TotalStudents      int   `json:"total_students"`
	ActiveInvoices     int   `json:"active_invoices"`
}

type ReportSummary struct {
	TotalInvoiced      int64   `json:"total_invoiced"`
	TotalCollected     int64   `json:"total_collected"`
	OutstandingBalance int64   `json:"outstanding_balance"`
	CollectionRate     float64 `json:"collection_rate"`
}

type AccountSummary struct {
	StudentID          uint   `json:"student_id"`
	StudentNumber      string `json:"student_number"`
	StudentName        string `json:"student_name"`
	Programme          string `json:"programme"`
	TotalInvoiced      int64  `json:"total_invoiced"`
	TotalPaid          int64  `json:"total_paid"`
	OutstandingBalance int64  `json:"outstanding_balance"`
	InvoiceCount       int    `json:"invoice_count"`
}

type Programme struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type Semester struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	StartsOn string `json:"starts_on"`
	EndsOn   string `json:"ends_on"`
}
