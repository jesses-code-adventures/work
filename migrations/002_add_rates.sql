-- Add rate column to clients table
ALTER TABLE clients ADD COLUMN hourly_rate DECIMAL(10,2) DEFAULT 0.00;

-- Add rate column to sessions table to record the rate at the time of work
ALTER TABLE sessions ADD COLUMN hourly_rate DECIMAL(10,2);

---- create above / drop below ----

-- Remove rate columns
ALTER TABLE sessions DROP COLUMN hourly_rate;
ALTER TABLE clients DROP COLUMN hourly_rate;