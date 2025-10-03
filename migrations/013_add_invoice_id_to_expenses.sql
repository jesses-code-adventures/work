ALTER TABLE expenses ADD COLUMN invoice_id TEXT;

-- Add foreign key constraint to invoices table
-- Note: SQLite doesn't support adding foreign key constraints to existing tables
-- But we can create an index to help with performance
CREATE INDEX idx_expenses_invoice_id ON expenses(invoice_id);

-- Add a check to ensure invoice_id references a valid invoice if provided
-- This will be enforced at the application level since SQLite has limited FK support