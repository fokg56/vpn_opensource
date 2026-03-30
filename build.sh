#!/bin/bash

# Скрипт для сборки VPN Client

echo "🔨 Сборка VPN Client..."

# Проверяем, установлен ли Go
if ! command -v go &> /dev/null; then
    echo "❌ Go не установлен. Пожалуйста, установите Go 1.22 или выше."
    exit 1
fi

# Проверяем версию Go
GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
echo "✅ Найден Go версии $GO_VERSION"

# Скачиваем зависимости
echo "📦 Скачивание зависимостей..."
go mod download

if [ $? -ne 0 ]; then
    echo "❌ Ошибка скачивания зависимостей"
    exit 1
fi

# Собираем приложение
echo "🔧 Компиляция приложения..."

GOOS=$(uname -s | tr '[:upper:]' '[:lower:]')
if [ "$GOOS" == "darwin" ]; then
    GOOS="darwin"
elif [ "$GOOS" == "linux" ]; then
    GOOS="linux"
fi

BINARY_NAME="vpn-client"
if [ "$GOOS" == "windows" ]; then
    BINARY_NAME="vpn-client.exe"
fi

CGO_ENABLED=0 GOOS=$GOOS GOARCH=amd64 go build -ldflags="-w -s" -o $BINARY_NAME ./cmd

if [ $? -eq 0 ]; then
    echo "✅ Сборка успешна!"
    echo "📁 Бинарный файл: ./$BINARY_NAME"
    echo ""
    echo "🚀 Для запуска выполните:"
    echo "   ./$BINARY_NAME"
    echo ""
    echo "   или с указанием адреса:"
    echo "   ./$BINARY_NAME -addr 0.0.0.0:8000"
else
    echo "❌ Ошибка при компиляции"
    exit 1
fi
