.PHONY: help up down reset system-prune install tidy deps get vet fmt lint lint-fix clean test test-coverage build-migrate build-seed build deploy-build db-drop db-migrate db-seed db-down db-force db-reset migrate-create logs logs-api logs-postgres
# -------------------------------------------------------
# Variáveis
# -------------------------------------------------------
ENV_FILE = .env
include $(ENV_FILE)

# Cores
GREEN  = \033[0;32m
YELLOW = \033[1;33m
RED    = \033[0;31m
NC     = \033[0m

# Atalhos docker compose
COMPOSE  = docker compose -f docker-compose.yml --env-file $(ENV_FILE)

API      = $(COMPOSE) exec api
POSTGRES = $(COMPOSE) exec postgres

# -------------------------------------------------------
# Help
# -------------------------------------------------------
help: ## Mostra esta mensagem de ajuda
	@echo "$(GREEN)CRM Cia da Vacina - Comandos disponíveis:$(NC)"
	@awk -F'## ' '/^[a-zA-Z_-]+:.*## / {split($$1, a, ":"); printf "  $(YELLOW)%-18s$(NC) - %s\n", a[1], $$2}' Makefile

# -------------------------------------------------------
# Ambiente Docker
# -------------------------------------------------------
up: ## Inicia o ambiente de desenvolvimento
	@echo "$(YELLOW)Starting development environment...$(NC)"
	$(COMPOSE) up -d --build
	@echo "$(GREEN)Development environment started!$(NC)"

down: ## Para o ambiente de desenvolvimento
	@echo "$(YELLOW)Stopping development environment...$(NC)"
	$(COMPOSE) down
	@echo "$(GREEN)Development environment stopped!$(NC)"

reset: down up ## Derruba e recria o ambiente de desenvolvimento (sem mexer no banco)

system-prune: ## Limpa o sistema de containers Docker (CUIDADO: apaga todos os containers, imagens e volumes de todo o sistema, não apenas os do projeto)
	@echo "$(YELLOW)Pruning containers...$(NC)"
	docker container prune -f
	@echo "$(GREEN)Containers pruned!$(NC)"
	@echo "$(YELLOW)Pruning images...$(NC)"
	docker image prune -a -f
	@echo "$(GREEN)Images pruned!$(NC)"
	@echo "$(YELLOW)Pruning volumes...$(NC)"
	docker volume prune -f
	@echo "$(GREEN)Volumes pruned!$(NC)"
	@echo "$(YELLOW)Pruning networks...$(NC)"
	docker network prune -f
	@echo "$(GREEN)Networks pruned!$(NC)"

# -------------------------------------------------------
# Dependências
# -------------------------------------------------------
install: ## Instala as dependências do ambiente de desenvolvimento
	@echo "$(YELLOW)Installing dependencies...$(NC)"
	$(API) go mod download
	@echo "$(GREEN)Dependencies installed!$(NC)"

tidy: ## Organiza as dependências do ambiente de desenvolvimento
	@echo "$(YELLOW)Organizing dependencies...$(NC)"
	$(API) go mod tidy
	@echo "$(GREEN)Dependencies organized!$(NC)"

deps: install tidy ## Instala e organiza as dependências

get: ## Baixa uma dependência passada como parâmetro
	@read -p "Package: " package; \
	$(API) go get $$package
	@echo "$(GREEN)Package got!$(NC)"

# -------------------------------------------------------
# GoLang
# -------------------------------------------------------
vet: ## Verifica erros de compilação
	@echo "$(YELLOW)Checking for errors...$(NC)"
	$(API) go vet ./...
	@echo "$(GREEN)No errors found!$(NC)"

fmt: ## Formata o código
	@echo "$(YELLOW)Formatting code...$(NC)"
	$(API) go fmt ./...
	@echo "$(GREEN)Code formatted!$(NC)"

lint: ## Verifica erros de estilo (requer golangci-lint na imagem)
	@echo "$(YELLOW)Checking for style errors...$(NC)"
	$(API) golangci-lint run ./...
	@echo "$(GREEN)No style errors found!$(NC)"

lint-fix: ## Corrige erros de estilo
	@echo "$(YELLOW)Fixing style errors...$(NC)"
	$(API) golangci-lint run --fix ./...
	@echo "$(GREEN)Style errors fixed!$(NC)"

clean: ## Limpa os binários
	@echo "$(YELLOW)Cleaning binaries...$(NC)"
	rm -rf bin/
	@echo "$(GREEN)Binaries cleaned!$(NC)"

test: ## Roda os testes
	@echo "$(YELLOW)Running tests...$(NC)"
	$(API) go test -p 1 -v ./...
	@echo "$(GREEN)Tests ran!$(NC)"

# -p 1: várias suites de teste de usecase batem no Postgres real (não mocks)
# e mutam app_settings (singleton, id=1) e meta_channel_configs (linha global
# única por canal em instagram/facebook) — estado compartilhado entre
# pacotes. Rodar pacotes em paralelo (default do go test) causa corrida
# nessas tabelas; -p 1 serializa a execução por pacote e evita isso. Ver
# internal/testutil/testutil.go (SnapshotAppSettings).
test-coverage: ## Roda os testes e gera o relatório de cobertura
	@echo "$(YELLOW)Running tests and generating coverage report...$(NC)"
	$(API) go test -p 1 -coverprofile=coverage.out ./...
	$(API) go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)Tests ran and coverage report generated!$(NC)"

# -------------------------------------------------------
# Build
# -------------------------------------------------------
build-migrate: ## Builda o cmd/migrate
	@echo "$(YELLOW)Building migrate...$(NC)"
	$(API) sh -c 'CGO_ENABLED=0 go build -buildvcs=false -o bin/migrate ./cmd/migrate'
	@echo "$(GREEN)Migrate built!$(NC)"

build-seed: ## Builda o cmd/seed
	@echo "$(YELLOW)Building seed...$(NC)"
	$(API) sh -c 'CGO_ENABLED=0 go build -buildvcs=false -o bin/seed ./cmd/seed'
	@echo "$(GREEN)Seed built!$(NC)"

build: build-migrate build-seed ## Builda o serviço api
	@echo "$(YELLOW)Building server...$(NC)"
	$(API) sh -c 'CGO_ENABLED=0 go build -buildvcs=false -o bin/server ./cmd/server'
	@echo "$(GREEN)Server built!$(NC)"

deploy-build: ## Builda a imagem Docker scratch simulando o build de deploy (uso: make deploy-build IMAGE=meu-nome[:tag])
	@if [ -z "$(IMAGE)" ]; then \
		echo "$(RED)Error: variável IMAGE não informada.$(NC)"; \
		echo "Uso: make deploy-build IMAGE=<nome-da-imagem>[:tag]"; \
		exit 1; \
	fi
	@echo "$(YELLOW)Building Docker image $(IMAGE)...$(NC)"
	docker build -t $(IMAGE) -f Dockerfile .
	@echo "$(GREEN)Docker image $(IMAGE) built!$(NC)"

# -------------------------------------------------------
# Banco de dados
# -------------------------------------------------------
db-drop: build-migrate ## Builda o cmd/migrate e dropa o banco de dados
	@echo "$(YELLOW)Dropping database...$(NC)"
	$(API) ./bin/migrate drop
	@echo "$(GREEN)Database dropped!$(NC)"

db-migrate: build-migrate ## Builda o cmd/migrate e executa as migrations
	@echo "$(YELLOW)Running migrations...$(NC)"
	$(API) ./bin/migrate up
	@echo "$(GREEN)Migrations complete!$(NC)"

db-seed: build-seed ## Builda o cmd/seed e executa os seeders
	@echo "$(YELLOW)Running seeders...$(NC)"
	$(API) ./bin/seed dev
	@echo "$(GREEN)Seeders complete!$(NC)"

db-down: build-migrate ## Builda o cmd/migrate e desfaz as migrations
	@echo "$(YELLOW)Rolling back migrations...$(NC)"
	$(API) ./bin/migrate down
	@echo "$(GREEN)Migrations rolled back!$(NC)"

db-force: build-migrate ## Builda o cmd/migrate e força a versão das migrations (uso: make db-force V=<version>)
	@if [ -z "$(V)" ]; then \
		echo "$(RED)Error: V variable not provided.$(NC)"; \
		echo "Usage: make db-force V=<version>"; \
		exit 1; \
	fi
	@echo "$(YELLOW)Forcing migrations...$(NC)"
	$(API) ./bin/migrate force $(V)
	@echo "$(GREEN)Migrations forced!$(NC)"

db-reset: db-drop db-migrate db-seed ## Dropa, recria e aplica migrations e seeders no banco de dados

migrate-create: ## Cria uma nova migration
	@read -p "Migration name: " name; \
	$(API) migrate create -ext sql -dir ./conf/migrations -seq $$name

# -------------------------------------------------------
# Logs
# -------------------------------------------------------
logs: ## Exibe os logs do ambiente de desenvolvimento
	@echo "$(YELLOW)Showing development environment logs...$(NC)"
	$(COMPOSE) logs -f
	@echo "$(GREEN)Development environment logs shown!$(NC)"

logs-api: ## Exibe os logs do serviço api
	@echo "$(YELLOW)Showing api service logs...$(NC)"
	$(COMPOSE) logs api -f
	@echo "$(GREEN)Api service logs shown!$(NC)"

logs-postgres: ## Exibe os logs do serviço postgres
	@echo "$(YELLOW)Showing postgres service logs...$(NC)"
	$(COMPOSE) logs postgres -f
	@echo "$(GREEN)Postgres service logs shown!$(NC)"
