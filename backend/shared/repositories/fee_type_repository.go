package repositories

import (
	"database/sql"
	"errors"
	"time"
)

type FeeTypeRepository struct {
	db *sql.DB
}

func NewFeeTypeRepository(db *sql.DB) *FeeTypeRepository {
	return &FeeTypeRepository{db: db}
}

func (r *FeeTypeRepository) ListAll() ([]map[string]interface{}, error) {
	query := `
		SELECT ft.id, ft.name, ft.programme_id, ft.semester_id, ft.amount, ft.effective_from,
		       ft.created_by, ft.created_at, p.name as programme_name, s.name as semester_name
		FROM fee_types ft
		JOIN programmes p ON ft.programme_id = p.id
		JOIN semesters s ON ft.semester_id = s.id
		ORDER BY ft.created_at DESC
	`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var feeTypes []map[string]interface{}
	for rows.Next() {
		var id, programmeID, semesterID, createdBy uint
		var name string
		var amount int64
		var effectiveFrom, createdAt time.Time
		var programmeName, semesterName string
		if err := rows.Scan(&id, &name, &programmeID, &semesterID, &amount, &effectiveFrom,
			&createdBy, &createdAt, &programmeName, &semesterName); err != nil {
			return nil, err
		}
		feeTypes = append(feeTypes, map[string]interface{}{
			"id":             id,
			"name":           name,
			"programme_id":   programmeID,
			"semester_id":    semesterID,
			"amount":         amount,
			"effective_from": effectiveFrom.Format("2006-01-02"),
			"created_by":     createdBy,
			"created_at":     createdAt,
			"programme_name": programmeName,
			"semester_name":  semesterName,
		})
	}

	return feeTypes, nil
}

func (r *FeeTypeRepository) GetByID(id uint) (map[string]interface{}, error) {
	query := `
		SELECT ft.id, ft.name, ft.programme_id, ft.semester_id, ft.amount, ft.effective_from,
		       ft.created_by, ft.created_at, p.name as programme_name, s.name as semester_name
		FROM fee_types ft
		JOIN programmes p ON ft.programme_id = p.id
		JOIN semesters s ON ft.semester_id = s.id
		WHERE ft.id = ?
	`
	var feeTypeID uint
	var name string
	var programmeID, semesterID, createdBy uint
	var amount int64
	var effectiveFrom, createdAt time.Time
	var programmeName, semesterName string

	err := r.db.QueryRow(query, id).Scan(&feeTypeID, &name, &programmeID, &semesterID, &amount, &effectiveFrom,
		&createdBy, &createdAt, &programmeName, &semesterName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("fee type not found")
		}
		return nil, err
	}

	return map[string]interface{}{
		"id":             feeTypeID,
		"name":           name,
		"programme_id":   programmeID,
		"semester_id":    semesterID,
		"amount":         amount,
		"effective_from": effectiveFrom.Format("2006-01-02"),
		"created_by":     createdBy,
		"created_at":     createdAt,
		"programme_name": programmeName,
		"semester_name":  semesterName,
	}, nil
}

func (r *FeeTypeRepository) Create(name string, programmeID, semesterID, createdBy uint, amount int64, effectiveFrom string) (uint, error) {
	query := `
		INSERT INTO fee_types (name, programme_id, semester_id, amount, effective_from, created_by)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	result, err := r.db.Exec(query, name, programmeID, semesterID, amount, effectiveFrom, createdBy)
	if err != nil {
		return 0, err
	}
	id, _ := result.LastInsertId()
	return uint(id), nil
}

func (r *FeeTypeRepository) Update(id uint, name string, amount int64, effectiveFrom string) (bool, error) {
	query := `
		UPDATE fee_types
		SET name = ?, amount = ?, effective_from = ?
		WHERE id = ?
	`
	result, err := r.db.Exec(query, name, amount, effectiveFrom, id)
	if err != nil {
		return false, err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return rows > 0, nil
}
