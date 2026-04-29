package api

import (
	"encoding/json"
	"net/http"
	"time"
	"observatory/internal/domain"
	"observatory/internal/service"
)

type Handler struct {
	repo  domain.ProviderRepository
	cache *service.StatusCache
}

func NewHandler(repo domain.ProviderRepository, cache *service.StatusCache) *Handler {
	return &Handler{repo: repo, cache: cache}
}

func (h *Handler) GetStatus(w http.ResponseWriter, r *http.Request) {
	if providers, found := h.cache.Get(); found {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(providers)
		return
	}

	providers, err := h.repo.ListAll(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.cache.Set(providers)
	type apiResponse struct {
		Name          string    `json:"name"`
		Slug          string    `json:"slug"`
		Status        string    `json:"status"`
		LastCheckedAt time.Time `json:"last_checked_at"`
		URL           string    `json:"url"`
	}

	var response []apiResponse
	for _, p := range providers {
		response = append(response, apiResponse{
			Name:          p.Name,
			Slug:          p.Slug,
			Status:        string(p.CurrentStatus),
			LastCheckedAt: p.LastCheckedAt,
			URL:           p.URL,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status": "ok"}`))
}
