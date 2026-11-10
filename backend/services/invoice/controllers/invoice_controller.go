package controllers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"

	"github.com/kwizera/university-invoices/internal/middleware"
	"github.com/kwizera/university-invoices/services/invoice/models"
	"github.com/kwizera/university-invoices/shared/repositories"
)

type InvoiceController struct {
	invoiceRepo     *repositories.InvoiceRepository
	invoiceItemRepo *repositories.InvoiceItemRepository
	studentRepo     *repositories.StudentRepository
	validator       *validator.Validate
}

func NewInvoiceController(db *sql.DB) *InvoiceController {
	return &InvoiceController{
		invoiceRepo:     repositories.NewInvoiceRepository(db),
		invoiceItemRepo: repositories.NewInvoiceItemRepository(db),
		studentRepo:     repositories.NewStudentRepository(db),
		validator:       validator.New(),
	}
}

func (c *InvoiceController) ListInvoices(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(uint)
	role := r.Context().Value(middleware.UserRoleKey).(string)

	var studentID uint
	if role == "student" {
		sid, _, _, err := c.studentRepo.GetStudentByUserID(userID)
		if err != nil {
			http.Error(w, `{"code":"NOT_FOUND","message":"Student not found"}`, http.StatusNotFound)
			return
		}
		studentID = sid
	}
	// Finance and admin can see all invoices (studentID = 0 means no filter)

	status := r.URL.Query().Get("status")
	// Students cannot see Draft invoices
	if role == "student" && status == "" {
		status = "Issued,Partially Paid,Paid,Overdue"
	}
	semesterID := r.URL.Query().Get("semester_id")
	programmeID := r.URL.Query().Get("programme_id")
	queryStr := r.URL.Query().Get("q")
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page == 0 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize == 0 {
		pageSize = 20
	}

	invoices, total, err := c.invoiceRepo.List(studentID, status, semesterID, programmeID, queryStr, from, to, page, pageSize)
	if err != nil {
		http.Error(w, `{"code":"INTERNAL_ERROR","message":"Failed to fetch invoices: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	if len(invoices) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	response := map[string]interface{}{
		"data":      invoices,
		"page":      page,
		"page_size": pageSize,
		"total":     total,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (c *InvoiceController) GetInvoice(w http.ResponseWriter, r *http.Request) {
	invoiceID, _ := strconv.ParseUint(chi.URLParam(r, "invoiceId"), 10, 32)
	userID := r.Context().Value(middleware.UserIDKey).(uint)
	role := r.Context().Value(middleware.UserRoleKey).(string)

	invoice, err := c.invoiceRepo.GetByID(uint(invoiceID))
	if err != nil {
		http.Error(w, `{"code":"NOT_FOUND","message":"Invoice not found"}`, http.StatusNotFound)
		return
	}

	if role == "student" {
		sid, _, _, _ := c.studentRepo.GetStudentByUserID(userID)
		if sid != invoice["student_id"].(uint) {
			http.Error(w, `{"code":"FORBIDDEN","message":"You can only view your own invoices"}`, http.StatusForbidden)
			return
		}
		// Students cannot view Draft invoices
		if invoice["status"] == "Draft" {
			http.Error(w, `{"code":"FORBIDDEN","message":"You cannot view draft invoices"}`, http.StatusForbidden)
			return
		}
	}

	items, err := c.invoiceItemRepo.GetByInvoiceID(uint(invoiceID))
	if err != nil {
		http.Error(w, `{"code":"INTERNAL_ERROR","message":"Failed to fetch invoice items"}`, http.StatusInternalServerError)
		return
	}

	invoice["items"] = items

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(invoice)
}

func (c *InvoiceController) CreateInvoice(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(uint)
	role := r.Context().Value(middleware.UserRoleKey).(string)

	// Only finance can create invoices
	if role != "finance" {
		http.Error(w, `{"code":"FORBIDDEN","message":"Only finance can create invoices"}`, http.StatusForbidden)
		return
	}

	var req models.InvoiceInput
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"code":"VALIDATION_ERROR","message":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if err := c.validator.Struct(&req); err != nil {
		http.Error(w, `{"code":"VALIDATION_ERROR","message":"One or more fields are invalid"}`, http.StatusBadRequest)
		return
	}

	totalAmount := int64(0)
	for _, item := range req.Items {
		totalAmount += item.Amount
	}

	invoiceID, err := c.invoiceRepo.Create(req.StudentID, req.ProgrammeID, req.SemesterID, userID, req.DueDate, totalAmount)
	if err != nil {
		// Check for foreign key constraint or duplicate errors
		errMsg := err.Error()
		if strings.Contains(errMsg, "1452") || strings.Contains(errMsg, "foreign key constraint") {
			http.Error(w, `{"code":"CONFLICT","message":"Invalid reference: student, programme, or semester not found"}`, http.StatusConflict)
			return
		}
		if strings.Contains(errMsg, "1062") || strings.Contains(errMsg, "Duplicate entry") {
			http.Error(w, `{"code":"CONFLICT","message":"Invoice number already exists"}`, http.StatusConflict)
			return
		}
		http.Error(w, `{"code":"INTERNAL_ERROR","message":"Failed to create invoice"}`, http.StatusInternalServerError)
		return
	}

	for _, item := range req.Items {
		c.invoiceItemRepo.Create(invoiceID, item.FeeTypeID, "Fee", item.Amount)
	}

	if req.Action == "issue" {
		c.invoiceRepo.Issue(invoiceID)
	}

	invoice, _ := c.invoiceRepo.GetByID(invoiceID)
	items, _ := c.invoiceItemRepo.GetByInvoiceID(invoiceID)
	invoice["items"] = items

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(invoice)
}

func (c *InvoiceController) UpdateInvoice(w http.ResponseWriter, r *http.Request) {
	invoiceID, _ := strconv.ParseUint(chi.URLParam(r, "invoiceId"), 10, 32)

	var req models.InvoiceInput
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"code":"VALIDATION_ERROR","message":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if err := c.validator.Struct(&req); err != nil {
		http.Error(w, `{"code":"VALIDATION_ERROR","message":"One or more fields are invalid"}`, http.StatusBadRequest)
		return
	}

	// Check if invoice exists and is in Draft status
	invoice, err := c.invoiceRepo.GetByID(uint(invoiceID))
	if err != nil {
		http.Error(w, `{"code":"NOT_FOUND","message":"Invoice not found"}`, http.StatusNotFound)
		return
	}

	status, ok := invoice["status"].(string)
	if !ok {
		http.Error(w, `{"code":"INTERNAL_ERROR","message":"Failed to get invoice status"}`, http.StatusInternalServerError)
		return
	}

	if status != "Draft" {
		http.Error(w, `{"code":"CONFLICT","message":"Only draft invoices can be updated"}`, http.StatusConflict)
		return
	}

	totalAmount := int64(0)
	for _, item := range req.Items {
		totalAmount += item.Amount
	}

	updated, err := c.invoiceRepo.Update(uint(invoiceID), req.DueDate, totalAmount)
	if err != nil {
		http.Error(w, `{"code":"INTERNAL_ERROR","message":"Failed to update invoice: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	// Update invoice items regardless of whether header values changed
	c.invoiceItemRepo.DeleteByInvoiceID(uint(invoiceID))
	for _, item := range req.Items {
		c.invoiceItemRepo.Create(uint(invoiceID), item.FeeTypeID, "Fee", item.Amount)
	}

	invoice, _ = c.invoiceRepo.GetByID(uint(invoiceID))
	items, _ := c.invoiceItemRepo.GetByInvoiceID(uint(invoiceID))
	invoice["items"] = items

	w.Header().Set("Content-Type", "application/json")

	// Add a message if values didn't change
	if !updated {
		invoice["message"] = "Invoice updated (no changes to due date or total amount)"
	}

	json.NewEncoder(w).Encode(invoice)
}

func (c *InvoiceController) IssueInvoice(w http.ResponseWriter, r *http.Request) {
	role := r.Context().Value(middleware.UserRoleKey).(string)

	// Only finance can issue invoices
	if role != "finance" {
		http.Error(w, `{"code":"FORBIDDEN","message":"Only finance can issue invoices"}`, http.StatusForbidden)
		return
	}

	invoiceID, _ := strconv.ParseUint(chi.URLParam(r, "invoiceId"), 10, 32)

	err := c.invoiceRepo.Issue(uint(invoiceID))
	if err != nil {
		http.Error(w, `{"code":"INTERNAL_ERROR","message":"Failed to issue invoice"}`, http.StatusInternalServerError)
		return
	}

	invoice, _ := c.invoiceRepo.GetByID(uint(invoiceID))
	items, _ := c.invoiceItemRepo.GetByInvoiceID(uint(invoiceID))
	invoice["items"] = items

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(invoice)
}
