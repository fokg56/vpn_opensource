# VPN Client - Структура проекта

```
vpn-client/
│
├── 📄 go.mod                      # Go модули и зависимости
├── 📄 go.sum                      # Checksums для зависимостей
│
├── 📄 README.md                   # Основная документация (что и как использовать)
├── 📄 QUICKSTART.md               # Быстрый старт для новичков
├── 📄 INSTALL.md                  # Подробные инструкции установки
├── 📄 ARCHITECTURE.md             # Архитектурная документация
├── 📄 PROJECT_STRUCTURE.md        # Этот файл
│
├── 📄 LICENSE                     # MIT лицензия
├── 📄 .gitignore                  # Git ignore правила
│
├── 📜 Dockerfile                  # Docker образ для контейнеризации
├── 📜 docker-compose.yml          # Docker Compose конфигурация
├── 📜 Makefile                    # Make команды для разработки
│
├── 🏗️  build.sh                   # Скрипт сборки для Linux/macOS
├── 🏗️  build.bat                  # Скрипт сборки для Windows
│
├── 📋 xray-server-config.json    # Пример конфигурации Xray сервера
│
├── 📁 cmd/                        # [ТОЧКА ВХОДА]
│   └── main.go                    # Главная программа
│       ├── Парсинг флагов (-addr)
│       ├── Создание HTTP сервера
│       ├── Обработка сигналов (SIGINT, SIGTERM)
│       └── Основной loop
│
├── 📁 internal/                   # [ВНУТРЕННИЕ МОДУЛИ]
│   │
│   ├── 📁 crypto/                 # Шифрование и криптография
│   │   ├── chacha20.go            # Реализация ChaCha20
│   │   │   ├── ChaCha20Cipher struct
│   │   │   ├── NewChaCha20Cipher()
│   │   │   ├── Encrypt()
│   │   │   ├── Decrypt()
│   │   │   ├── GenerateNonce()
│   │   │   ├── ValidateKey()
│   │   │   ├── StreamEncrypt()
│   │   │   └── StreamDecrypt()
│   │   │
│   │   └── chacha20_test.go       # Юнит-тесты для ChaCha20
│   │       ├── TestNewChaCha20Cipher
│   │       ├── TestInvalidKeyLength
│   │       ├── TestEncryptDecrypt
│   │       ├── TestGenerateNonce
│   │       ├── TestValidateKey
│   │       ├── BenchmarkEncrypt
│   │       └── BenchmarkDecrypt
│   │
│   ├── 📁 vless/                  # VLESS протокол
│   │   ├── protocol.go            # Реализация VLESS
│   │   │   ├── CommandType const (CmdTCP, CmdUDP)
│   │   │   ├── Handshake struct
│   │   │   ├── RequestHeader struct
│   │   │   ├── ClientConnection struct
│   │   │   ├── NewRequestHeader()
│   │   │   ├── EncodeRequest()
│   │   │   ├── DecodeRequest()
│   │   │   ├── NewClientConnection()
│   │   │   ├── Handshake()
│   │   │   ├── Read()
│   │   │   ├── Write()
│   │   │   ├── Close()
│   │   │   ├── SetDeadline()
│   │   │   ├── SetReadDeadline()
│   │   │   └── SetWriteDeadline()
│   │   │
│   │   └── protocol_test.go       # Тесты VLESS (если нужны)
│   │
│   ├── 📁 client/                 # VPN клиент логика
│   │   ├── tunnelmanager.go       # Менеджер туннеля
│   │   │   ├── TunnelConfig struct
│   │   │   ├── TunnelManager struct
│   │   │   ├── NewTunnelManager()
│   │   │   ├── Start()
│   │   │   ├── Stop()
│   │   │   ├── IsRunning()
│   │   │   ├── GetLogChan()
│   │   │   ├── GetStatusChan()
│   │   │   ├── acceptConnections()
│   │   │   ├── handleConnection()
│   │   │   ├── relayData()
│   │   │   ├── sendLog()
│   │   │   └── sendStatus()
│   │   │
│   │   └── tunnelmanager_test.go  # Тесты TunnelManager (опционально)
│   │
│   └── 📁 web/                    # Веб-серв и HTTP обработчики
│       ├── server.go              # HTTP сервер и WebSocket
│       │   ├── Server struct
│       │   ├── ConnectRequest struct
│       │   ├── NewServer()
│       │   ├── Start()
│       │   ├── handleIndex()
│       │   ├── handleConnect()
│       │   ├── handleDisconnect()
│       │   ├── handleStatus()
│       │   ├── handleWebSocket()
│       │   ├── broadcastLogs()
│       │   ├── broadcastStatus()
│       │   ├── broadcastMessage()
│       │   └── indexHTML const (HTML/CSS/JS)
│       │
│       └── server_test.go         # Тесты сервера (опционально)
│
├── 📁 web/                        # [СТАТИЧЕСКИЕ ФАЙЛЫ]
│   └── static/                    # (Зарезервировано для расширения)
│       ├── index.html             # (Встроена в server.go)
│       ├── style.css              # (Встроена в server.go)
│       └── script.js              # (Встроена в server.go)
│
└── 📁 docs/                       # [ДОКУМЕНТАЦИЯ] (опционально)
    ├── API.md
    ├── DEVELOPMENT.md
    └── TROUBLESHOOTING.md
```

## Описание модулей

### `/cmd/main.go` - Главное приложение
- **Размер**: ~30 строк
- **Назначение**: Точка входа приложения
- **Зависимости**: `internal/web`
- **Функции**:
  - Парсинг флагов командной строки (`-addr`)
  - Инициализация и запуск веб-сервера
  - Обработка сигналов корректного завершения

### `/internal/crypto/chacha20.go` - Криптография
- **Размер**: ~200 строк
- **Назначение**: Шифрование/дешифрование трафика
- **Зависимости**: `golang.org/x/crypto/chacha20`
- **Функции**:
  - ChaCha20 шифрование потока данных
  - Генерация безопасного nonce
  - Валидация ключей
  - Потоковое шифрование/дешифрование

### `/internal/vless/protocol.go` - VLESS протокол
- **Размер**: ~250 строк
- **Назначение**: Реализация VLESS протокола для соединения с Xray
- **Зависимости**: стандартная `net` библиотека
- **Функции**:
  - Кодирование/декодирование VLESS пакетов
  - Управление соединением с сервером
  - Рукопожатие с сервером

### `/internal/client/tunnelmanager.go` - Менеджер туннеля
- **Размер**: ~250 строк
- **Назначение**: Управление VPN туннелем, relay трафика
- **Зависимости**: `internal/crypto`, `internal/vless`
- **Функции**:
  - Создание локального TCP слушателя
  - Bidirectional релей между локальным и удаленным соединением
  - Шифрование/дешифрование данных
  - Логирование и управление статусом

### `/internal/web/server.go` - Веб-сервер
- **Размер**: ~500 строк
- **Назначение**: HTTP сервер, REST API, WebSocket, HTML интерфейс
- **Зависимости**: `net/http`, `github.com/gorilla/websocket`, `internal/client`
- **Функции**:
  - HTTP endpoints для управления туннелем
  - WebSocket для real-time логов и статуса
  - Встроенный HTML/CSS/JS интерфейс
  - JSON API

## Файловые зависимости

```
cmd/main.go
    ↓
internal/web/server.go
    ├── internal/client/tunnelmanager.go
    │   ├── internal/crypto/chacha20.go
    │   │   └── golang.org/x/crypto/chacha20
    │   └── internal/vless/protocol.go
    │       └── net (std)
    │
    ├── github.com/gorilla/websocket
    ├── github.com/google/uuid
    └── net/http (std)
```

## Общая статистика

| Метрика | Значение |
|---------|----------|
| **Файлов Go кода** | 5 основных + 1 тест |
| **Строк кода** | ~1200-1500 |
| **Главных посттрок** | 6 (main, crypto, vless, client, web, test) |
| **Зависимостей** | 3 внешниx |
| **Модулей/пакетов** | 5 |
| **HTTP endpoints** | 4 + WebSocket |
| **Форматов данных** | JSON, VLESS binary |

## Расширение структуры

### Добавление нового модуля

```
internal/
└── mynewmodule/
    ├── mynewmodule.go        # Основной файл
    ├── mynewmodule_test.go   # Тесты
    └── types.go              # Типы (опционально)
```

### Добавление новых endpoint'ов

Отредактируйте `internal/web/server.go`:
```go
http.HandleFunc("/api/mynewendpoint", s.handleMyNewEndpoint)
```

### Добавление CLI флагов

Отредактируйте `cmd/main.go`:
```go
myFlag := flag.String("myflag", "default", "Description")
```

## Интеграция с CI/CD

### GitHub Actions

```yaml
name: Build and Test
on: [push]
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - uses: actions/setup-go@v2
      - run: go build ./cmd
      - run: go test ./...
```

## Соглашения кодирования

- **Имена пакетов**: lowercase, no underscores
- **Имена типов**: PascalCase
- **Имена функций**: PascalCase (экспортируемые), camelCase (приватные)
- **Комментарии**: На русском языке, начиная с имени функции
- **Обработка ошибок**: Всегда проверяйте и оборачивайте ошибки

## Развертывание

### Локальное

```bash
make build
./ vpn-client
```

### Docker

```bash
docker build -t vpn-client .
docker run -p 8080:8080 vpn-client
```

### Production (Systemd)

```bash
sudo cp vpn-client /usr/local/bin/
sudo systemctl enable vpn-client
sudo systemctl start vpn-client
```

---

**Last Updated**: 2024  
**Version**: 1.0.0
