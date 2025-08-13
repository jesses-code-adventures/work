-- Create clients table
CREATE TABLE clients (
    id TEXT PRIMARY KEY NOT NULL, -- UUID v7
    name VARCHAR(255) UNIQUE NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP NOT NULL
);

-- Create sessions table
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

-- Create indexes for performance
CREATE INDEX idx_sessions_client_id ON sessions(client_id);
CREATE INDEX idx_sessions_start_time ON sessions(start_time);
CREATE INDEX idx_sessions_end_time ON sessions(end_time);
CREATE INDEX idx_clients_name ON clients(name);

-- Create trigger to update updated_at on clients
CREATE TRIGGER clients_updated_at 
    AFTER UPDATE ON clients 
    BEGIN
        UPDATE clients SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
    END;

-- Create trigger to update updated_at on sessions
CREATE TRIGGER sessions_updated_at 
    AFTER UPDATE ON sessions 
    BEGIN
        UPDATE sessions SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
    END;

---- create above / drop below ----

-- Drop triggers
DROP TRIGGER IF EXISTS sessions_updated_at;
DROP TRIGGER IF EXISTS clients_updated_at;

-- Drop indexes
DROP INDEX IF EXISTS idx_clients_name;
DROP INDEX IF EXISTS idx_sessions_end_time;
DROP INDEX IF EXISTS idx_sessions_start_time;
DROP INDEX IF EXISTS idx_sessions_client_id;

-- Drop tables
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS clients;