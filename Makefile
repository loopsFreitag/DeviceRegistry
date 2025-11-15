.PHONY: help setup build up down restart logs migrate-up migrate-down migrate-status migrate-create shell clean build-docker

help:
	@echo "Available commands:"
	@echo "  make setup          - Create common-infra network"
	@echo "  make build          - Build Go binary"
	@echo "  make build-docker   - Build docker images"
	@echo "  make up             - Start services (dev profile)"
	@echo "  make down           - Stop all services"
	@echo "  make restart        - Restart deviceregistry service"
	@echo "  make logs           - View logs"
	@echo "  make shell          - Open shell in container"
	@echo "  make migrate-status - Check migration status"
	@echo "  make migrate-create - Create new migration (name=your_migration_name)"
	@echo "  make clean          - Remove containers and volumes"

setup:
	@docker network create common-infra 2>/dev/null || echo "Network common-infra already exists"

# Build Go binary (called by Dockerfile)
build:
	@echo "Building deviceregistry binary..."
	@mkdir -p build
	@go build -ldflags "-X main.Version=$(VERSION)" -o build/deviceregistry .
	@echo "Binary built successfully at build/deviceregistry"

build-docker:
	docker compose build

up: setup build-docker
	docker compose --profile dev up -d

down:
	docker compose --profile dev --profile fullstack down

restart:
	docker compose restart deviceregistry

logs:
	docker compose logs -f deviceregistry

shell:
	docker compose exec deviceregistry /bin/sh

migrate-status:
	docker compose exec deviceregistry ./deviceregistry migrate status

migrate-create:
	@if [ -z "$(name)" ]; then \
		echo "Error: name is required. Usage: make migrate-create name=your_migration_name"; \
		exit 1; \
	fi
	@mkdir -p db/migrations
	goose -dir ./db/migrations create $(name) sql

clean:
	docker compose --profile dev --profile fullstack down -v
	docker rmi deviceregistry:dev 2>/dev/null || true
	rm -rf build/

docker-update:
	docker compose up -d --build deviceregistry
