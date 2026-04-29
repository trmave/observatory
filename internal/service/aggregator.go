package service

import (
	"context"
	"fmt"
	"log"
	"sync"

	"observatory/internal/domain"
)

type Aggregator struct {
	repo     domain.ProviderRepository
	checkers []domain.Checker
}

func NewAggregator(repo domain.ProviderRepository, checkers []domain.Checker) *Aggregator {
	return &Aggregator{
		repo:     repo,
		checkers: checkers,
	}
}

func (s *Aggregator) RunAllChecks(ctx context.Context) {
	var wg sync.WaitGroup
	
	for _, checker := range s.checkers {
		wg.Add(1)
		go func(c domain.Checker) {
			defer wg.Done()
			
			// Ejecutamos el check
			result, err := c.Check(ctx)
			if err != nil {
				log.Printf("Error checking %s: %v", c.GetSlug(), err)
				return
			}

			// Persistimos el resultado histórico
			if err := s.repo.SaveCheck(ctx, result); err != nil {
				log.Printf("Error saving check for %s: %v", c.GetSlug(), err)
			}

			// Actualizar estado en DB
			msg := result.Message
			if result.Status == domain.StatusError && result.ErrorMessage != "" {
				msg = result.ErrorMessage
			}
			err = s.repo.UpdateStatus(ctx, result.ProviderID, result.Status, result.Timestamp, msg)
			if err != nil {
				log.Printf("Error updating status for %s: %v", result.ProviderSlug, err)
			}

			// Si el estado es crítico o warning, generar un incidente (si no existe uno activo)
			if result.Status == domain.StatusCritical || result.Status == domain.StatusWarning {
				s.repo.SaveIncident(ctx, &domain.Incident{
					ProviderID:  result.ProviderID,
					Title:       fmt.Sprintf("Service Anomaly: %s", result.Status),
					Description: result.RawResponse,
					Status:      "open",
					Severity:    string(result.Status),
					StartedAt:   result.Timestamp,
				})
			}
			
			log.Printf("Check completed for %s (ID: %d): %s (%dms) - Message: %s", c.GetSlug(), result.ProviderID, result.Status, result.Latency.Milliseconds(), result.Message)
		}(checker)
	}

	wg.Wait()
}
