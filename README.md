# 🔭 Observatory

Observatory es una plataforma de monitoreo unificada para servicios Cloud y SaaS. Proporciona una vista en tiempo real y normalizada del estado de salud de proveedores críticos.

## ✨ Características
- **Multi-Nube**: AWS, Azure, GCP y GitHub integrados.
- **Normalización**: Estados heterogéneos mapeados a un estándar común (OK, Warning, Critical).
- **Dashboard Premium**: Interfaz en modo oscuro con HTMX para interactividad fluida.
- **Caché Inteligente**: Reducción de carga en DB mediante caché en memoria.
- **Docker Ready**: Listo para desplegar en cualquier entorno con Docker.

## 🚀 Instalación y Uso

### Local (Go)
1. `go mod tidy`
2. `make run`
3. Visita `http://localhost:8080`

### Docker
```bash
docker-compose up --build
```

## 🛠️ Estructura del Proyecto
- `cmd/observatory`: Punto de entrada del servidor.
- `internal/provider`: Adaptadores para APIs externas.
- `internal/service`: Lógica de agregación y caché.
- `internal/web`: Templates HTML y lógica del dashboard.

## 🔌 Añadir nuevos proveedores
Para añadir un proveedor (ej. Slack):
1. Crea `internal/provider/slack.go` implementando la interfaz `Checker`.
2. Añade la configuración en `configs/config.yaml`.
3. Regístralo en el switch de `main.go`.
