package domain

import (
	"context"
	"time"
)

// Status representa el estado normalizado de un servicio
type Status string

const (
	StatusOK       Status = "ok"
	StatusWarning  Status = "warning"
	StatusCritical Status = "critical"
	StatusError    Status = "error"
	StatusUnknown  Status = "unknown"
)

// Provider representa un servicio externo monitoreado
type Provider struct {
	ID            int64     `json:"id"`
	Name          string    `json:"name"`
	Slug          string    `json:"slug"`
	Type          string    `json:"type"`
	URL           string    `json:"url"`
	Enabled       bool      `json:"enabled"`
	CurrentStatus Status    `json:"current_status"`
	LastMessage   string    `json:"last_message"`
	LastCheckedAt time.Time `json:"last_checked_at"`
}

// CheckResult es el resultado de una operacion de polling
type CheckResult struct {
	ProviderID   int64         `json:"provider_id"`
	ProviderSlug string        `json:"provider_slug"`
	Status       Status        `json:"status"`
	Message      string        `json:"message"`
	Latency      time.Duration `json:"latency"`
	RawResponse  string        `json:"raw_response"`
	ErrorMessage string        `json:"error_message,omitempty"`
	Timestamp    time.Time     `json:"timestamp"`
}

type Incident struct {
	ID           int64
	ProviderID   int64
	ProviderName string
	Title        string
	Description  string
	Status       string
	Severity     string
	StartedAt    time.Time
	ResolvedAt   *time.Time
}

// Checker es la interfaz que debe implementar cada adaptador de proveedor
type Checker interface {
	Check(ctx context.Context) (CheckResult, error)
	GetSlug() string
}

// ProviderRepository define como persistimos los datos
type ProviderRepository interface {
	GetOrCreateProvider(ctx context.Context, p *Provider) (int64, error)
	ListAll(ctx context.Context) ([]*Provider, error)
	UpdateStatus(ctx context.Context, providerID int64, status Status, lastChecked time.Time, message string) error
	SaveCheck(ctx context.Context, result CheckResult) error
	GetBySlug(ctx context.Context, slug string) (*Provider, error)
	ListRecentIncidents(ctx context.Context, limit int) ([]*Incident, error)
	SaveIncident(ctx context.Context, inc *Incident) error
}
