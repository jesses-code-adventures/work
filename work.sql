CREATE TABLE clients (
    id TEXT PRIMARY KEY NOT NULL, -- UUID v7
    name VARCHAR(255) UNIQUE NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP NOT NULL
, hourly_rate DECIMAL(10,2) DEFAULT 0.00, company_name VARCHAR(255), contact_name VARCHAR(255), email VARCHAR(255), phone VARCHAR(50), address_line1 VARCHAR(255), address_line2 VARCHAR(255), city VARCHAR(100), state VARCHAR(100), postal_code VARCHAR(20), country VARCHAR(100), dir VARCHAR(255), abn VARCHAR(20), retainer_amount DECIMAL(10,2), retainer_hours DECIMAL(10,2), retainer_basis TEXT CHECK (
    retainer_basis IS NULL OR 
    retainer_basis IN ('day', 'week', 'month', 'quarter', 'year')
));
CREATE TABLE sessions (
    id TEXT PRIMARY KEY NOT NULL, -- UUID v7
    client_id TEXT NOT NULL,
    start_time DATETIME NOT NULL,
    end_time DATETIME,
    description TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP NOT NULL, hourly_rate DECIMAL(10,2), full_work_summary TEXT, outside_git TEXT, invoice_id text,
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
CREATE TABLE IF NOT EXISTS "invoices_backup_before_datetime_migration" (
    id text primary key not null, -- uuid v7
    client_id text not null,
    invoice_number varchar(50) unique not null,
    period_type varchar(20) not null, -- 'day', 'week', 'fortnight', 'month'
    period_start_date datetime not null,
    period_end_date datetime not null,
    subtotal_amount decimal(10,2) not null default 0.00,
    gst_amount decimal(10,2) not null default 0.00,
    total_amount decimal(10,2) not null default 0.00,
    generated_date datetime default current_timestamp not null,
    created_at datetime default current_timestamp not null,
    updated_at datetime default current_timestamp not null,
    foreign key (client_id) references clients(id)
);
CREATE INDEX idx_invoices_client_id on "invoices_backup_before_datetime_migration"(client_id);
CREATE INDEX idx_invoices_invoice_number on "invoices_backup_before_datetime_migration"(invoice_number);
CREATE INDEX idx_invoices_period_dates on "invoices_backup_before_datetime_migration"(period_start_date, period_end_date);
CREATE INDEX idx_sessions_invoice_id on sessions(invoice_id);
CREATE TRIGGER invoices_updated_at 
    after update on "invoices_backup_before_datetime_migration" 
    begin
        update invoices set updated_at = current_timestamp where id = new.id;
    end;
CREATE TABLE IF NOT EXISTS "payments_backup_before_datetime_migration" (
	id text primary key not null, -- uuid v7
	invoice_id text not null,
	amount decimal(10,2) not null,
	payment_date datetime not null,
	created_at datetime default current_timestamp not null,
	updated_at datetime default current_timestamp not null,
	foreign key (invoice_id) references invoices(id)
);
CREATE TABLE IF NOT EXISTS "invoices" (
    id text primary key not null, -- uuid v7
    client_id text not null,
    invoice_number varchar(50) unique not null,
    period_type varchar(20) not null, -- 'day', 'week', 'fortnight', 'month'
    period_start_date datetime not null,
    period_end_date datetime not null,
    subtotal_amount decimal(10,2) not null default 0.00,
    gst_amount decimal(10,2) not null default 0.00,
    total_amount decimal(10,2) not null default 0.00,
    generated_date datetime default current_timestamp not null,
    created_at datetime default current_timestamp not null,
    updated_at datetime default current_timestamp not null,
    foreign key (client_id) references clients(id)
);
CREATE TABLE IF NOT EXISTS "payments" (
	id text primary key not null, -- uuid v7
	invoice_id text not null,
	amount decimal(10,2) not null,
	payment_date datetime not null,
	created_at datetime default current_timestamp not null,
	updated_at datetime default current_timestamp not null,
	foreign key (invoice_id) references invoices(id)
);
CREATE VIEW v_invoices AS
SELECT 
	i.*,
	CAST(COALESCE(SUM(p.amount), 0.0) AS REAL) as amount_paid,
	MAX(p.payment_date) as payment_date
FROM invoices i
LEFT JOIN payments p ON p.invoice_id = i.id
GROUP BY i.id
/* v_invoices(id,client_id,invoice_number,period_type,period_start_date,period_end_date,subtotal_amount,gst_amount,total_amount,generated_date,created_at,updated_at,amount_paid,payment_date) */;
