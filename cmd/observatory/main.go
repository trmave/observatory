package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"observatory/internal/api"
	"observatory/internal/config"
	"observatory/internal/domain"
	"observatory/internal/mcp"
	"observatory/internal/provider"
	"observatory/internal/service"
	"observatory/internal/storage"
	"observatory/internal/web"
)

func main() {
	// 1. Cargar Configuración
	cfg, err := config.Load("configs/config.yaml")
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// 2. Inicializar Base de Datos
	repo, err := storage.NewSQLiteRepository(cfg.Database.Path)
	if err != nil {
		log.Fatalf("Error connecting to db: %v", err)
	}

	// 2.5 Modo MCP (Model Context Protocol)
	if len(os.Args) > 1 && os.Args[1] == "--mcp" {
		mcpServer := mcp.NewMCPServer(repo)
		mcpServer.Serve()
		return
	}

	// 3. Ejecutar Migraciones iniciales (simple para el MVP)
	runMigrations(cfg.Database.Path)

	// 4. Registrar Proveedores y crear Checkers
	var checkers []domain.Checker
	for _, pc := range cfg.Providers {
		if !pc.Enabled {
			continue
		}
		
		pID, err := repo.GetOrCreateProvider(context.Background(), &domain.Provider{
			Name:    pc.Name,
			Slug:    pc.Slug,
			Type:    pc.Type,
			URL:     pc.URL,
			Enabled: pc.Enabled,
			CurrentStatus: domain.StatusUnknown,
		})
		if err != nil {
			log.Printf("Error ensuring provider %s: %v", pc.Slug, err)
			continue
		}

		// Registro de Checkers: Prioridad 1 - Checkers específicos por Slug
		var handled bool
		switch pc.Slug {
		case "aws":
			checkers = append(checkers, provider.NewAWSChecker(pID, pc.Slug, pc.APIURL))
			handled = true
		case "azure":
			checkers = append(checkers, provider.NewAzureChecker(pID, pc.Slug, pc.APIURL))
			handled = true
		case "gcp":
			checkers = append(checkers, provider.NewGCPChecker(pID, pc.Slug, pc.APIURL))
			handled = true
		case "slack":
			checkers = append(checkers, provider.NewSlackChecker(pID, pc.Slug, pc.APIURL))
			handled = true
		case "stripe":
			checkers = append(checkers, provider.NewStripeChecker(pID, pc.Slug, pc.APIURL))
			handled = true
		}

		if handled {
			continue
		}

		// Prioridad 2 - Checkers genéricos por Tipo
		if pc.Type == "statuspage" || pc.Type == "status_page" || pc.Type == "stripe" {
			checkers = append(checkers, provider.NewStatusPageChecker(pID, pc.Slug, pc.APIURL))
			continue
		}

		log.Printf("Warning: No checker implementation for provider %s (type: %s)", pc.Slug, pc.Type)
	}

	// 6. Iniciar Loop de Polling en Goroutine
	statusCache := service.NewStatusCache(10 * time.Second)
	aggregator := service.NewAggregator(repo, checkers)

	// 6. Iniciar Loop de Polling en Goroutine
	go func() {
		ticker := time.NewTicker(time.Duration(cfg.Polling.IntervalSeconds) * time.Second)
		log.Printf("Polling started. Interval: %ds", cfg.Polling.IntervalSeconds)
		
		// Ejecutar primer check inmediatamente
		aggregator.RunAllChecks(context.Background())

		for range ticker.C {
			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.Polling.TimeoutSeconds)*time.Second)
			aggregator.RunAllChecks(ctx)
			cancel()
		}
	}()

	// 7. Handlers API y Web
	apiHandler := api.NewHandler(repo, statusCache)
	webHandler := web.NewWebHandler(repo, statusCache)

	// Rutas API
	http.HandleFunc("/api/v1/health", apiHandler.Health)
	http.HandleFunc("/api/v1/status", apiHandler.GetStatus)

	// Rutas Web
	http.HandleFunc("/", webHandler.Index)
	http.HandleFunc("/web/search", webHandler.Search)
	http.HandleFunc("/web/refresh", webHandler.Refresh)

	// Servir archivos estáticos
	fs := http.FileServer(http.Dir("web/static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	log.Printf("Server starting on http://0.0.0.0:%d (Accessible from your local network)", cfg.Server.Port)
	if err := http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", cfg.Server.Port), nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func runMigrations(dbPath string) {
	// Lectura simple del archivo SQL
	migrationPath := "internal/storage/migrations/001_initial_schema.sql"
	content, err := os.ReadFile(migrationPath)
	if err != nil {
		log.Printf("Warning: Could not read migrations: %v", err)
		return
	}

	db, err := storage.NewSQLiteRepository(dbPath)
	if err != nil {
		log.Printf("Error opening db for migrations: %v", err)
		return
	}

	// Ejecutar el script SQL
	// Nota: En una implementación de producción usaríamos una librería de migraciones robusta.
	// Pero para el MVP, esto es suficiente y limpio.
	_, err = db.GetDB().Exec(string(content))
	if err != nil {
		log.Printf("Error executing migrations: %v", err)
		return
	}
	log.Println("Migrations executed successfully.")
}
