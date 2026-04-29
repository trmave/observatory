package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"observatory/internal/domain"
)

type GitHubChecker struct {
	id     int64
	slug   string
	apiURL string
	client *http.Client
}

type gitHubSummary struct {
	Status struct {
		Indicator   string `json:"indicator"` // none, minor, major, critical
		Description string `json:"description"`
	} `json:"status"`
}

func NewGitHubChecker(id int64, slug, apiURL string) *GitHubChecker {
	return &GitHubChecker{
		id:     id,
		slug:   slug,
		apiURL: apiURL,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *GitHubChecker) GetSlug() string {
	return c.slug
}

func (c *GitHubChecker) Check(ctx context.Context) (domain.CheckResult, error) {
	start := time.Now()
	result := domain.CheckResult{
		ProviderID: c.id,
		Status:     domain.StatusUnknown,
	}

	req, err := http.NewRequestWithContext(ctx, "GET", c.apiURL, nil)
	if err != nil {
		return result, err
	}

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

	var summary gitHubSummary
	if err := json.NewDecoder(resp.Body).Decode(&summary); err != nil {
		result.Status = domain.StatusError
		result.ErrorMessage = fmt.Sprintf("failed to decode response: %v", err)
		return result, nil
	}

	// Normalización del estado
	switch summary.Status.Indicator {
	case "none":
		result.Status = domain.StatusOK
	case "minor":
		result.Status = domain.StatusWarning
	case "major", "critical":
		result.Status = domain.StatusCritical
	default:
		result.Status = domain.StatusUnknown
	}

	raw, _ := json.Marshal(summary)
	result.RawResponse = string(raw)

	return result, nil
}
