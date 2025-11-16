.PHONY: help setup build swagger up down restart logs migrate-up shell clean build-docker update

help:
	@echo "Available commands:"
	@echo "  make setup          - Create common-infra network"
	@echo "  make swagger        - Generate Swagger docs"
	@echo "  make build          - Build Go binary"
	@echo "  make build-docker   - Build docker images"
	@echo "  make up             - Start services (dev profile)"
	@echo "  make down           - Stop all services"
	@echo "  make restart        - Restart deviceregistry service"
	@echo "  make logs           - View logs"
	@echo "  make shell          - Open shell in container"
	@echo "  make migrate-up     - Migration up"
	@echo "  make clean          - Remove containers and volumes"
	@echo "  make update         - Rebuild and update running container"

setup:
	@docker network create common-infra 2>/dev/null || echo "Network common-infra already exists"

# Generate Swagger documentation
swagger:
	@echo "Generating Swagger documentation..."
	@swag init
	@echo "Swagger docs generated successfully"

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

migrate-up:
	docker compose exec deviceregistry ./deviceregistry migrate up

clean:
	docker compose --profile dev --profile fullstack down -v
	docker rmi deviceregistry:dev 2>/dev/null || true
	rm -rf build/

update:
	docker compose up -d --build deviceregistry
