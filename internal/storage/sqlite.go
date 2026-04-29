package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"observatory/internal/domain"
	_ "modernc.org/sqlite"
)

type SQLiteRepository struct {
	db *sql.DB
}

func NewSQLiteRepository(dbPath string) (*SQLiteRepository, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("error opening db: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("error pinging db: %w", err)
	}

	// Optimización para concurrencia en SQLite
	db.SetMaxOpenConns(1)
	_, _ = db.Exec("PRAGMA journal_mode=WAL;")
	_, _ = db.Exec("PRAGMA busy_timeout=5000;")

	// Migración automática para LastMessage (si no existe)
	_, _ = db.Exec("ALTER TABLE providers ADD COLUMN last_message TEXT DEFAULT '';")

	return &SQLiteRepository{db: db}, nil
}

func (r *SQLiteRepository) GetDB() *sql.DB {
	return r.db
}

func (r *SQLiteRepository) GetOrCreateProvider(ctx context.Context, p *domain.Provider) (int64, error) {
	var id int64
	err := r.db.QueryRowContext(ctx, "SELECT id FROM providers WHERE slug = ?", p.Slug).Scan(&id)
	if err == sql.ErrNoRows {
		res, err := r.db.ExecContext(ctx, 
			"INSERT INTO providers (name, slug, provider_type, url, enabled, current_status) VALUES (?, ?, ?, ?, ?, ?)",
			p.Name, p.Slug, p.Type, p.URL, p.Enabled, p.CurrentStatus,
		)
		if err != nil {
			return 0, err
		}
		return res.LastInsertId()
	} else if err == nil {
		// Si existe, aseguramos que la URL esté actualizada
		_, _ = r.db.ExecContext(ctx, "UPDATE providers SET url = ? WHERE id = ?", p.URL, id)
	}
	return id, err
}

func (r *SQLiteRepository) ListAll(ctx context.Context) ([]*domain.Provider, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id, name, slug, provider_type, url, enabled, current_status, last_checked_at, last_message FROM providers")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var providers []*domain.Provider
	for rows.Next() {
		p := &domain.Provider{}
		var lastChecked sql.NullTime
		var lastMessage sql.NullString
		if err := rows.Scan(&p.ID, &p.Name, &p.Slug, &p.Type, &p.URL, &p.Enabled, &p.CurrentStatus, &lastChecked, &lastMessage); err != nil {
			return nil, err
		}
		if lastChecked.Valid {
			p.LastCheckedAt = lastChecked.Time
		}
		if lastMessage.Valid {
			p.LastMessage = lastMessage.String
		}
		providers = append(providers, p)
	}
	return providers, nil
}

func (r *SQLiteRepository) UpdateStatus(ctx context.Context, providerID int64, status domain.Status, lastChecked time.Time, message string) error {
	_, err := r.db.ExecContext(ctx, 
		"UPDATE providers SET current_status = ?, last_checked_at = ?, last_message = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		status, lastChecked, message, providerID,
	)
	return err
}

func (r *SQLiteRepository) SaveCheck(ctx context.Context, result domain.CheckResult) error {
	_, err := r.db.ExecContext(ctx,
		"INSERT INTO status_checks (provider_id, status, latency_ms, raw_response) VALUES (?, ?, ?, ?)",
		result.ProviderID, result.Status, result.Latency.Milliseconds(), result.RawResponse,
	)
	return err
}

func (r *SQLiteRepository) GetBySlug(ctx context.Context, slug string) (*domain.Provider, error) {
	p := &domain.Provider{}
	var lastChecked sql.NullTime
	var lastMessage sql.NullString
	query := "SELECT id, name, slug, provider_type, url, enabled, current_status, last_checked_at, last_message FROM providers WHERE slug = ?"
	err := r.db.QueryRowContext(ctx, query, slug).Scan(&p.ID, &p.Name, &p.Slug, &p.Type, &p.URL, &p.Enabled, &p.CurrentStatus, &lastChecked, &lastMessage)
	
	if err != nil {
		return nil, err
	}
	if lastChecked.Valid {
		p.LastCheckedAt = lastChecked.Time
	}
	if lastMessage.Valid {
		p.LastMessage = lastMessage.String
	}
	return p, nil
}

func (r *SQLiteRepository) ListRecentIncidents(ctx context.Context, limit int) ([]*domain.Incident, error) {
	query := `
		SELECT i.id, i.provider_id, i.title, i.description, i.status, i.severity, i.started_at, i.resolved_at, p.name
		FROM incidents i
		JOIN providers p ON i.provider_id = p.id
		ORDER BY i.started_at DESC LIMIT ?`
	
	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var incidents []*domain.Incident
	for rows.Next() {
		inc := &domain.Incident{}
		var resolvedAt sql.NullTime
		if err := rows.Scan(&inc.ID, &inc.ProviderID, &inc.Title, &inc.Description, &inc.Status, &inc.Severity, &inc.StartedAt, &resolvedAt, &inc.ProviderName); err != nil {
			return nil, err
		}
		if resolvedAt.Valid {
			inc.ResolvedAt = &resolvedAt.Time
		}
		incidents = append(incidents, inc)
	}
	return incidents, nil
}

func (r *SQLiteRepository) SaveIncident(ctx context.Context, inc *domain.Incident) error {
	query := "INSERT INTO incidents (provider_id, title, description, status, severity, started_at) VALUES (?, ?, ?, ?, ?, ?)"
	_, err := r.db.ExecContext(ctx, query, inc.ProviderID, inc.Title, inc.Description, inc.Status, inc.Severity, inc.StartedAt)
	return err
}
