package provider

import (
	"context"
	"io"
	"net/http"
	"strings"
	"time"

	"observatory/internal/domain"
)

type AzureChecker struct {
	id     int64
	slug   string
	apiURL string
	client *http.Client
}

func NewAzureChecker(id int64, slug, apiURL string) *AzureChecker {
	return &AzureChecker{
		id:     id,
		slug:   slug,
		apiURL: apiURL,
		client: &http.Client{Timeout: 15 * time.Second},
	}
}

func (c *AzureChecker) GetSlug() string {
	return c.slug
}

func (c *AzureChecker) Check(ctx context.Context) (domain.CheckResult, error) {
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

	body, _ := io.ReadAll(resp.Body)
	result.Latency = time.Since(start)
	content := strings.ToLower(string(body))

	// Azure usa terminología específica en su tabla de estados
	if strings.Contains(content, "good") || strings.Contains(content, "operating normally") {
		result.Status = domain.StatusOK
	} else if strings.Contains(content, "warning") || strings.Contains(content, "advisory") {
		result.Status = domain.StatusWarning
	} else if strings.Contains(content, "critical") || strings.Contains(content, "outage") {
		result.Status = domain.StatusCritical
	} else {
		result.Status = domain.StatusOK // Asumimos salud si el sitio responde
	}

	return result, nil
}
