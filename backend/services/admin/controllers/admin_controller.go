package controllers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"golang.org/x/crypto/bcrypt"

	"github.com/kwizera/university-invoices/shared/repositories"
)

type AdminController struct {
	db        *sql.DB
	userRepo  *repositories.UserRepository
	auditRepo *repositories.AuditRepository
	validator *validator.Validate
}

func NewAdminController(db *sql.DB) *AdminController {
	return &AdminController{
		db:        db,
		userRepo:  repositories.NewUserRepository(db),
		auditRepo: repositories.NewAuditRepository(db),
		validator: validator.New(),
	}
}

func (c *AdminController) ListUsers(w http.ResponseWriter, r *http.Request) {
	query := `
		SELECT id, email, full_name, role, access_status, created_at
		FROM users
		ORDER BY created_at DESC
	`
	db := c.userRepo

	rows, err := db.Query(query)
	if err != nil {
		http.Error(w, `{"code":"INTERNAL_ERROR","message":"Failed to fetch users"}`, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var users []map[string]interface{}
	for rows.Next() {
		var id uint
		var email, fullName, role, accessStatus string
		var createdAt string
		if err := rows.Scan(&id, &email, &fullName, &role, &accessStatus, &createdAt); err != nil {
			continue
		}
		users = append(users, map[string]interface{}{
			"id":            id,
			"email":         email,
			"full_name":     fullName,
			"role":          role,
			"access_status": accessStatus,
			"created_at":    createdAt,
		})
	}

	if len(users) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func (c *AdminController) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email         string `json:"email" validate:"required,email"`
		FullName      string `json:"full_name" validate:"required"`
		Password      string `json:"password" validate:"required,min=8"`
		Role          string `json:"role" validate:"required,oneof=student finance admin"`
		StudentNumber string `json:"student_number"`
		ProgrammeID   uint   `json:"programme_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"code":"VALIDATION_ERROR","message":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if err := c.validator.Struct(&req); err != nil {
		http.Error(w, `{"code":"VALIDATION_ERROR","message":"One or more fields are invalid"}`, http.StatusBadRequest)
		return
	}

	// For students, require student_number and programme_id
	if req.Role == "student" {
		if req.StudentNumber == "" {
			http.Error(w, `{"code":"VALIDATION_ERROR","message":"student_number is required for student role"}`, http.StatusBadRequest)
			return
		}
		if req.ProgrammeID == 0 {
			http.Error(w, `{"code":"VALIDATION_ERROR","message":"programme_id is required for student role"}`, http.StatusBadRequest)
			return
		}
	}

	// Check if email already exists
	var count int
	checkQuery := `SELECT COUNT(*) FROM users WHERE email = ?`
	err := c.db.QueryRow(checkQuery, req.Email).Scan(&count)
	if err != nil {
		http.Error(w, `{"code":"INTERNAL_ERROR","message":"Failed to check email"}`, http.StatusInternalServerError)
		return
	}
	if count > 0 {
		http.Error(w, `{"code":"CONFLICT","message":"Email already exists"}`, http.StatusConflict)
		return
	}

	// Check if student_number already exists (for students)
	if req.Role == "student" {
		var studentCount int
		studentCheckQuery := `SELECT COUNT(*) FROM students WHERE student_number = ?`
		err := c.db.QueryRow(studentCheckQuery, req.StudentNumber).Scan(&studentCount)
		if err != nil {
			http.Error(w, `{"code":"INTERNAL_ERROR","message":"Failed to check student number"}`, http.StatusInternalServerError)
			return
		}
		if studentCount > 0 {
			http.Error(w, `{"code":"CONFLICT","message":"Student number already exists"}`, http.StatusConflict)
			return
		}
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, `{"code":"INTERNAL_ERROR","message":"Failed to hash password"}`, http.StatusInternalServerError)
		return
	}

	// Start transaction
	tx, err := c.db.Begin()
	if err != nil {
		http.Error(w, `{"code":"INTERNAL_ERROR","message":"Failed to start transaction"}`, http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Insert user
	query := `
		INSERT INTO users (email, full_name, password_hash, role, access_status)
		VALUES (?, ?, ?, ?, 'invited')
	`
	result, err := tx.Exec(query, req.Email, req.FullName, string(hashedPassword), req.Role)
	if err != nil {
		http.Error(w, `{"code":"INTERNAL_ERROR","message":"Failed to create user"}`, http.StatusInternalServerError)
		return
	}

	userID, _ := result.LastInsertId()

	// For students, also create student record
	if req.Role == "student" {
		studentQuery := `
			INSERT INTO students (user_id, student_number, programme_id)
			VALUES (?, ?, ?)
		`
		_, err = tx.Exec(studentQuery, userID, req.StudentNumber, req.ProgrammeID)
		if err != nil {
			http.Error(w, `{"code":"INTERNAL_ERROR","message":"Failed to create student record"}`, http.StatusInternalServerError)
			return
		}
	}

	// Log audit before commit
	c.auditRepo.LogInTx(tx, "user", nil, "User created", "user", strconv.Itoa(int(userID)), "", req.Email)

	// Commit transaction
	if err := tx.Commit(); err != nil {
		http.Error(w, `{"code":"INTERNAL_ERROR","message":"Failed to commit transaction"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (c *AdminController) UpdateUserAccess(w http.ResponseWriter, r *http.Request) {
	userID, _ := strconv.ParseUint(chi.URLParam(r, "userId"), 10, 32)

	var req struct {
		Role         string `json:"role" validate:"required,oneof=student finance admin"`
		AccessStatus string `json:"access_status" validate:"required,oneof=active suspended invited"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"code":"VALIDATION_ERROR","message":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if err := c.validator.Struct(&req); err != nil {
		http.Error(w, `{"code":"VALIDATION_ERROR","message":"One or more fields are invalid"}`, http.StatusBadRequest)
		return
	}

	query := `
		UPDATE users
		SET role = ?, access_status = ?
		WHERE id = ?
	`
	_, err := c.userRepo.Exec(query, req.Role, req.AccessStatus, userID)
	if err != nil {
		http.Error(w, `{"code":"INTERNAL_ERROR","message":"Failed to update user"}`, http.StatusInternalServerError)
		return
	}

	c.auditRepo.Log("user", nil, "User access updated", "user", strconv.Itoa(int(userID)), "", req.Role)

	w.WriteHeader(http.StatusOK)
}

func (c *AdminController) ListRoles(w http.ResponseWriter, r *http.Request) {
	roles := []map[string]interface{}{
		{
			"role":        "student",
			"description": "Can view their own invoices and payments",
		},
		{
			"role":        "finance",
			"description": "Can manage invoices, payments, and reports",
		},
		{
			"role":        "admin",
			"description": "Full system access including user management",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(roles)
}

func (c *AdminController) ListAuditLogs(w http.ResponseWriter, r *http.Request) {
	recordType := r.URL.Query().Get("record_type")
	recordRef := r.URL.Query().Get("record_ref")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page == 0 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize == 0 {
		pageSize = 20
	}

	logs, total, err := c.auditRepo.List(recordType, recordRef, page, pageSize)
	if err != nil {
		http.Error(w, `{"code":"INTERNAL_ERROR","message":"Failed to fetch audit logs"}`, http.StatusInternalServerError)
		return
	}

	if len(logs) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	response := map[string]interface{}{
		"data":      logs,
		"page":      page,
		"page_size": pageSize,
		"total":     total,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
