# Observatory: Especificación Técnica y de Diseño

## 1. Resumen Ejecutivo
Observatory es un monorepo en Go diseñado para el monitoreo consolidado de proveedores Cloud y SaaS. Su objetivo es normalizar estados heterogéneos y ofrecer una vista operativa unificada.

## 2. Requisitos Técnicos

### Backend (Go)
- **Checker Interface**: Todos los proveedores deben implementar `Check(ctx context.Context) (domain.StatusResult, error)`.
- **Agregador Concurrente**: Procesamiento paralelo de proveedores con timeouts estrictos por cada uno.
- **Normalización**: Mapeo de estados externos a: `ok`, `warning`, `critical`, `error`, `unknown`.
- **Persistencia**: SQLite con migraciones integradas.

### Base de Datos (SQLite)
- **providers**: Catálogo de servicios (id, name, slug, provider_type, enabled).
- **status_checks**: Histórico de ejecuciones (provider_id, status, latency, raw_response, created_at).
- **incidents**: Tracking de fallos (provider_id, title, status, started_at, resolved_at).
- **provider_errors**: Logs de errores técnicos de integración.

### Frontend (Go SSR + HTMX)
- Dashboard con Cards por proveedor.
- Búsqueda y filtrado dinámico mediante HTMX.
- Tema oscuro/claro con estética premium.

## 3. Diseño de API (v1)
- `GET /api/v1/health`: Estado del sistema.
- `GET /api/v1/status`: Resumen actual consolidado.
- `GET /api/v1/incidents`: Listado de incidentes activos y recientes.
- `GET /api/v1/providers/{slug}/history`: Serie temporal para gráficos de disponibilidad.

## 4. Estrategia de Integración
| Proveedor | Fuente | Tipo |
|-----------|--------|------|
| GitHub    | Status API | JSON |
| AWS       | Health API | JSON/RSS |
| Azure     | Status Page | Scraping/API |
| GCP       | Incidents API | JSON |

## 5. Roadmap
- **Fase 1**: Diseño y Especificación (Completada).
- **Fase 2**: Scaffolding y Estructura.
- **Fase 3**: Backend MVP (GitHub + SQLite).
- **Fase 4**: Multi-proveedor (AWS, Azure, GCP).
- **Fase 5**: Frontend y Dashboard.
- **Fase 6**: Pulido y Optimización.
