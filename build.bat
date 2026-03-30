@echo off
REM Скрипт для сборки VPN Client на Windows

echo 🔨 Сборка VPN Client...

REM Проверяем, установлен ли Go
where go >nul 2>nul
if %errorlevel% neq 0 (
    echo ❌ Go не установлен. Пожалуйста, установите Go 1.22 или выше.
    exit /b 1
)

REM Проверяем версию Go
for /f "tokens=3" %%i in ('go version') do set GO_VERSION=%%i
echo ✅ Найден Go версии %GO_VERSION%

REM Скачиваем зависимости
echo 📦 Скачивание зависимостей...
go mod download

if %errorlevel% neq 0 (
    echo ❌ Ошибка скачивания зависимостей
    exit /b 1
)

REM Собираем приложение
echo 🔧 Компиляция приложения...

set BINARY_NAME=vpn-client.exe
set CGO_ENABLED=0
set GOOS=windows
set GOARCH=amd64

go build -ldflags="-w -s" -o %BINARY_NAME% .\cmd

if %errorlevel% eq 0 (
    echo ✅ Сборка успешна!
    echo 📁 Бинарный файл: .\%BINARY_NAME%
    echo.
    echo 🚀 Для запуска выполните:
    echo    %BINARY_NAME%
    echo.
    echo    или с указанием адреса:
    echo    %BINARY_NAME% -addr 0.0.0.0:8000
) else (
    echo ❌ Ошибка при компиляции
    exit /b 1
)

pause
