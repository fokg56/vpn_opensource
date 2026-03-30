# VPN Client Installation & Usage Guide

## Быстрый старт

### Linux / macOS

```bash
# Сборка
bash build.sh

# Запуск
./vpn-client -addr 0.0.0.0:8080

# Откройте браузер
http://localhost:8080
```

### Windows

```bash
# Сборка (двойной клик на build.bat или в CMD)
build.bat

# Запуск
vpn-client.exe -addr 0.0.0.0:8080

# Откройте браузер
http://localhost:8080
```

## Системные требования

- **OS**: Linux, macOS или Windows
- **Go**: 1.22 или выше
- **RAM**: Минимум 128 MB
- **Disk**: Минимум 50 MB для компиляции

## Установка зависимостей

### Linux (Ubuntu/Debian)

```bash
# Установка Go
wget https://go.dev/dl/go1.22.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.22.0.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
```

### macOS

```bash
# Через Homebrew
brew install go

# Или вручную с go.dev
```

### Windows

Скачайте установщик с https://go.dev/dl/ и запустите инсталлер.

## Использование веб-интерфейса

### 1. Параметры подключения

| Параметр | Описание | Пример |
|----------|---------|--------|
| Адрес сервера | IP или домен Xray сервера | example.com, 192.168.1.1 |
| Порт | Порт сервера VLESS | 443 |
| UUID | Уникальный идентификатор | 00000000-0000-0000-0000-000000000000 |
| ChaCha20 Ключ | 32-символьный ключ для шифрования | `openssl rand -hex 32` |
| Локальный порт | Порт локального SOCKS5 прокси | 1080 |

### 2. Получение UUID и ключа

**Генерирование UUID**:
```bash
# Linux/macOS
uuidgen

# Windows PowerShell
[guid]::NewGuid().ToString()
```

**Генерирование ChaCha20 ключа**:
```bash
# Linux/macOS
openssl rand -hex 32

# Windows (PowerShell)
$bytes = New-Object byte[] 32
[System.Security.Cryptography.RNGCryptoServiceProvider]::Create().GetBytes($bytes)
[System.BitConverter]::ToString($bytes) -replace '-'
```

### 3. Подключение

1. Заполните все поля параметров
2. Нажмите **"Подключиться"**
3. Дождитесь статуса **"✅ Подключено"**
4. Используйте прокси `127.0.0.1:1080` в ваших приложениях

### 4. Использование прокси

**Firefox**:
- Settings → Network Settings → Manual proxy configuration
- SOCKS Host: 127.0.0.1, Port: 1080

**Chrome** (требует расширение или системные настройки):
- Используйте системные настройки прокси

**curl**:
```bash
curl -x socks5://127.0.0.1:1080 https://example.com
```

**Python**:
```python
import requests
from requests.adapters import HTTPAdapter
from requests.packages.urllib3.util.ssl_ import create_urllib3_context

session = requests.Session()
session.proxies = {
    'http': 'socks5://127.0.0.1:1080',
    'https': 'socks5://127.0.0.1:1080'
}
response = session.get('https://example.com')
```

## Сетап Xray сервера

### 1. Установка Xray

```bash
# Linux
curl -L https://github.com/XTLS/Xray-core/releases/download/v1.8.0/Xray-linux-64.zip -o xray.zip
unzip xray.zip
chmod +x xray
```

### 2. Конфигурация

Используйте файл `xray-server-config.json` из проекта:

```bash
./xray -c xray-server-config.json
```

### 3. Замена UUID и сертификатов

Отредактируйте `xray-server-config.json`:

```json
{
  "clients": [
    {
      "id": "YOUR-UUID-HERE",
      "level": 0
    }
  ]
}
```

## Отладка

### Просмотр логов

Логи выводятся в реальном времени через веб-интерфейс.

### Проверка подключения

```bash
# Проверка доступности сервера
telnet example.com 443

# Linux/macOS
nc -zv example.com 443

# Windows PowerShell
Test-NetConnection -ComputerName example.com -Port 443
```

### Включение DNS логирования

Отредактируйте конфиг для более подробного логирования:

```json
{
  "log": {
    "loglevel": "debug"
  }
}
```

## Проблемы и решения

| Проблема | Решение |
|----------|---------|
| Port already in use | Измените локальный порт или убейте процесс на порту |
| Connection refused | Проверьте адрес и порт сервера, убедитесь что сервер запущен |
| Invalid key | Убедитесь что ключ 32 символа |
| No logs appearing | Проверьте WebSocket соединение, включите JavaScript |
| DNS not working | Используйте DoH или настройте DNS прокси отдельно |

## Производительность

Рекомендуемые настройки для различных сценариев:

### Для web browsing
- Ключ: Любой 32-символьный ключ
- Порт: 443
- Сервер: Стабильный сервер с хорошей пропускной способностью

### Для потокового видео
- Локальный порт: Любой свободный
- Сервер: Выберите сервер близко к вашему местоположению

### Для файлообмена
- Используйте стабильное соединение
- Проверьте пропускную способность сервера

## Безопасность

⚠️ **Важно**:
- Никогда не используйте слабые ключи (123456, password и т.д.)
- Используйте HTTPS для веб-интерфейса в production
- Защитите доступ к приложению паролем
- Используйте TLS 1.3 между клиентом и сервером
- Регулярно обновляйте Go и зависимости

### Генерирование безопасного ключа

```bash
# Криптографически стойкий ключ (рекомендуется)
openssl rand -hex 32

# Или используйте встроенный генератор (в most OS)
head -c 32 /dev/urandom | xxd -p
```

## Дополнительно

### Автозапуск (Linux)

Создайте systemd сервис:

```ini
# /etc/systemd/system/vpn-client.service
[Unit]
Description=VPN Client
After=network.target

[Service]
Type=simple
User=vpn
ExecStart=/usr/local/bin/vpn-client -addr 0.0.0.0:8080
Restart=always

[Install]
WantedBy=multi-user.target
```

```bash
sudo systemctl enable vpn-client
sudo systemctl start vpn-client
```

### Использование в Docker

```dockerfile
FROM golang:1.22-alpine as builder
WORKDIR /app
COPY . .
RUN go build -o vpn-client ./cmd

FROM alpine:latest
COPY --from=builder /app/vpn-client /usr/local/bin/
EXPOSE 8080
CMD ["vpn-client", "-addr", "0.0.0.0:8080"]
```

## Поддержка

Для проблем и предложений создавайте issue на GitHub.

## Лицензия

MIT License - см. LICENSE файл

---

**Версия**: 1.0.0  
**Последнее обновление**: 2024
