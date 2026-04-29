# Makefile para Observatory

.PHONY: run build test clean migrate

run:
	go run cmd/observatory/main.go

build:
	go build -o bin/observatory cmd/observatory/main.go

test:
	go test ./...

clean:
	rm -f observatory.db
	rm -rf bin/

migrate:
	@echo "Las migraciones se ejecutan automáticamente al arrancar el servidor."
