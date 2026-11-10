package repositories

import (
	"database/sql"
)

type SemesterRepository struct {
	db *sql.DB
}

func NewSemesterRepository(db *sql.DB) *SemesterRepository {
	return &SemesterRepository{db: db}
}

func (r *SemesterRepository) ListAll() ([]map[string]interface{}, error) {
	query := `SELECT id, name, starts_on, ends_on FROM semesters ORDER BY starts_on DESC`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var semesters []map[string]interface{}
	for rows.Next() {
		var id uint
		var name, startsOn, endsOn string
		if err := rows.Scan(&id, &name, &startsOn, &endsOn); err != nil {
			return nil, err
		}
		semesters = append(semesters, map[string]interface{}{
			"id":        id,
			"name":      name,
			"starts_on": startsOn,
			"ends_on":   endsOn,
		})
	}

	return semesters, nil
}

func (r *SemesterRepository) GetByID(id uint) (string, string, string, error) {
	var name, startsOn, endsOn string
	query := `SELECT name, starts_on, ends_on FROM semesters WHERE id = ?`
	err := r.db.QueryRow(query, id).Scan(&name, &startsOn, &endsOn)
	if err != nil {
		return "", "", "", err
	}
	return name, startsOn, endsOn, nil
}
