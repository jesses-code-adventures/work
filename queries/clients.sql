-- name: CreateClient :one
INSERT INTO clients (id, name, hourly_rate, company_name, contact_name, email, phone, address_line1, address_line2, city, state, postal_code, country, tax_number, dir)
VALUES (sqlc.arg(id), sqlc.arg(name), sqlc.narg(hourly_rate), sqlc.narg(company_name), sqlc.narg(contact_name), sqlc.narg(email), sqlc.narg(phone), sqlc.narg(address_line1), sqlc.narg(address_line2), sqlc.narg(city), sqlc.narg(state), sqlc.narg(postal_code), sqlc.narg(country), sqlc.narg(tax_number), sqlc.narg(dir))
RETURNING *;

-- name: GetClientByName :one
SELECT * FROM clients
WHERE name = sqlc.arg(name);

-- name: GetClientByID :one
SELECT * FROM clients
WHERE id = sqlc.arg(id);

-- name: ListClients :many
SELECT * FROM clients
ORDER BY name;

-- name: UpdateClient :one
UPDATE clients 
SET 
	hourly_rate = sqlc.narg(hourly_rate),
    company_name = sqlc.narg(company_name),
    contact_name = sqlc.narg(contact_name),
    email = sqlc.narg(email),
    phone = sqlc.narg(phone),
    address_line1 = sqlc.narg(address_line1),
    address_line2 = sqlc.narg(address_line2),
    city = sqlc.narg(city),
    state = sqlc.narg(state),
    postal_code = sqlc.narg(postal_code),
    country = sqlc.narg(country),
    tax_number = sqlc.narg(tax_number),
    dir = sqlc.narg(dir)
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: GetClientsWithDirectories :many
SELECT * FROM clients
WHERE dir IS NOT NULL AND dir != ''
ORDER BY name;
