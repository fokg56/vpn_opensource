# VPN Client - Go Web Application

Полнофункциональный VPN-клиент на языке Go с веб-интерфейсом. Поддерживает шифрование ChaCha20 и протокол VLESS для совместимости с Xray-серверами.

## Возможности

-  Шифрование трафика ChaCha20
-  Протокол VLESS (совместимость с Xray)
-  Веб-интерфейс для управления подключением
-  Логи в реальном времени через WebSocket
-  Модульная архитектура
-  Проверка корректности ключей
-  Безопасная генерация nonce

## Требования

- Go 1.22 или выше
- `gorilla/websocket` для WebSocket поддержки
- `golang.org/x/crypto` для ChaCha20

## Установка и сборка

### 1. Клонирование проекта

```bash
git clone https://github.com/user/vpn-client.git
cd vpn-client
```

### 2. Установка зависимостей

```bash
go mod download
```

### 3. Сборка

```bash
# Для Linux/macOS
go build -o vpn-client ./cmd

# Для Windows
go build -o vpn-client.exe ./cmd
```

### 4. Запуск

```bash
# Запуск с адресом по умолчанию (0.0.0.0:8080)
./vpn-client

# Запуск с определенным адресом и портом
./vpn-client -addr 127.0.0.1:3000
```

Откройте браузер и перейдите на `http://localhost:8080`

## Структура проекта

```
vpn-client/
├── cmd/
│   └── main.go              # Точка входа приложения
├── internal/
│   ├── crypto/
│   │   └── chacha20.go      # Реализация ChaCha20
│   ├── vless/
│   │   └── protocol.go      # Реализация VLESS протокола
│   ├── client/
│   │   └── tunnelmanager.go # Управление VPN туннелем
│   └── web/
│       └── server.go        # HTTP сервер и WebSocket
├── web/
│   └── static/              # Статические файлы (если нужны)
├── go.mod                   # Go модули
├── go.sum                   # Checksums
└── README.md               # Этот файл
```

## Использование

### 1. Заполнение параметров

**Адрес сервера**: IP или доменное имя Xray сервера  
**Порт**: Порт Xray сервера (обычно 443)  
**UUID**: UUID пользователя (генерируется на сервере)  
**ChaCha20 Ключ**: 32-символьный ключ для шифрования  
**Локальный порт**: Порт локального SOCKS5 прокси (1080 по умолчанию)

### 2. Подключение

1. Заполните все параметры
2. Нажмите кнопку "Подключиться"
3. Дождитесь статуса "Подключено"
4. Используйте локальный прокси на `127.0.0.1:1080`

### 3. Отключение

Нажмите кнопку "Отключиться" или закройте приложение.

## Конфигурация Xray сервера

Пример конфигурации Xray сервера для тестирования:

```json
{
  "log": {
    "loglevel": "info"
  },
  "inbounds": [
    {
      "port": 443,
      "protocol": "vless",
      "settings": {
        "clients": [
          {
            "id": "00000000-0000-0000-0000-000000000000",
            "level": 0,
            "alterId": 0
          }
        ],
        "decryption": "none"
      },
      "streamSettings": {
        "network": "tcp",
        "tcpSettings": {},
        "security": "tls",
        "tlsSettings": {
          "serverName": "example.com",
          "certificates": [
            {
              "certificateFile": "/path/to/cert.crt",
              "keyFile": "/path/to/key.key"
            }
          ]
        }
      }
    }
  ],
  "outbounds": [
    {
      "protocol": "freedom",
      "settings": {}
    }
  ]
}
```

## API Endpoints

### GET /
Возвращает главную HTML страницу с веб-интерфейсом.

### POST /api/connect
Инициирует подключение к VPN серверу.

**Body (JSON)**:
```json
{
  "server_addr": "example.com",
  "server_port": 443,
  "uuid": "00000000-0000-0000-0000-000000000000",
  "key": "32-символьный-ключ-для-chacha20",
  "local_port": 1080
}
```

**Response**:
```json
{
  "status": "connected",
  "message": "Подключено к example.com:443"
}
```

### POST /api/disconnect
Отключается от VPN сервера.

**Response**:
```json
{
  "status": "disconnected",
  "message": "Отключено"
}
```

### GET /api/status
Получает текущий статус подключения.

**Response**:
```json
{
  "status": "connected" | "disconnected"
}
```

### WebSocket /ws
WebSocket соединение для получения логов и обновлений статуса в реальном времени.

**Message Format**:
```json
{
  "type": "log",
  "data": "[15:04:05] Сообщение лога"
}
```

```json
{
  "type": "status",
  "data": "connected" | "disconnected"
}
```

## Безопасность

- ✅ Использование криптографически стойких библиотек (`golang.org/x/crypto`)
- ✅ Безопасная генерация random nonce
- ✅ Валидация входных данных
- ✅ Проверка размера ключа (32 байта для ChaCha20)
- ✅ HTTPS готовность (использует стандартный Go http пакет)
- ⚠️ **Warning**: Используйте HTTPS в production окружении

## Генерирование UUID

```bash
# Linux/macOS
uuidgen

# Windows PowerShell
[guid]::NewGuid()

# Online
Посетите https://www.uuidgenerator.net/
```

## Генерирование ChaCha20 ключа

```bash
# Linux/macOS
openssl rand -hex 32

# Python
python3 -c "import os; print(os.urandom(32).hex())"

# Go
go run -c "package main; func main() { crypto/rand.Read(...) }"
```

## Troubleshooting

### Ошибка: "Ключ должен быть 32 байта"
Убедитесь, что вы предоставили 32-символьный ключ. Если вы используете текстовый ключ, приложение автоматически дополнит его нулями.

### Ошибка подключения к серверу
- Проверьте адрес и порт сервера
- Убедитесь, что сервер Xray запущен и доступен
- Проверьте брандмауэр и сетевые настройки

### WebSocket не подключается
- Убедитесь, что вы используете правильный протокол (ws для http, wss для https)
- Проверьте консоль браузера на ошибки

### Нет логов
- Убедитесь, что JavaScript включен в браузере
- Проверьте консоль браузера на ошибки
- Убедитесь, что WebSocket соединение установлено


## Контрибьютинг

Приветствуются pull requests. Для больших изменений сначала откройте issue для обсуждения.

## Разработка

### Запуск в режиме разработки

```bash
go run ./cmd -addr localhost:8080
```

## Версия

v1.0.0 (Go 1.22+)
