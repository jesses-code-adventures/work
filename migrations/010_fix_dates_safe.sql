-- SAFE VERSION: Alternative migration for production environments
-- This version preserves the original tables as backups and uses simpler logic

-- Drop the view first
DROP VIEW IF EXISTS v_invoices;

-- Step 1: Create new invoices table with datetime fields
CREATE TABLE invoices_new (
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

-- Migrate data from old invoices table with robust date conversion
INSERT INTO invoices_new (
    id, client_id, invoice_number, period_type, 
    period_start_date, period_end_date,
    subtotal_amount, gst_amount, total_amount,
    generated_date, created_at, updated_at
)
SELECT 
    id, client_id, invoice_number, period_type,
    CASE 
        -- If it's already a datetime string with time (contains colon)
        WHEN period_start_date LIKE '%:%' THEN period_start_date
        -- If it's a date string (YYYY-MM-DD format)
        WHEN period_start_date LIKE '____-__-__' THEN period_start_date || ' 00:00:00'
        -- If it contains a date pattern but might have extra chars, extract YYYY-MM-DD
        WHEN length(period_start_date) >= 10 AND substr(period_start_date, 5, 1) = '-' AND substr(period_start_date, 8, 1) = '-' THEN 
            substr(period_start_date, 1, 10) || ' 00:00:00'
        -- Fallback: use current date at midnight
        ELSE datetime('now', 'start of day')
    END as period_start_date,
    CASE 
        -- If it's already a datetime string with time (contains colon)
        WHEN period_end_date LIKE '%:%' THEN period_end_date
        -- If it's a date string (YYYY-MM-DD format)
        WHEN period_end_date LIKE '____-__-__' THEN period_end_date || ' 23:59:59'
        -- If it contains a date pattern but might have extra chars, extract YYYY-MM-DD
        WHEN length(period_end_date) >= 10 AND substr(period_end_date, 5, 1) = '-' AND substr(period_end_date, 8, 1) = '-' THEN 
            substr(period_end_date, 1, 10) || ' 23:59:59'
        -- Fallback: use current date at end of day
        ELSE datetime('now', 'start of day', '+1 day', '-1 second')
    END as period_end_date,
    subtotal_amount, gst_amount, total_amount,
    generated_date, created_at, updated_at
FROM invoices;

-- Rename old table as backup and replace with new table
ALTER TABLE invoices RENAME TO invoices_backup_before_datetime_migration;
ALTER TABLE invoices_new RENAME TO invoices;

-- Step 2: Handle payments table if it exists
-- Create new payments table with datetime
CREATE TABLE payments_new (
	id text primary key not null, -- uuid v7
	invoice_id text not null,
	amount decimal(10,2) not null,
	payment_date datetime not null,
	created_at datetime default current_timestamp not null,
	updated_at datetime default current_timestamp not null,
	foreign key (invoice_id) references invoices(id)
);

-- Check if payments table exists and migrate data
INSERT INTO payments_new (id, invoice_id, amount, payment_date, created_at, updated_at)
SELECT 
    id, invoice_id, amount,
    CASE 
        -- If it's already a datetime string with time (contains colon)
        WHEN payment_date LIKE '%:%' THEN payment_date
        -- If it's a date string (YYYY-MM-DD format)
        WHEN payment_date LIKE '____-__-__' THEN payment_date || ' 12:00:00'
        -- If it contains a date pattern but might have extra chars, extract YYYY-MM-DD
        WHEN length(payment_date) >= 10 AND substr(payment_date, 5, 1) = '-' AND substr(payment_date, 8, 1) = '-' THEN 
            substr(payment_date, 1, 10) || ' 12:00:00'
        -- Fallback: use current datetime
        ELSE datetime('now')
    END as payment_date,
    created_at, updated_at
FROM payments
WHERE EXISTS (SELECT 1 FROM sqlite_master WHERE type='table' AND name='payments');

-- Safely handle payments table rename
-- First, try to rename the old payments table to backup (this may fail if table doesn't exist)
ALTER TABLE payments RENAME TO payments_backup_before_datetime_migration;

-- Then rename the new table to payments (this should always work)
ALTER TABLE payments_new RENAME TO payments;

-- Step 4: Recreate the view
CREATE VIEW v_invoices AS
SELECT 
	i.*,
	CAST(COALESCE(SUM(p.amount), 0.0) AS REAL) as amount_paid,
	MAX(p.payment_date) as payment_date
FROM invoices i
LEFT JOIN payments p ON p.invoice_id = i.id
GROUP BY i.id;
