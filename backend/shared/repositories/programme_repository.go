package repositories

import (
	"database/sql"
)

type ProgrammeRepository struct {
	db *sql.DB
}

func NewProgrammeRepository(db *sql.DB) *ProgrammeRepository {
	return &ProgrammeRepository{db: db}
}

func (r *ProgrammeRepository) ListAll() ([]map[string]interface{}, error) {
	query := `SELECT id, name FROM programmes ORDER BY name`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var programmes []map[string]interface{}
	for rows.Next() {
		var id uint
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			return nil, err
		}
		programmes = append(programmes, map[string]interface{}{
			"id":   id,
			"name": name,
		})
	}

	return programmes, nil
}

func (r *ProgrammeRepository) GetByID(id uint) (string, error) {
	var name string
	query := `SELECT name FROM programmes WHERE id = ?`
	err := r.db.QueryRow(query, id).Scan(&name)
	if err != nil {
		return "", err
	}
	return name, nil
}
