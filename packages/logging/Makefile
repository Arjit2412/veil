# Veil Logging Infrastructure Makefile

.PHONY: help start stop restart status logs build test clean

# Default target
help: ## Show this help message
	@echo "Veil Logging Infrastructure"
	@echo ""
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

start: ## Start the observability stack (Loki + Grafana + Promtail)
	@echo "Starting observability stack..."
	@./scripts/start-observability.sh

stop: ## Stop the observability stack
	@echo "Stopping observability stack..."
	@./scripts/start-observability.sh --stop

restart: ## Restart the observability stack
	@echo "Restarting observability stack..."
	@./scripts/start-observability.sh --restart

status: ## Show status of running services
	@./scripts/start-observability.sh --status

logs: ## Show logs from all services
	@docker-compose logs -f

logs-loki: ## Show Loki logs
	@docker-compose logs -f loki

logs-grafana: ## Show Grafana logs
	@docker-compose logs -f grafana

logs-promtail: ## Show Promtail logs
	@docker-compose logs -f promtail

build: ## Build the logging Go module
	@echo "Building logging module..."
	@go mod tidy
	@go build ./...

test: ## Run tests for the logging module
	@echo "Running tests..."
	@go test ./...

test-integration: ## Run integration tests (requires running stack)
	@echo "Running integration tests..."
	@go test -tags=integration ./...

clean: ## Clean up Docker resources
	@echo "Cleaning up Docker resources..."
	@docker-compose down -v --remove-orphans
	@docker system prune -f

dashboard-import: ## Import additional dashboards
	@echo "Importing dashboards..."
	@curl -X POST \
		-H "Content-Type: application/json" \
		-d @config/grafana/dashboards/veil-api-gateway.json \
		http://admin:admin@localhost:3000/api/dashboards/db

dashboard-export: ## Export current dashboards
	@echo "Exporting dashboards..."
	@mkdir -p exports
	@curl -H "Authorization: Bearer admin:admin" \
		http://localhost:3000/api/dashboards/uid/veil-api-gateway \
		| jq '.dashboard' > exports/veil-api-gateway-$(shell date +%Y%m%d-%H%M%S).json

config-validate: ## Validate configuration files
	@echo "Validating Loki configuration..."
	@docker run --rm -v $(PWD)/config/loki-config.yaml:/etc/loki/local-config.yaml \
		grafana/loki:2.9.2 -config.file=/etc/loki/local-config.yaml -verify-config
	@echo "Validating Promtail configuration..."
	@docker run --rm -v $(PWD)/config/promtail-config.yaml:/etc/promtail/config.yml \
		grafana/promtail:2.9.2 -config.file=/etc/promtail/config.yml -dry-run

health-check: ## Check health of all services
	@echo "Checking service health..."
	@echo -n "Loki: "
	@curl -s -o /dev/null -w "%{http_code}" http://localhost:3100/ready && echo " ✓" || echo " ✗"
	@echo -n "Grafana: "
	@curl -s -o /dev/null -w "%{http_code}" http://localhost:3000/api/health && echo " ✓" || echo " ✗"
	@echo -n "Promtail: "
	@curl -s -o /dev/null -w "%{http_code}" http://localhost:9080/metrics && echo " ✓" || echo " ✗"

install-deps: ## Install required dependencies
	@echo "Installing dependencies..."
	@command -v docker >/dev/null 2>&1 || { echo "Docker is required but not installed. Please install Docker."; exit 1; }
	@command -v docker-compose >/dev/null 2>&1 || { echo "Docker Compose is required but not installed. Please install Docker Compose."; exit 1; }
	@command -v curl >/dev/null 2>&1 || { echo "curl is required but not installed. Please install curl."; exit 1; }
	@command -v jq >/dev/null 2>&1 || { echo "jq is recommended for dashboard management. Please install jq."; }
	@echo "All dependencies are available."

dev-setup: install-deps build ## Set up development environment
	@echo "Setting up development environment..."
	@chmod +x scripts/start-observability.sh
	@echo "Development environment ready!"

quick-start: dev-setup start ## Complete setup and start (recommended for first run)
	@echo ""
	@echo "🎉 Quick start complete!"
	@echo "Visit http://localhost:3000 to access Grafana (admin/admin)"

# Development targets
go-mod-init: ## Initialize Go module (if needed)
	@go mod init github.com/try-veil/veil/packages/logging 2>/dev/null || true

go-mod-tidy: ## Tidy Go module dependencies
	@go mod tidy

fmt: ## Format Go code
	@go fmt ./...

lint: ## Run Go linter
	@golangci-lint run ./... 2>/dev/null || echo "golangci-lint not available, skipping..."

# Utility targets
ports-check: ## Check if required ports are available
	@echo "Checking port availability..."
	@for port in 3000 3100 9080; do \
		if nc -z localhost $$port 2>/dev/null; then \
			echo "Port $$port: ❌ In use"; \
		else \
			echo "Port $$port: ✅ Available"; \
		fi; \
	done

backup-config: ## Backup current configuration
	@echo "Backing up configuration..."
	@mkdir -p backups
	@tar -czf backups/veil-logging-config-$(shell date +%Y%m%d-%H%M%S).tar.gz config/

restore-config: ## Restore configuration from backup (use BACKUP=filename)
	@if [ -z "$(BACKUP)" ]; then \
		echo "Please specify backup file: make restore-config BACKUP=filename.tar.gz"; \
		exit 1; \
	fi
	@echo "Restoring configuration from $(BACKUP)..."
	@tar -xzf backups/$(BACKUP)

# Docker targets
docker-pull: ## Pull latest Docker images
	@docker-compose pull

docker-build: ## Build custom Docker images (if any)
	@docker-compose build

docker-clean: ## Clean Docker resources more aggressively
	@docker-compose down -v --remove-orphans
	@docker image prune -f
	@docker volume prune -f