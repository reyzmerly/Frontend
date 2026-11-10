package repositories

import (
	"database/sql"
	"errors"
)

type StudentRepository struct {
	db *sql.DB
}

func NewStudentRepository(db *sql.DB) *StudentRepository {
	return &StudentRepository{db: db}
}

func (r *StudentRepository) GetStudentByUserID(userID uint) (uint, string, uint, error) {
	var id uint
	var studentNumber string
	var programmeID uint

	query := `SELECT id, student_number, programme_id FROM students WHERE user_id = ?`
	err := r.db.QueryRow(query, userID).Scan(&id, &studentNumber, &programmeID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, "", 0, errors.New("student not found")
		}
		return 0, "", 0, err
	}

	return id, studentNumber, programmeID, nil
}

func (r *StudentRepository) SearchStudents(queryStr string, limit int) ([]map[string]interface{}, error) {
	query := `
		SELECT s.id, s.student_number, u.full_name, u.email, p.name as programme
		FROM students s
		JOIN users u ON s.user_id = u.id
		JOIN programmes p ON s.programme_id = p.id
		WHERE s.student_number LIKE ? OR u.full_name LIKE ?
		LIMIT ?
	`

	rows, err := r.db.Query(query, "%"+queryStr+"%", "%"+queryStr+"%", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var students []map[string]interface{}
	for rows.Next() {
		var id uint
		var studentNumber, fullName, email, programme string
		if err := rows.Scan(&id, &studentNumber, &fullName, &email, &programme); err != nil {
			return nil, err
		}
		students = append(students, map[string]interface{}{
			"id":             id,
			"student_number": studentNumber,
			"full_name":      fullName,
			"email":          email,
			"programme":      programme,
		})
	}

	return students, nil
}
