package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"observatory/internal/domain"
)

type GCPChecker struct {
	id     int64
	slug   string
	apiURL string
	client *http.Client
}

type gcpIncident struct {
	ID        string    `json:"id"`
	Service   string    `json:"service_name"`
	Severity  string    `json:"severity"` // high, medium, low
	Status    string    `json:"status_impact"`
	BeginTime time.Time `json:"begin"`
	EndTime   time.Time `json:"end"`
}

func NewGCPChecker(id int64, slug, apiURL string) *GCPChecker {
	return &GCPChecker{
		id:     id,
		slug:   slug,
		apiURL: apiURL,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *GCPChecker) GetSlug() string {
	return c.slug
}

func (c *GCPChecker) Check(ctx context.Context) (domain.CheckResult, error) {
	start := time.Now()
	result := domain.CheckResult{
		ProviderID:   c.id,
		ProviderSlug: c.slug,
		Status:       domain.StatusUnknown,
		Timestamp:    time.Now(),
	}

	resp, err := c.client.Get(c.apiURL)
	if err != nil {
		result.Status = domain.StatusError
		result.ErrorMessage = err.Error()
		result.Latency = time.Since(start)
		return result, nil
	}
	defer resp.Body.Close()

	result.Latency = time.Since(start)

	var incidents []gcpIncident
	if err := json.NewDecoder(resp.Body).Decode(&incidents); err != nil {
		result.Status = domain.StatusError
		result.ErrorMessage = "failed to decode gcp response"
		return result, nil
	}

	// Lógica de normalización: Buscamos incidentes activos (sin end time)
	activeHigh := 0
	activeMedium := 0

	for _, inc := range incidents {
		if inc.EndTime.IsZero() || inc.EndTime.After(time.Now()) {
			switch inc.Severity {
			case "high":
				activeHigh++
			case "medium":
				activeMedium++
			}
		}
	}

	if activeHigh > 0 {
		result.Status = domain.StatusCritical
	} else if activeMedium > 0 {
		result.Status = domain.StatusWarning
	} else {
		result.Status = domain.StatusOK
	}

	raw, _ := json.Marshal(incidents)
	result.RawResponse = string(raw)

	return result, nil
}
