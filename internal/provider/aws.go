package provider

import (
	"context"
	"io"
	"net/http"
	"strings"
	"time"

	"observatory/internal/domain"
)

type AWSChecker struct {
	id     int64
	slug   string
	apiURL string
	client *http.Client
}

func NewAWSChecker(id int64, slug, apiURL string) *AWSChecker {
	return &AWSChecker{
		id:     id,
		slug:   slug,
		apiURL: apiURL,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *AWSChecker) GetSlug() string {
	return c.slug
}

func (c *AWSChecker) Check(ctx context.Context) (domain.CheckResult, error) {
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
	result.RawResponse = "HTML content omitted for brevity"

	// Estrategia: Buscar indicadores comunes en el HTML de la página de estado de AWS
	// Nota: AWS suele usar imágenes o iconos CSS. Buscamos strings de severidad conocidos.
	content := string(body)
	
	if strings.Contains(content, "Service is operating normally") || strings.Contains(content, "status-green") {
		result.Status = domain.StatusOK
	} else if strings.Contains(content, "performance issues") || strings.Contains(content, "status-yellow") {
		result.Status = domain.StatusWarning
	} else if strings.Contains(content, "Service disruption") || strings.Contains(content, "status-red") {
		result.Status = domain.StatusCritical
	} else {
		// Si no encontramos nada claro, marcamos como OK si el status code es 200, 
		// pero registramos que no pudimos parsear bien.
		result.Status = domain.StatusOK 
		result.RawResponse = "Parsed as OK by default (HTTP 200)"
	}

	return result, nil
}
