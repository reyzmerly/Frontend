package repositories

import (
	"database/sql"
	"errors"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) GetByEmailOrStudentID(identifier string) (uint, string, string, string, string, string, error) {
	var id uint
	var email, fullName, passwordHash, role, accessStatus string

	query := `
		SELECT id, email, full_name, password_hash, role, access_status
		FROM users
		WHERE email = ? OR id IN (
			SELECT user_id FROM students WHERE student_number = ?
		)
	`
	err := r.db.QueryRow(query, identifier, identifier).Scan(
		&id, &email, &fullName, &passwordHash, &role, &accessStatus,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, "", "", "", "", "", errors.New("user not found")
		}
		return 0, "", "", "", "", "", err
	}

	return id, email, fullName, passwordHash, role, accessStatus, nil
}

func (r *UserRepository) GetByID(id uint) (uint, string, string, string, string, string, error) {
	var email, fullName, passwordHash, role, accessStatus string

	query := `SELECT id, email, full_name, password_hash, role, access_status FROM users WHERE id = ?`
	err := r.db.QueryRow(query, id).Scan(
		&id, &email, &fullName, &passwordHash, &role, &accessStatus,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, "", "", "", "", "", errors.New("user not found")
		}
		return 0, "", "", "", "", "", err
	}

	return id, email, fullName, passwordHash, role, accessStatus, nil
}

func (r *UserRepository) GetStudentIDByUserID(userID uint) (string, error) {
	var studentNumber string
	query := `SELECT student_number FROM students WHERE user_id = ?`
	err := r.db.QueryRow(query, userID).Scan(&studentNumber)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
		return "", err
	}
	return studentNumber, nil
}

func (r *UserRepository) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return r.db.Query(query, args...)
}

func (r *UserRepository) Exec(query string, args ...interface{}) (sql.Result, error) {
	return r.db.Exec(query, args...)
}
