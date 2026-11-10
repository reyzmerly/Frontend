package models

import "time"

type User struct {
	ID            uint      `json:"id"`
	Email         string    `json:"email"`
	FullName      string    `json:"full_name"`
	PasswordHash  string    `json:"-"`
	Role          string    `json:"role"`
	AccessStatus  string    `json:"access_status"`
	StudentNumber string    `json:"student_number,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type LoginRequest struct {
	Identifier string `json:"identifier" validate:"required"`
	Password   string `json:"password" validate:"required"`
}

type LoginResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	User        User   `json:"user"`
}

type ForgotPasswordRequest struct {
	Identifier string `json:"identifier" validate:"required"`
}
