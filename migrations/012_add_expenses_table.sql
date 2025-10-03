CREATE TABLE expenses (
    id TEXT PRIMARY KEY NOT NULL, -- UUID v7
    amount DECIMAL(10,2) NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP NOT NULL,
    expense_date DATETIME NOT NULL,
    reference TEXT,
    client_id TEXT,
    FOREIGN KEY (client_id) REFERENCES clients(id)
);

CREATE INDEX idx_expenses_client_id ON expenses(client_id);
CREATE INDEX idx_expenses_expense_date ON expenses(expense_date);
CREATE INDEX idx_expenses_created_at ON expenses(created_at);

CREATE TRIGGER expenses_updated_at 
    AFTER UPDATE ON expenses 
    BEGIN
        UPDATE expenses SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
    END;