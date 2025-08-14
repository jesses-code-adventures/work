CREATE TABLE clients (
    id TEXT PRIMARY KEY NOT NULL, -- UUID v7
    name VARCHAR(255) UNIQUE NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE TABLE sessions (
    id TEXT PRIMARY KEY NOT NULL, -- UUID v7
    client_id TEXT NOT NULL,
    start_time DATETIME NOT NULL,
    end_time DATETIME,
    description TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP NOT NULL,
    FOREIGN KEY (client_id) REFERENCES clients(id)
);

CREATE INDEX idx_sessions_client_id ON sessions(client_id);
CREATE INDEX idx_sessions_start_time ON sessions(start_time);
CREATE INDEX idx_sessions_end_time ON sessions(end_time);
CREATE INDEX idx_clients_name ON clients(name);

CREATE TRIGGER clients_updated_at 
    AFTER UPDATE ON clients 
    BEGIN
        UPDATE clients SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
    END;

CREATE TRIGGER sessions_updated_at 
    AFTER UPDATE ON sessions 
    BEGIN
        UPDATE sessions SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
    END;
