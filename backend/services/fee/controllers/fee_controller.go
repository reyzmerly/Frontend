package controllers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"

	"github.com/kwizera/university-invoices/internal/middleware"
	"github.com/kwizera/university-invoices/services/fee/models"
	"github.com/kwizera/university-invoices/shared/repositories"
)

type FeeController struct {
	feeRepo   *repositories.FeeTypeRepository
	validator *validator.Validate
}

func NewFeeController(db *sql.DB) *FeeController {
	return &FeeController{
		feeRepo:   repositories.NewFeeTypeRepository(db),
		validator: validator.New(),
	}
}

func (c *FeeController) ListFeeTypes(w http.ResponseWriter, r *http.Request) {
	feeTypes, err := c.feeRepo.ListAll()
	if err != nil {
		feeTypes = []map[string]interface{}{}
	}

	if len(feeTypes) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(feeTypes)
}

func (c *FeeController) GetFeeType(w http.ResponseWriter, r *http.Request) {
	feeTypeID, _ := strconv.ParseUint(chi.URLParam(r, "feeTypeId"), 10, 32)

	feeType, err := c.feeRepo.GetByID(uint(feeTypeID))
	if err != nil {
		http.Error(w, `{"code":"NOT_FOUND","message":"Fee type not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(feeType)
}

func (c *FeeController) CreateFeeType(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(uint)

	var req models.CreateFeeTypeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"code":"VALIDATION_ERROR","message":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if err := c.validator.Struct(&req); err != nil {
		http.Error(w, `{"code":"VALIDATION_ERROR","message":"One or more fields are invalid"}`, http.StatusBadRequest)
		return
	}

	feeTypeID, err := c.feeRepo.Create(req.Name, req.ProgrammeID, req.SemesterID, userID, req.Amount, req.EffectiveFrom)
	if err != nil {
		http.Error(w, `{"code":"INTERNAL_ERROR","message":"Failed to create fee type"}`, http.StatusInternalServerError)
		return
	}

	feeType, _ := c.feeRepo.GetByID(feeTypeID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(feeType)
}

func (c *FeeController) UpdateFeeType(w http.ResponseWriter, r *http.Request) {
	feeTypeID, _ := strconv.ParseUint(chi.URLParam(r, "feeTypeId"), 10, 32)

	var req models.UpdateFeeTypeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"code":"VALIDATION_ERROR","message":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if err := c.validator.Struct(&req); err != nil {
		http.Error(w, `{"code":"VALIDATION_ERROR","message":"One or more fields are invalid"}`, http.StatusBadRequest)
		return
	}

	updated, err := c.feeRepo.Update(uint(feeTypeID), req.Name, req.Amount, req.EffectiveFrom)
	if err != nil {
		http.Error(w, `{"code":"INTERNAL_ERROR","message":"Failed to update fee type"}`, http.StatusInternalServerError)
		return
	}

	feeType, _ := c.feeRepo.GetByID(uint(feeTypeID))

	w.Header().Set("Content-Type", "application/json")

	// Add a message if values didn't change
	if !updated {
		feeType["message"] = "Fee type updated (no changes to values)"
	}

	json.NewEncoder(w).Encode(feeType)
}
