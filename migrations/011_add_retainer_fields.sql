ALTER TABLE clients ADD COLUMN retainer_amount DECIMAL(10,2);
ALTER TABLE clients ADD COLUMN retainer_hours DECIMAL(10,2);
ALTER TABLE clients ADD COLUMN retainer_basis TEXT CHECK (
    retainer_basis IS NULL OR 
    retainer_basis IN ('day', 'week', 'month', 'quarter', 'year')
);
