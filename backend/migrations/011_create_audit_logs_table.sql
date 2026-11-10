CREATE TABLE audit_logs (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  actor_type ENUM('user','system','gateway') NOT NULL,
  actor_id BIGINT UNSIGNED NULL,
  action VARCHAR(60) NOT NULL,
  record_type ENUM('invoice','payment','receipt','fee_type','user') NOT NULL,
  record_ref VARCHAR(40) NOT NULL,
  old_value VARCHAR(255) NULL,
  new_value VARCHAR(255) NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_audit_time (created_at),
  INDEX idx_audit_record (record_type, record_ref)
) ENGINE=InnoDB;
