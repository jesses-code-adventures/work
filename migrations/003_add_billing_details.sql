-- Add billing details to clients table
ALTER TABLE clients ADD COLUMN company_name VARCHAR(255);
ALTER TABLE clients ADD COLUMN contact_name VARCHAR(255);
ALTER TABLE clients ADD COLUMN email VARCHAR(255);
ALTER TABLE clients ADD COLUMN phone VARCHAR(50);
ALTER TABLE clients ADD COLUMN address_line1 VARCHAR(255);
ALTER TABLE clients ADD COLUMN address_line2 VARCHAR(255);
ALTER TABLE clients ADD COLUMN city VARCHAR(100);
ALTER TABLE clients ADD COLUMN state VARCHAR(100);
ALTER TABLE clients ADD COLUMN postal_code VARCHAR(20);
ALTER TABLE clients ADD COLUMN country VARCHAR(100);
ALTER TABLE clients ADD COLUMN tax_number VARCHAR(50);

---- create above / drop below ----

-- Remove billing details columns
ALTER TABLE clients DROP COLUMN tax_number;
ALTER TABLE clients DROP COLUMN country;
ALTER TABLE clients DROP COLUMN postal_code;
ALTER TABLE clients DROP COLUMN state;
ALTER TABLE clients DROP COLUMN city;
ALTER TABLE clients DROP COLUMN address_line2;
ALTER TABLE clients DROP COLUMN address_line1;
ALTER TABLE clients DROP COLUMN phone;
ALTER TABLE clients DROP COLUMN email;
ALTER TABLE clients DROP COLUMN contact_name;
ALTER TABLE clients DROP COLUMN company_name;