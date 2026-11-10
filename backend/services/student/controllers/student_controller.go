package controllers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/kwizera/university-invoices/internal/middleware"
	"github.com/kwizera/university-invoices/shared/repositories"
)

type StudentController struct {
	studentRepo  *repositories.StudentRepository
	invoiceRepo  *repositories.InvoiceRepository
	paymentRepo  *repositories.PaymentRepository
	semesterRepo *repositories.SemesterRepository
}

func NewStudentController(db *sql.DB) *StudentController {
	return &StudentController{
		studentRepo:  repositories.NewStudentRepository(db),
		invoiceRepo:  repositories.NewInvoiceRepository(db),
		paymentRepo:  repositories.NewPaymentRepository(db),
		semesterRepo: repositories.NewSemesterRepository(db),
	}
}

func (c *StudentController) GetDashboard(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(uint)
	studentID, _, _, err := c.studentRepo.GetStudentByUserID(userID)
	if err != nil {
		http.Error(w, `{"code":"NOT_FOUND","message":"Student not found"}`, http.StatusNotFound)
		return
	}

	semesters, _ := c.semesterRepo.ListAll()
	semesterName := ""
	if len(semesters) > 0 {
		semesterName = semesters[0]["name"].(string)
	}

	invoices, _, err := c.invoiceRepo.List(studentID, "", "", "", "", "", "", 1, 5)
	if err != nil {
		invoices = []map[string]interface{}{}
	}

	payments, _, err := c.paymentRepo.List(studentID, "", 1, 1)
	if err != nil {
		payments = []map[string]interface{}{}
	}

	// If no invoices and no payments, return 204 No Content
	if len(invoices) == 0 && len(payments) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	totalInvoiced := int64(0)
	totalPaid := int64(0)
	unpaidCount := 0
	for _, inv := range invoices {
		totalInvoiced += inv["total_amount"].(int64)
		balance := inv["balance"].(int64)
		if balance > 0 {
			unpaidCount++
		}
	}

	for _, pay := range payments {
		if pay["status"].(string) == "Paid" {
			totalPaid += pay["amount"].(int64)
		}
	}

	outstandingBalance := totalInvoiced - totalPaid

	dashboard := map[string]interface{}{
		"semester":             semesterName,
		"outstanding_balance":  outstandingBalance,
		"unpaid_invoice_count": unpaidCount,
		"total_invoiced":       totalInvoiced,
		"invoice_count":        len(invoices),
		"total_paid":           totalPaid,
		"next_due_date":        "",
		"recent_invoices":      invoices,
	}

	if len(payments) > 0 {
		dashboard["latest_payment"] = payments[0]
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dashboard)
}
