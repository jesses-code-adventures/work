-- name: CreateInvoice :one
INSERT INTO invoices (id, client_id, invoice_number, period_type, period_start_date, period_end_date, subtotal_amount, gst_amount, total_amount)
VALUES (sqlc.arg(id), sqlc.arg(client_id), sqlc.arg(invoice_number), sqlc.arg(period_type), sqlc.arg(period_start_date), sqlc.arg(period_end_date), sqlc.arg(subtotal_amount), sqlc.arg(gst_amount), sqlc.arg(total_amount))
RETURNING *;

-- name: GetInvoiceByID :one
SELECT i.*, c.name as client_name
FROM v_invoices i
JOIN clients c ON i.client_id = c.id
WHERE i.id = sqlc.arg(id);

-- name: GetInvoiceByNumber :one
SELECT i.*, c.name as client_name
FROM v_invoices i
JOIN clients c ON i.client_id = c.id
WHERE i.invoice_number = sqlc.arg(invoice_number);

-- name: ListInvoices :many
SELECT i.*, c.name as client_name
FROM v_invoices i
JOIN clients c ON i.client_id = c.id
ORDER BY i.generated_date DESC
LIMIT sqlc.arg(limit_count);

-- name: GetInvoicesByClient :many
SELECT i.*, c.name as client_name
FROM v_invoices i
JOIN clients c ON i.client_id = c.id
WHERE c.name = sqlc.arg(client_name)
ORDER BY i.generated_date DESC;

-- name: GetInvoicesByPeriod :many
SELECT i.*, c.name as client_name
FROM v_invoices i
JOIN clients c ON i.client_id = c.id
WHERE i.period_start_date = sqlc.arg(period_start_date) 
  AND i.period_end_date = sqlc.arg(period_end_date)
  AND i.period_type = sqlc.arg(period_type)
ORDER BY c.name;

-- name: DeleteInvoice :exec
DELETE FROM invoices
WHERE id = sqlc.arg(id);

-- name: UpdateSessionInvoiceID :exec
UPDATE sessions
SET invoice_id = sqlc.arg(invoice_id)
WHERE id = sqlc.arg(session_id);

-- name: GetSessionsForPeriodWithoutInvoice :many
SELECT s.*, c.name as client_name
FROM sessions s
JOIN clients c ON s.client_id = c.id
WHERE s.start_time >= sqlc.arg(start_date) 
  AND s.start_time <= sqlc.arg(end_date)
  AND s.end_time IS NOT NULL
  AND s.invoice_id IS NULL
ORDER BY c.name, s.start_time;

-- name: GetSessionsByInvoiceID :many
SELECT s.*, c.name as client_name
FROM sessions s
JOIN clients c ON s.client_id = c.id
WHERE s.invoice_id = sqlc.arg(invoice_id)
ORDER BY s.start_time;

-- name: ClearSessionInvoiceIDs :exec
UPDATE sessions
SET invoice_id = NULL
WHERE invoice_id = sqlc.arg(invoice_id);

-- name: GetSessionsForPeriodWithoutInvoiceByClient :many
SELECT s.*, c.name as client_name
FROM sessions s
JOIN clients c ON s.client_id = c.id
WHERE s.start_time >= sqlc.arg(start_date) 
  AND s.start_time <= sqlc.arg(end_date)
  AND s.end_time IS NOT NULL
  AND s.invoice_id IS NULL
  AND c.name = sqlc.arg(client_name)
ORDER BY s.start_time;

-- name: GetInvoicesByPeriodAndClient :many
SELECT i.*, c.name as client_name
FROM v_invoices i
JOIN clients c ON i.client_id = c.id
WHERE i.period_start_date = sqlc.arg(period_start_date) 
  AND i.period_end_date = sqlc.arg(period_end_date)
  AND i.period_type = sqlc.arg(period_type)
  AND c.name = sqlc.arg(client_name)
ORDER BY i.generated_date;

-- name: PayInvoice :exec
INSERT INTO payments (id, invoice_id, amount, payment_date)
VALUES (sqlc.arg(id), sqlc.arg(invoice_id), sqlc.arg(amount), sqlc.arg(payment_date));
