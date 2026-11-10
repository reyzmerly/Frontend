package controllers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	paymentservice "github.com/kwizera/university-invoices/services/payment"
	"github.com/kwizera/university-invoices/shared/repositories"
)

type WebhookController struct {
	paymentRepo *repositories.PaymentRepository
	auditRepo   *repositories.AuditRepository
}

func NewWebhookController(db *sql.DB) *WebhookController {
	return &WebhookController{
		paymentRepo: repositories.NewPaymentRepository(db),
		auditRepo:   repositories.NewAuditRepository(db),
	}
}

func (c *WebhookController) HandleMoPayWebhook(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		TxRef  string `json:"tx_ref"`
		Status string `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, `{"code":"VALIDATION_ERROR","message":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if strings.EqualFold(strings.TrimSpace(payload.Status), "success") || strings.EqualFold(strings.TrimSpace(payload.Status), "paid") {
		if err := paymentservice.CompleteSuccessfulPayment(c.paymentRepo.GetDB(), payload.TxRef, "gateway"); err != nil {
			http.Error(w, `{"code":"INTERNAL_ERROR","message":"Failed to process successful payment"}`, http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		return
	}

	tx, err := c.paymentRepo.GetDB().Begin()
	if err != nil {
		http.Error(w, `{"code":"INTERNAL_ERROR","message":"Database error"}`, http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	payment, err := c.paymentRepo.GetByTransactionIDInTx(tx, payload.TxRef)
	if err != nil {
		w.WriteHeader(http.StatusOK)
		return
	}

	currentStatus := payment["status"].(string)
	if !strings.EqualFold(strings.TrimSpace(payload.Status), "failed") && !strings.EqualFold(strings.TrimSpace(payload.Status), "cancelled") {
		w.WriteHeader(http.StatusOK)
		return
	}

	if currentStatus != "Pending" {
		w.WriteHeader(http.StatusOK)
		return
	}

	newStatus := "Failed"
	err = c.paymentRepo.UpdateStatusInTx(tx, payment["id"].(uint), newStatus, nil)
	if err != nil {
		http.Error(w, `{"code":"INTERNAL_ERROR","message":"Failed to update payment"}`, http.StatusInternalServerError)
		return
	}

	c.auditRepo.LogInTx(tx, "gateway", nil, "Payment status updated", "payment", payment["mopay_reference"].(string), "Pending", newStatus)

	if err := tx.Commit(); err != nil {
		http.Error(w, `{"code":"INTERNAL_ERROR","message":"Failed to commit transaction"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
