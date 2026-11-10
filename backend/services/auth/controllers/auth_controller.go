package controllers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"golang.org/x/crypto/bcrypt"

	"github.com/kwizera/university-invoices/internal/auth"
	"github.com/kwizera/university-invoices/internal/middleware"
	"github.com/kwizera/university-invoices/services/auth/models"
	"github.com/kwizera/university-invoices/shared/repositories"
)

type AuthController struct {
	userRepo    *repositories.UserRepository
	studentRepo *repositories.StudentRepository
	jwtManager  *auth.JWTManager
	validator   *validator.Validate
}

func NewAuthController(db *sql.DB, jwtSecret string) *AuthController {
	return &AuthController{
		userRepo:    repositories.NewUserRepository(db),
		studentRepo: repositories.NewStudentRepository(db),
		jwtManager:  auth.NewJWTManager(jwtSecret, 24*time.Hour),
		validator:   validator.New(),
	}
}

func (c *AuthController) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"code":"VALIDATION_ERROR","message":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if err := c.validator.Struct(&req); err != nil {
		http.Error(w, `{"code":"VALIDATION_ERROR","message":"One or more fields are invalid"}`, http.StatusBadRequest)
		return
	}

	userID, email, fullName, passwordHash, role, accessStatus, err := c.userRepo.GetByEmailOrStudentID(req.Identifier)
	if err != nil {
		http.Error(w, `{"code":"UNAUTHORIZED","message":"Invalid credentials"}`, http.StatusUnauthorized)
		return
	}

	if accessStatus != "active" {
		http.Error(w, `{"code":"FORBIDDEN","message":"Account is not active"}`, http.StatusForbidden)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password)); err != nil {
		http.Error(w, `{"code":"UNAUTHORIZED","message":"Invalid credentials"}`, http.StatusUnauthorized)
		return
	}

	_, studentNumber, _, _ := c.studentRepo.GetStudentByUserID(userID)

	token, err := c.jwtManager.GenerateToken(userID, role, studentNumber)
	if err != nil {
		http.Error(w, `{"code":"INTERNAL_ERROR","message":"Failed to generate token"}`, http.StatusInternalServerError)
		return
	}

	response := models.LoginResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   86400,
		User: models.User{
			ID:            userID,
			Email:         email,
			FullName:      fullName,
			Role:          role,
			AccessStatus:  accessStatus,
			StudentNumber: studentNumber,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (c *AuthController) Logout(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func (c *AuthController) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req models.ForgotPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"code":"VALIDATION_ERROR","message":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	if err := c.validator.Struct(&req); err != nil {
		http.Error(w, `{"code":"VALIDATION_ERROR","message":"One or more fields are invalid"}`, http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (c *AuthController) GetMe(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(uint)
	role := r.Context().Value(middleware.UserRoleKey).(string)
	studentID, hasStudentID := r.Context().Value(middleware.StudentIDKey).(string)

	_, email, fullName, _, accessStatus, _, err := c.userRepo.GetByID(userID)
	if err != nil {
		http.Error(w, `{"code":"NOT_FOUND","message":"User not found"}`, http.StatusNotFound)
		return
	}

	user := models.User{
		ID:           userID,
		Email:        email,
		FullName:     fullName,
		Role:         role,
		AccessStatus: accessStatus,
	}

	if hasStudentID {
		user.StudentNumber = studentID
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}
