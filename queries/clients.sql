-- name: CreateClient :one
INSERT INTO clients (id, name, hourly_rate)
VALUES (sqlc.arg(id), sqlc.arg(name), sqlc.narg(hourly_rate))
RETURNING *;

-- name: GetClientByName :one
SELECT * FROM clients
WHERE name = sqlc.arg(name);

-- name: GetClientById :one
SELECT * FROM clients
WHERE id = sqlc.arg(id);

-- name: ListClients :many
SELECT * FROM clients
ORDER BY name;

-- name: UpdateClientRate :one
UPDATE clients 
SET hourly_rate = sqlc.arg(hourly_rate)
WHERE id = sqlc.arg(id)
RETURNING *;