package controllers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"

	"github.com/kwizera/university-invoices/internal/config"
	"github.com/kwizera/university-invoices/internal/middleware"
	"github.com/kwizera/university-invoices/services/hdev"
	"github.com/kwizera/university-invoices/services/payment/models"
	"github.com/kwizera/university-invoices/shared/repositories"
	"github.com/kwizera/university-invoices/shared/utils"
)

type PaymentController struct {
	paymentRepo *repositories.PaymentRepository
	invoiceRepo *repositories.InvoiceRepository
	receiptRepo *repositories.ReceiptRepository
	noteRepo    *repositories.PaymentNoteRepository
	studentRepo *repositories.StudentRepository
	auditRepo   *repositories.AuditRepository
	hdevService *hdev.Service
	validator   *validator.Validate
}

func NewPaymentController(db *sql.DB, cfg *config.Config) *PaymentController {
	return &PaymentController{
		paymentRepo: repositories.NewPaymentRepository(db),
		invoiceRepo: repositories.NewInvoiceRepository(db),
		receiptRepo: repositories.NewReceiptRepository(db),
		noteRepo:    repositories.NewPaymentNoteRepository(db),
		studentRepo: repositories.NewStudentRepository(db),
		auditRepo:   repositories.NewAuditRepository(db),
		hdevService: hdev.NewService(cfg.HDevAPIID, cfg.HDevAPIKey, cfg.PaymentCallbackURL, cfg.PaymentTimeoutMS),
		validator:   validator.New(),
	}
}

func (c *PaymentController) ListPayments(w http.ResponseWriter, r *http.Request) {
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

	status := r.URL.Query().Get("status")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page == 0 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize == 0 {
		pageSize = 20
	}

	payments, total, err := c.paymentRepo.List(studentID, status, page, pageSize)
	if err != nil {
		payments = []map[string]interface{}{}
		total = 0
	}

	if len(payments) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	response := map[string]interface{}{
		"data":      payments,
		"page":      page,
		"page_size": pageSize,
		"total":     total,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (c *PaymentController) GetPayment(w http.ResponseWriter, r *http.Request) {
	paymentID, _ := strconv.ParseUint(chi.URLParam(r, "paymentId"), 10, 32)
	userID := r.Context().Value(middleware.UserIDKey).(uint)
	role := r.Context().Value(middleware.UserRoleKey).(string)

	payment, err := c.paymentRepo.GetByID(uint(paymentID))
	if err != nil {
		http.Error(w, `{"code":"NOT_FOUND","message":"Payment not found"}`, http.StatusNotFound)
		return
	}

	if role == "student" {
		sid, _, _, _ := c.studentRepo.GetStudentByUserID(userID)
		// Get the actual student_id from the payment record
		var actualStudentID uint
		if val, ok := payment["student_id"]; ok && val != nil {
			actualStudentID = val.(uint)
		} else {
			// If student_id is not in the response, query it from the database
			query := `SELECT student_id FROM payments WHERE id = ?`
			err := c.paymentRepo.GetDB().QueryRow(query, paymentID).Scan(&actualStudentID)
			if err != nil {
				http.Error(w, `{"code":"INTERNAL_ERROR","message":"Failed to verify payment ownership"}`, http.StatusInternalServerError)
				return
			}
		}
		if sid != actualStudentID {
			http.Error(w, `{"code":"FORBIDDEN","message":"You can only view your own payments"}`, http.StatusForbidden)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(payment)
}

func (c *PaymentController) CreatePayment(w http.ResponseWriter, r *http.Request) {
	invoiceID, _ := strconv.ParseUint(chi.URLParam(r, "invoiceId"), 10, 32)
	userID := r.Context().Value(middleware.UserIDKey).(uint)
	role := r.Context().Value(middleware.UserRoleKey).(string)

	if role != "student" {
		http.Error(w, `{"code":"FORBIDDEN","message":"Only students can create payments"}`, http.StatusForbidden)
		return
	}

	var req models.CreatePaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"code":"VALIDATION_ERROR","message":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if err := c.validator.Struct(&req); err != nil {
		http.Error(w, `{"code":"VALIDATION_ERROR","message":"One or more fields are invalid"}`, http.StatusBadRequest)
		return
	}

	sid, _, _, err := c.studentRepo.GetStudentByUserID(userID)
	if err != nil {
		http.Error(w, `{"code":"NOT_FOUND","message":"Student not found"}`, http.StatusNotFound)
		return
	}

	invoice, err := c.invoiceRepo.GetByID(uint(invoiceID))
	if err != nil {
		http.Error(w, `{"code":"NOT_FOUND","message":"Invoice not found"}`, http.StatusNotFound)
		return
	}

	if invoice["student_id"].(uint) != sid {
		http.Error(w, `{"code":"FORBIDDEN","message":"Invoice does not belong to you"}`, http.StatusForbidden)
		return
	}

	status := invoice["status"].(string)
	if status != "Issued" && status != "Partially Paid" && status != "Overdue" {
		http.Error(w, `{"code":"CONFLICT","message":"Invoice cannot be paid"}`, http.StatusConflict)
		return
	}

	balance := invoice["balance"].(int64)
	if req.Amount < 1 || req.Amount > balance {
		http.Error(w, `{"code":"VALIDATION_ERROR","message":"Amount must be between 1 and the outstanding balance"}`, http.StatusBadRequest)
		return
	}

	hasPending, _ := c.paymentRepo.HasPendingPayment(uint(invoiceID))
	if hasPending {
		http.Error(w, `{"code":"CONFLICT","message":"A pending payment already exists for this invoice"}`, http.StatusConflict)
		return
	}

	hdevReference := "HDEV-" + time.Now().Format("20060102") + "-" + utils.RandomString(4)
	transactionRef := hdevReference
	expiresAt := time.Now().Add(2 * time.Minute)

	// Normalize phone number (remove '+' and non-digits)
	normalizedPhone := req.MobileNumber
	normalizedPhone = strings.ReplaceAll(normalizedPhone, "+", "")
	normalizedPhone = strings.ReplaceAll(normalizedPhone, " ", "")

	// Call HDEV API to initiate payment first
	hdevReq := hdev.InitiatePaymentRequest{
		Ref:    "pay",
		Tel:    normalizedPhone,
		TxRef:  transactionRef,
		Amount: int(req.Amount),
		Link:   c.hdevService.CallbackURL(),
	}

	hdevResp, err := c.hdevService.InitiatePayment(hdevReq)
	if err != nil {
		http.Error(w, `{"code":"INTERNAL_ERROR","message":"Unable to initiate payment with the payment provider. Please try again later."}`, http.StatusBadGateway)
		return
	}

	// Use the transaction reference returned by HDEV for status checking
	actualTxRef := hdevResp.TxRef
	if actualTxRef == "" {
		actualTxRef = transactionRef
	}

	// Only create payment record after successful HDEV initiation
	_, err = c.paymentRepo.Create(uint(invoiceID), sid, req.Amount, req.MobileNumber, hdevReference, actualTxRef, expiresAt)
	if err != nil {
		http.Error(w, `{"code":"INTERNAL_ERROR","message":"Failed to create payment"}`, http.StatusInternalServerError)
		return
	}

	c.auditRepo.Log("user", &userID, "Payment created", "payment", hdevReference, "", "Pending")

	w.WriteHeader(http.StatusAccepted)
}

func (c *PaymentController) CancelPayment(w http.ResponseWriter, r *http.Request) {
	paymentID, _ := strconv.ParseUint(chi.URLParam(r, "paymentId"), 10, 32)
	userID := r.Context().Value(middleware.UserIDKey).(uint)

	payment, err := c.paymentRepo.GetByID(uint(paymentID))
	if err != nil {
		http.Error(w, `{"code":"NOT_FOUND","message":"Payment not found"}`, http.StatusNotFound)
		return
	}

	sid, _, _, _ := c.studentRepo.GetStudentByUserID(userID)
	if sid != payment["student_id"].(uint) {
		http.Error(w, `{"code":"FORBIDDEN","message":"You can only cancel your own payments"}`, http.StatusForbidden)
		return
	}

	err = c.paymentRepo.Cancel(uint(paymentID))
	if err != nil {
		if err.Error() == "payment not found or not in pending status" {
			http.Error(w, `{"code":"CONFLICT","message":"Payment cannot be cancelled"}`, http.StatusConflict)
			return
		}
		http.Error(w, `{"code":"INTERNAL_ERROR","message":"Failed to cancel payment"}`, http.StatusInternalServerError)
		return
	}

	c.auditRepo.Log("user", &userID, "Payment cancelled", "payment", payment["mopay_reference"].(string), "Pending", "Cancelled")

	w.WriteHeader(http.StatusNoContent)
}

func (c *PaymentController) ListPaymentNotes(w http.ResponseWriter, r *http.Request) {
	paymentID, _ := strconv.ParseUint(chi.URLParam(r, "paymentId"), 10, 32)
	role := r.Context().Value(middleware.UserRoleKey).(string)

	if role != "finance" {
		http.Error(w, `{"code":"FORBIDDEN","message":"Only finance can view payment notes"}`, http.StatusForbidden)
		return
	}

	notes, err := c.noteRepo.List(uint(paymentID))
	if err != nil {
		http.Error(w, `{"code":"INTERNAL_ERROR","message":"Failed to fetch notes"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(notes)
}

func (c *PaymentController) CreatePaymentNote(w http.ResponseWriter, r *http.Request) {
	paymentID, _ := strconv.ParseUint(chi.URLParam(r, "paymentId"), 10, 32)
	userID := r.Context().Value(middleware.UserIDKey).(uint)
	role := r.Context().Value(middleware.UserRoleKey).(string)

	if role != "finance" {
		http.Error(w, `{"code":"FORBIDDEN","message":"Only finance can add payment notes"}`, http.StatusForbidden)
		return
	}

	var req models.CreatePaymentNoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"code":"VALIDATION_ERROR","message":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if err := c.validator.Struct(&req); err != nil {
		http.Error(w, `{"code":"VALIDATION_ERROR","message":"One or more fields are invalid"}`, http.StatusBadRequest)
		return
	}

	err := c.noteRepo.Create(uint(paymentID), userID, req.Note)
	if err != nil {
		http.Error(w, `{"code":"INTERNAL_ERROR","message":"Failed to create note"}`, http.StatusInternalServerError)
		return
	}

	c.auditRepo.Log("user", &userID, "Payment note added", "payment", strconv.Itoa(int(paymentID)), "", req.Note)

	w.WriteHeader(http.StatusCreated)
}
