package web

import (
	"html/template"
	"net/http"
	"strings"
	"time"
	"observatory/internal/domain"
	"observatory/internal/service"
)

type WebHandler struct {
	repo      domain.ProviderRepository
	templates *template.Template
	cache     *service.StatusCache
}

func NewWebHandler(repo domain.ProviderRepository, cache *service.StatusCache) *WebHandler {
	_ = LoadTranslations("configs/i18n.json")
	tmpl := template.Must(template.ParseGlob("internal/web/templates/*.html"))
	return &WebHandler{
		repo:      repo,
		templates: tmpl,
		cache:     cache,
	}
}

func (h *WebHandler) Index(w http.ResponseWriter, r *http.Request) {
	lang := GetLanguage(r.Header.Get("Accept-Language"))
	
	var providers []*domain.Provider
	var err error

	if cached, found := h.cache.Get(); found {
		providers = cached
	} else {
		providers, err = h.repo.ListAll(r.Context())
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	incidents, _ := h.repo.ListRecentIncidents(r.Context(), 5)
	
	activeIncidents := 0
	for _, p := range providers {
		if p.CurrentStatus == domain.StatusCritical || p.CurrentStatus == domain.StatusWarning {
			activeIncidents++
		}
	}

	data := map[string]interface{}{
		"Providers":       providers,
		"Incidents":       incidents,
		"ActiveIncidents": activeIncidents,
		"Now":             time.Now().Format("15:04:05"),
		"T":               translations[lang],
	}

	h.templates.ExecuteTemplate(w, "index.html", data)
}

func (h *WebHandler) Search(w http.ResponseWriter, r *http.Request) {
	lang := GetLanguage(r.Header.Get("Accept-Language"))
	query := strings.ToLower(r.URL.Query().Get("search"))
	
	var allProviders []*domain.Provider
	var err error

	if cached, found := h.cache.Get(); found {
		allProviders = cached
	} else {
		allProviders, err = h.repo.ListAll(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		h.cache.Set(allProviders)
	}

	var filtered []*domain.Provider
	for _, p := range allProviders {
		if query == "" || strings.Contains(strings.ToLower(p.Name), query) || strings.Contains(strings.ToLower(p.Slug), query) {
			filtered = append(filtered, p)
		}
	}

	data := struct {
		Providers []*domain.Provider
		T         map[string]string
	}{
		Providers: filtered,
		T:         translations[lang],
	}

	h.templates.ExecuteTemplate(w, "grid.html", data)
}

func (h *WebHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	// Reutilizamos la lógica de búsqueda sin query para refrescar todo el grid
	h.Search(w, r)
}
