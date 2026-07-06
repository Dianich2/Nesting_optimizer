include .env
export

APP_NAME := nesting_optimizer
MIGRATIONS_DIR := migrations
MIGRATE_DSN := postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSL_MODE)

.PHONY: up down restart logs ps run build clean \
	    migrate-up migrate-down migrate-create migrate-status migrate-reset \
		swag dev lint fmt

up:
	docker compose up -d

down:
	docker compose down

restart: down up

logs:
	docker compose logs -f

ps:
	docker compose ps

run:
	go run ./cmd/api

build:
	go build -o bin/$(APP_NAME) ./cmd/api

clean:
	rm -rf bin

migrate-up:
	goose -dir $(MIGRATIONS_DIR) postgres "$(MIGRATE_DSN)" up

migrate-down:
	goose -dir $(MIGRATIONS_DIR) postgres "$(MIGRATE_DSN)" down

migrate-create:
	goose -dir $(MIGRATIONS_DIR) create $(name) sql

migrate-status:
	goose -dir $(MIGRATIONS_DIR) postgres "$(MIGRATE_DSN)" status

migrate-reset:
	goose -dir $(MIGRATIONS_DIR) postgres "$(MIGRATE_DSN)" reset

swag:
	swag init -g ./cmd/api/main.go -o ./docs

dev: up migrate-up swag run

lint:
	golangci-lint run ./...

fmt:
	golangci-lint fmt ./...



