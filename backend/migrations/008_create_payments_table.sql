CREATE TABLE payments (
  id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
  invoice_id BIGINT UNSIGNED NOT NULL,
  student_id BIGINT UNSIGNED NOT NULL,
  amount BIGINT UNSIGNED NOT NULL,
  status ENUM('Pending','Paid','Failed','Cancelled','Expired') NOT NULL DEFAULT 'Pending',
  mobile_number VARCHAR(20) NOT NULL,
  mopay_reference VARCHAR(40) NOT NULL UNIQUE,
  mopay_transaction_id VARCHAR(100) NULL UNIQUE,
  gateway_response JSON NULL,
  expires_at DATETIME NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  status_updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  paid_at DATETIME NULL,
  pending_invoice_id BIGINT UNSIGNED
    GENERATED ALWAYS AS (IF(status = 'Pending', invoice_id, NULL)) STORED,
  UNIQUE KEY uq_one_pending_per_invoice (pending_invoice_id),
  FOREIGN KEY (invoice_id) REFERENCES invoices(id),
  FOREIGN KEY (student_id) REFERENCES students(id),
  INDEX idx_pay_status (status, created_at)
) ENGINE=InnoDB;
