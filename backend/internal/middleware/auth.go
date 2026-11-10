package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const (
	UserIDKey    contextKey = "userID"
	UserRoleKey  contextKey = "userRole"
	StudentIDKey contextKey = "studentID"
)

type Claims struct {
	UserID    uint   `json:"user_id"`
	Role      string `json:"role"`
	StudentID string `json:"student_id,omitempty"`
	jwt.RegisteredClaims
}

type AuthMiddleware struct {
	jwtSecret string
}

func NewAuthMiddleware(jwtSecret string) *AuthMiddleware {
	return &AuthMiddleware{
		jwtSecret: jwtSecret,
	}
}

func (am *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, `{"code":"UNAUTHORIZED","message":"Missing authorization header"}`, http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, `{"code":"UNAUTHORIZED","message":"Invalid authorization header format"}`, http.StatusUnauthorized)
			return
		}

		tokenString := parts[1]
		claims := &Claims{}

		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(am.jwtSecret), nil
		})

		if err != nil || !token.Valid {
			http.Error(w, `{"code":"UNAUTHORIZED","message":"Invalid or expired token"}`, http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, UserRoleKey, claims.Role)
		if claims.StudentID != "" {
			ctx = context.WithValue(ctx, StudentIDKey, claims.StudentID)
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func RequireRole(allowedRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role := r.Context().Value(UserRoleKey)
			if role == nil {
				http.Error(w, `{"code":"FORBIDDEN","message":"User role not found"}`, http.StatusForbidden)
				return
			}

			roleStr, ok := role.(string)
			if !ok {
				http.Error(w, `{"code":"FORBIDDEN","message":"Invalid role format"}`, http.StatusForbidden)
				return
			}

			allowed := false
			for _, allowedRole := range allowedRoles {
				if roleStr == allowedRole {
					allowed = true
					break
				}
			}

			if !allowed {
				http.Error(w, `{"code":"FORBIDDEN","message":"Role not allowed"}`, http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
