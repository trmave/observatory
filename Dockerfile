# Etapa de construcción
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copiar archivos de dependencias
COPY go.mod go.sum ./
RUN go mod download

# Copiar el resto del código
COPY . .

# Compilar el binario (sin CGO)
RUN CGO_ENABLED=0 GOOS=linux go build -o observatory cmd/observatory/main.go

# Etapa final (Runtime)
FROM alpine:latest

WORKDIR /root/

# Copiar el binario desde la etapa de construcción
COPY --from=builder /app/observatory .
COPY --from=builder /app/configs ./configs
COPY --from=builder /app/internal/web/templates ./internal/web/templates
COPY --from=builder /app/internal/storage/migrations ./internal/storage/migrations
COPY --from=builder /app/web/static ./web/static

# Exponer el puerto
EXPOSE 8080

# Comando para ejecutar
CMD ["./observatory"]
