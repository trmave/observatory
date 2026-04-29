package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"observatory/internal/domain"
)

type StatusPageChecker struct {
	id     int64
	slug   string
	apiURL string
	client *http.Client
}

type statusPageSummary struct {
	Status struct {
		Indicator   string `json:"indicator"` // none, minor, major, critical
		Description string `json:"description"`
	} `json:"status"`
	Incidents []struct {
		Name      string    `json:"name"`
		Status    string    `json:"status"`
		UpdatedAt time.Time `json:"updated_at"`
	} `json:"incidents"`
	Components []struct {
		Name   string `json:"name"`
		Status string `json:"status"` // operational, partial_outage, major_outage
	} `json:"components"`
}

func NewStatusPageChecker(id int64, slug, apiURL string) *StatusPageChecker {
	return &StatusPageChecker{
		id:     id,
		slug:   slug,
		apiURL: apiURL,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *StatusPageChecker) GetSlug() string {
	return c.slug
}

func (c *StatusPageChecker) Check(ctx context.Context) (domain.CheckResult, error) {
	start := time.Now()
	result := domain.CheckResult{
		ProviderID:   c.id,
		ProviderSlug: c.slug,
		Status:       domain.StatusUnknown,
		Timestamp:    time.Now(),
	}

	req, err := http.NewRequestWithContext(ctx, "GET", c.apiURL, nil)
	if err != nil {
		return result, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	resp, err := c.client.Do(req)
	if err != nil {
		result.Status = domain.StatusError
		result.ErrorMessage = err.Error()
		result.Latency = time.Since(start)
		return result, nil
	}
	defer resp.Body.Close()

	result.Latency = time.Since(start)

	if resp.StatusCode != http.StatusOK {
		result.Status = domain.StatusError
		result.ErrorMessage = fmt.Sprintf("unexpected status code: %d", resp.StatusCode)
		return result, nil
	}

	var summary statusPageSummary
	if err := json.NewDecoder(resp.Body).Decode(&summary); err != nil {
		result.Status = domain.StatusError
		result.ErrorMessage = "failed to decode statuspage response"
		return result, nil
	}

	// Mapeo estándar de StatusPage.io
	switch summary.Status.Indicator {
	case "none":
		result.Status = domain.StatusOK
	case "minor", "degraded_performance", "partial":
		result.Status = domain.StatusWarning
	case "major", "critical":
		result.Status = domain.StatusCritical
	default:
		result.Status = domain.StatusUnknown
	}

	result.Message = summary.Status.Description
	// Si hay incidentes activos, usamos el nombre del incidente para dar más detalle (ej. CDN Outage)
	if len(summary.Incidents) > 0 {
		for _, inc := range summary.Incidents {
			if inc.Status != "resolved" && inc.Status != "postmortem" && inc.Status != "completed" {
				result.Message = inc.Name
				break
			}
		}
	}

	// Si hay componentes con fallos graves, los mencionamos de forma compacta
	var criticalComponents []string
	for _, comp := range summary.Components {
		if comp.Status == "major_outage" || comp.Status == "critical" {
			criticalComponents = append(criticalComponents, comp.Name)
			if len(criticalComponents) >= 2 { break } // Max 2 para no saturar
		}
	}
	if len(criticalComponents) > 0 {
		result.Message = fmt.Sprintf("%s [AFFECTING: %s]", result.Message, strings.Join(criticalComponents, ", "))
	}

	raw, _ := json.Marshal(summary)
	result.RawResponse = string(raw)

	return result, nil
}
