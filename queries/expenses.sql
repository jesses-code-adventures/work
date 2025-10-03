-- name: CreateExpense :one
INSERT INTO expenses (id, amount, expense_date, reference, client_id, invoice_id, description)
VALUES (sqlc.arg(id), sqlc.arg(amount), sqlc.arg(expense_date), sqlc.narg(reference), sqlc.narg(client_id), sqlc.narg(invoice_id), sqlc.narg(description))
RETURNING *;

-- name: GetExpenseByID :one
SELECT * FROM expenses
WHERE id = sqlc.arg(id);

-- name: ListExpenses :many
SELECT * FROM expenses
ORDER BY expense_date DESC;

-- name: ListExpensesByClient :many
SELECT * FROM expenses
WHERE client_id = sqlc.arg(client_id)
ORDER BY expense_date DESC;

-- name: ListExpensesByDateRange :many
SELECT * FROM expenses
WHERE expense_date >= sqlc.arg(start_date) AND expense_date <= sqlc.arg(end_date)
ORDER BY expense_date DESC;

-- name: ListExpensesByClientAndDateRange :many
SELECT * FROM expenses
WHERE client_id = sqlc.arg(client_id) 
  AND expense_date >= sqlc.arg(start_date) 
  AND expense_date <= sqlc.arg(end_date)
ORDER BY expense_date DESC;

-- name: UpdateExpense :one
UPDATE expenses 
SET 
    amount = sqlc.narg(amount),
    expense_date = sqlc.narg(expense_date),
    reference = sqlc.narg(reference),
    client_id = sqlc.narg(client_id),
    invoice_id = sqlc.narg(invoice_id),
    description = sqlc.narg(description)
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: DeleteExpense :exec
DELETE FROM expenses
WHERE id = sqlc.arg(id);

-- name: GetExpensesByReference :many
SELECT * FROM expenses
WHERE reference = sqlc.arg(reference)
ORDER BY expense_date DESC;

-- name: GetExpensesByInvoiceID :many
SELECT * FROM expenses
WHERE invoice_id = sqlc.arg(invoice_id)
ORDER BY expense_date DESC;

-- name: GetExpensesWithoutInvoiceByClient :many
SELECT * FROM expenses
WHERE client_id = sqlc.arg(client_id) AND invoice_id IS NULL
ORDER BY expense_date DESC;

-- name: GetExpensesWithoutInvoiceByClientAndDateRange :many
SELECT * FROM expenses
WHERE client_id = sqlc.arg(client_id) 
  AND invoice_id IS NULL
  AND expense_date >= sqlc.arg(start_date) 
  AND expense_date <= sqlc.arg(end_date)
ORDER BY expense_date DESC;

-- name: UpdateExpenseInvoiceID :exec
UPDATE expenses 
SET invoice_id = sqlc.narg(invoice_id)
WHERE id = sqlc.arg(id);

-- name: ClearExpenseInvoiceIDs :exec
UPDATE expenses 
SET invoice_id = NULL
WHERE invoice_id = sqlc.arg(invoice_id);