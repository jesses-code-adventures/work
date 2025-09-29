CREATE TABLE clients (
    id TEXT PRIMARY KEY NOT NULL, -- UUID v7
    name VARCHAR(255) UNIQUE NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP NOT NULL
, hourly_rate DECIMAL(10,2) DEFAULT 0.00, company_name VARCHAR(255), contact_name VARCHAR(255), email VARCHAR(255), phone VARCHAR(50), address_line1 VARCHAR(255), address_line2 VARCHAR(255), city VARCHAR(100), state VARCHAR(100), postal_code VARCHAR(20), country VARCHAR(100), dir VARCHAR(255), abn VARCHAR(20));
CREATE TABLE sessions (
    id TEXT PRIMARY KEY NOT NULL, -- UUID v7
    client_id TEXT NOT NULL,
    start_time DATETIME NOT NULL,
    end_time DATETIME,
    description TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP NOT NULL, hourly_rate DECIMAL(10,2), full_work_summary TEXT, outside_git TEXT, invoice_id TEXT,
    FOREIGN KEY (client_id) REFERENCES clients(id)
);
CREATE INDEX idx_sessions_client_id ON sessions(client_id);
CREATE INDEX idx_sessions_start_time ON sessions(start_time);
CREATE INDEX idx_sessions_end_time ON sessions(end_time);
CREATE INDEX idx_clients_name ON clients(name);
CREATE TRIGGER clients_updated_at 
    AFTER UPDATE ON clients 
    BEGIN
        UPDATE clients SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
    END;
CREATE TRIGGER sessions_updated_at 
    AFTER UPDATE ON sessions 
    BEGIN
        UPDATE sessions SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
    END;
CREATE TABLE invoices (
    id TEXT PRIMARY KEY NOT NULL, -- UUID v7
    client_id TEXT NOT NULL,
    invoice_number VARCHAR(50) UNIQUE NOT NULL,
    period_type VARCHAR(20) NOT NULL, -- 'day', 'week', 'fortnight', 'month'
    period_start_date DATE NOT NULL,
    period_end_date DATE NOT NULL,
    subtotal_amount DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    gst_amount DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    total_amount DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    amount_paid DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    generated_date DATETIME DEFAULT CURRENT_TIMESTAMP NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP NOT NULL,
    FOREIGN KEY (client_id) REFERENCES clients(id)
);
CREATE INDEX idx_invoices_client_id ON invoices(client_id);
CREATE INDEX idx_invoices_invoice_number ON invoices(invoice_number);
CREATE INDEX idx_invoices_period_dates ON invoices(period_start_date, period_end_date);
CREATE INDEX idx_sessions_invoice_id ON sessions(invoice_id);
CREATE TRIGGER invoices_updated_at 
    AFTER UPDATE ON invoices 
    BEGIN
        UPDATE invoices SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
    END;
