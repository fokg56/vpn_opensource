# VPN Client Makefile

.PHONY: help build run clean test fmt vet lint

help:
	@echo "VPN Client - Доступные команды:"
	@echo ""
	@echo "  make build       - Собрать приложение"
	@echo "  make run         - Запустить приложение"
	@echo "  make dev         - Запустить в режиме разработки"
	@echo "  make clean       - Очистить собранные файлы"
	@echo "  make test        - Запустить тесты"
	@echo "  make fmt         - Отформатировать код"
	@echo "  make vet         - Проверить код (vet)"
	@echo "  make lint        - Запустить linter (требует golangci-lint)"
	@echo "  make deps        - Скачать зависимости"
	@echo "  make help        - Показать эту справку"

BINARY_NAME := vpn-client
MAIN_PATH := ./cmd
GO := go
GOOS := $(shell uname -s | tr A-Z a-z)
GOARCH := amd64

ifeq ($(GOOS),darwin)
	BINARY_NAME := vpn-client
else ifeq ($(GOOS),windows)
	BINARY_NAME := vpn-client.exe
endif

build:
	@echo "🔨 Собираю приложение..."
	@CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) $(GO) build -ldflags="-w -s" -o $(BINARY_NAME) $(MAIN_PATH)
	@echo "✅ Приложение собрано: $(BINARY_NAME)"

run: build
	@echo "🚀 Запускаю приложение..."
	@./$(BINARY_NAME) -addr 0.0.0.0:8080

dev:
	@echo "🔧 Запускаю в режиме разработки..."
	@$(GO) run $(MAIN_PATH) -addr 127.0.0.1:8080

clean:
	@echo "🧹 Очищаю..."
	@$(GO) clean
	@rm -f $(BINARY_NAME) $(BINARY_NAME).exe
	@echo "✅ Очистка завершена"

deps:
	@echo "📦 Скачиваю зависимости..."
	@$(GO) mod download
	@$(GO) mod tidy
	@echo "✅ Зависимости скачаны"

test:
	@echo "🧪 Запускаю тесты..."
	@$(GO) test -v ./...

fmt:
	@echo "📝 Форматирую код..."
	@$(GO) fmt ./...
	@echo "✅ Форматирование завершено"

vet:
	@echo "🔍 Проверяю код (vet)..."
	@$(GO) vet ./...
	@echo "✅ Проверка завершена"

lint:
	@echo "🔍 Запускаю linter..."
	@golangci-lint run ./...
	@echo "✅ Lint проверка завершена"

.DEFAULT_GOAL := help
