package repositories

import (
	"database/sql"
	"time"
)

type AuditRepository struct {
	db *sql.DB
}

func NewAuditRepository(db *sql.DB) *AuditRepository {
	return &AuditRepository{db: db}
}

func (r *AuditRepository) Log(actorType string, actorID *uint, action, recordType, recordRef, oldValue, newValue string) error {
	query := `
		INSERT INTO audit_logs (actor_type, actor_id, action, record_type, record_ref, old_value, new_value)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	var actorIDVal interface{}
	if actorID != nil {
		actorIDVal = *actorID
	} else {
		actorIDVal = nil
	}
	_, err := r.db.Exec(query, actorType, actorIDVal, action, recordType, recordRef, oldValue, newValue)
	return err
}

func (r *AuditRepository) LogInTx(tx *sql.Tx, actorType string, actorID *uint, action, recordType, recordRef, oldValue, newValue string) error {
	query := `
		INSERT INTO audit_logs (actor_type, actor_id, action, record_type, record_ref, old_value, new_value)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	var actorIDVal interface{}
	if actorID != nil {
		actorIDVal = *actorID
	} else {
		actorIDVal = nil
	}
	_, err := tx.Exec(query, actorType, actorIDVal, action, recordType, recordRef, oldValue, newValue)
	return err
}

func (r *AuditRepository) List(recordType, recordRef string, page, pageSize int) ([]map[string]interface{}, int, error) {
	offset := (page - 1) * pageSize

	query := `
		SELECT id, actor_type, actor_id, action, record_type, record_ref, old_value, new_value, created_at
		FROM audit_logs
		WHERE 1=1
	`
	var args []interface{}

	if recordType != "" {
		query += " AND record_type = ?"
		args = append(args, recordType)
	}

	if recordRef != "" {
		query += " AND record_ref = ?"
		args = append(args, recordRef)
	}

	countQuery := "SELECT COUNT(*) FROM audit_logs WHERE 1=1"
	var total int
	countArgs := args
	if recordType != "" {
		countQuery += " AND record_type = ?"
	}
	if recordRef != "" {
		countQuery += " AND record_ref = ?"
	}
	if err := r.db.QueryRow(countQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query += " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	args = append(args, pageSize, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var logs []map[string]interface{}
	for rows.Next() {
		var id uint
		var actorType, action, recordType, recordRef string
		var actorID sql.NullInt64
		var oldValue, newValue sql.NullString
		var createdAt time.Time

		if err := rows.Scan(&id, &actorType, &actorID, &action, &recordType, &recordRef,
			&oldValue, &newValue, &createdAt); err != nil {
			return nil, 0, err
		}

		var actorIDVal *uint
		if actorID.Valid {
			aid := uint(actorID.Int64)
			actorIDVal = &aid
		}

		var oldValueVal, newValueVal *string
		if oldValue.Valid {
			oldValueVal = &oldValue.String
		}
		if newValue.Valid {
			newValueVal = &newValue.String
		}

		logs = append(logs, map[string]interface{}{
			"id":          id,
			"actor_type":  actorType,
			"actor_id":    actorIDVal,
			"action":      action,
			"record_type": recordType,
			"record_ref":  recordRef,
			"old_value":   oldValueVal,
			"new_value":   newValueVal,
			"created_at":  createdAt,
		})
	}

	return logs, total, nil
}
