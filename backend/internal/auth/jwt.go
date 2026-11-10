package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID    uint   `json:"user_id"`
	Role      string `json:"role"`
	StudentID string `json:"student_id,omitempty"`
	jwt.RegisteredClaims
}

type JWTManager struct {
	secretKey string
	expiry    time.Duration
}

func NewJWTManager(secretKey string, expiry time.Duration) *JWTManager {
	return &JWTManager{
		secretKey: secretKey,
		expiry:    expiry,
	}
}

func (m *JWTManager) GenerateToken(userID uint, role string, studentID string) (string, error) {
	claims := &Claims{
		UserID:    userID,
		Role:      role,
		StudentID: studentID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.expiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(m.secretKey))
}

func (m *JWTManager) ValidateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(m.secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, jwt.ErrSignatureInvalid
	}

	return claims, nil
}
