CREATE TABLE invoices (
    id TEXT PRIMARY KEY NOT NULL, -- UUID v7
    client_id TEXT NOT NULL,
    invoice_number VARCHAR(50) UNIQUE NOT NULL,
    period_type VARCHAR(20) NOT NULL, -- 'day', 'week', 'fortnight', 'month'
    period_start_date DATE NOT NULL,
    period_end_date DATE NOT NULL,
    subtotal_amount DECIMAL(10,2) NOT NULL,
    gst_amount DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    total_amount DECIMAL(10,2) NOT NULL,
    amount_paid DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    generated_date DATETIME DEFAULT CURRENT_TIMESTAMP NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP NOT NULL,
    FOREIGN KEY (client_id) REFERENCES clients(id)
);

ALTER TABLE sessions ADD COLUMN invoice_id TEXT;

CREATE INDEX idx_invoices_client_id ON invoices(client_id);
CREATE INDEX idx_invoices_invoice_number ON invoices(invoice_number);
CREATE INDEX idx_invoices_period_dates ON invoices(period_start_date, period_end_date);
CREATE INDEX idx_sessions_invoice_id ON sessions(invoice_id);

CREATE TRIGGER invoices_updated_at 
    AFTER UPDATE ON invoices 
    BEGIN
        UPDATE invoices SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
    END;