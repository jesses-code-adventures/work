-- name: CreateSession :one
INSERT INTO sessions (id, client_id, start_time, description, hourly_rate)
VALUES (sqlc.arg(id), sqlc.arg(client_id), sqlc.arg(start_time), sqlc.narg(description), sqlc.narg(hourly_rate))
RETURNING *;

-- name: GetActiveSession :one
SELECT s.*, c.name as client_name
FROM sessions s
JOIN clients c ON s.client_id = c.id
WHERE s.end_time IS NULL
ORDER BY s.start_time DESC
LIMIT 1;

-- name: StopSession :one
UPDATE sessions
SET end_time = sqlc.arg(end_time)
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: ListRecentSessions :many
SELECT s.*, c.name as client_name
FROM sessions s
JOIN clients c ON s.client_id = c.id
ORDER BY s.start_time DESC
LIMIT sqlc.arg(limit_count);

-- name: GetSessionsByClient :many
SELECT s.*, c.name as client_name
FROM sessions s
JOIN clients c ON s.client_id = c.id
WHERE c.name = sqlc.arg(client_name)
ORDER BY s.start_time DESC;

-- name: GetSessionsByDateRange :many
SELECT s.*, c.name as client_name
FROM sessions s
JOIN clients c ON s.client_id = c.id
WHERE s.start_time >= sqlc.arg(start_date) AND s.start_time <= sqlc.arg(end_date)
ORDER BY s.start_time DESC;

-- name: ListSessionsWithDateRange :many
SELECT s.*, c.name as client_name
FROM sessions s
JOIN clients c ON s.client_id = c.id
WHERE (sqlc.narg(start_date) IS NULL OR s.start_time >= sqlc.narg(start_date)) 
  AND (sqlc.narg(end_date) IS NULL OR s.start_time <= sqlc.narg(end_date))
ORDER BY s.start_time DESC
LIMIT sqlc.arg(limit_count);

-- name: DeleteAllSessions :exec
DELETE FROM sessions;

-- name: DeleteSessionsByDateRange :exec
DELETE FROM sessions
WHERE (sqlc.narg(start_date) IS NULL OR start_time >= sqlc.narg(start_date)) 
  AND (sqlc.narg(end_date) IS NULL OR start_time <= sqlc.narg(end_date));