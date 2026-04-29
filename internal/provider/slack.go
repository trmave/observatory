package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"observatory/internal/domain"
)

type SlackChecker struct {
	id     int64
	slug   string
	apiURL string
	client *http.Client
}

type slackStatusResponse struct {
	Status          string `json:"status"` // ok, active_incident
	DateCreated     string `json:"date_created"`
	DateUpdated     string `json:"date_updated"`
	ActiveIncidents []struct {
		Title    string `json:"title"`
		Type     string `json:"type"` // incident, maintenance
		Status   string `json:"status"`
		URL      string `json:"url"`
	} `json:"active_incidents"`
}

func NewSlackChecker(id int64, slug, apiURL string) *SlackChecker {
	return &SlackChecker{
		id:     id,
		slug:   slug,
		apiURL: apiURL,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *SlackChecker) GetSlug() string {
	return c.slug
}

func (c *SlackChecker) Check(ctx context.Context) (domain.CheckResult, error) {
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
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

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

	var data slackStatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		result.Status = domain.StatusError
		result.ErrorMessage = "failed to decode slack status response"
		return result, nil
	}

	if data.Status == "ok" {
		result.Status = domain.StatusOK
		result.Message = "Slack is up and running"
	} else {
		result.Status = domain.StatusWarning
		if len(data.ActiveIncidents) > 0 {
			result.Message = data.ActiveIncidents[0].Title
			if data.ActiveIncidents[0].Type == "outage" {
				result.Status = domain.StatusCritical
			}
		} else {
			result.Message = "Active incident detected"
		}
	}

	raw, _ := json.Marshal(data)
	result.RawResponse = string(raw)

	return result, nil
}
