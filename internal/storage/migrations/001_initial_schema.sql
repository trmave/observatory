-- 001_initial_schema.sql

-- Catálogo de proveedores
CREATE TABLE IF NOT EXISTS providers (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    slug TEXT NOT NULL UNIQUE,
    provider_type TEXT NOT NULL,
    url TEXT,
    enabled BOOLEAN DEFAULT 1,
    current_status TEXT DEFAULT 'unknown',
    last_checked_at DATETIME,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Histórico de chequeos de estado
CREATE TABLE IF NOT EXISTS status_checks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    provider_id INTEGER NOT NULL,
    status TEXT NOT NULL,
    latency_ms INTEGER,
    raw_response TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (provider_id) REFERENCES providers(id)
);

CREATE INDEX IF NOT EXISTS idx_status_checks_provider_time ON status_checks(provider_id, created_at);

-- Seguimiento de incidentes
CREATE TABLE IF NOT EXISTS incidents (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    provider_id INTEGER NOT NULL,
    title TEXT NOT NULL,
    description TEXT,
    severity TEXT NOT NULL,
    status TEXT DEFAULT 'open', -- open, resolved
    started_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    resolved_at DATETIME,
    FOREIGN KEY (provider_id) REFERENCES providers(id)
);

CREATE INDEX IF NOT EXISTS idx_incidents_status ON incidents(status);

-- Logs de errores técnicos (infraestructura/parsing)
CREATE TABLE IF NOT EXISTS provider_errors (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    provider_id INTEGER NOT NULL,
    error_message TEXT NOT NULL,
    stack_trace TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (provider_id) REFERENCES providers(id)
);
