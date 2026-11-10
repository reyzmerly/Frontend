-- Seed users for testing
-- Passwords are hashed with bcrypt (cost 10)
-- All passwords: "password123"

INSERT INTO users (email, full_name, password_hash, role, access_status) VALUES
('student@uk.ac.rw', 'Student User', '$2a$10$YmOCrsquccdZi2ZipzvTre.eR5kilwagGADRXsCvszEp9RIcvtI9.', 'student', 'active'),
('finance@uk.ac.rw', 'Finance Officer', '$2a$10$YmOCrsquccdZi2ZipzvTre.eR5kilwagGADRXsCvszEp9RIcvtI9.', 'finance', 'active'),
('admin@uk.ac.rw', 'System Admin', '$2a$10$YmOCrsquccdZi2ZipzvTre.eR5kilwagGADRXsCvszEp9RIcvtI9.', 'admin', 'active');
