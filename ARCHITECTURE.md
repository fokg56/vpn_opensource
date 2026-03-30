# VPN Client - Архитектурная документация

## Обзор архитектуры

VPN Client состоит из следующих компонентов:

```
┌─────────────────────────────────────────────────────┐
│                  Web UI (HTML/JS)                    │
│              (Веб-интерфейс браузера)                │
└──────────────────┬──────────────────────────────────┘
                   │ HTTP + WebSocket
                   │
┌──────────────────▼──────────────────────────────────┐
│            HTTP Server (web.Server)                  │
│        (internal/web/server.go)                      │
│  ┌─────────────────────────────────────────────┐    │
│  │ - GET  /          (HTML страница)           │    │
│  │ - POST /api/connect      (подключение)      │    │
│  │ - POST /api/disconnect   (отключение)       │    │
│  │ - GET  /api/status       (статус)           │    │
│  │ - WS   /ws               (WebSocket)        │    │
│  └─────────────────────────────────────────────┘    │
└──────────────────┬──────────────────────────────────┘
                   │
┌──────────────────▼──────────────────────────────────┐
│          Tunnel Manager (client.TunnelManager)       │
│        (internal/client/tunnelmanager.go)            │
│  ┌─────────────────────────────────────────────┐    │
│  │ - Управление VLESS соединением              │    │
│  │ - Проксирование TCP трафика                 │    │
│  │ - Шифрование/дешифрование ChaCha20          │    │
│  │ - Relay между локальным и VLESS сокетом     │    │
│  │ - Логирование и статус                      │    │
│  └─────────────────────────────────────────────┘    │
└──────────────────┬──────────────────────────────────┘
                   │
        ┌──────────┴──────────┐
        │                     │
┌───────▼───────┐      ┌──────▼─────────┐
│   VLESS       │      │  ChaCha20      │
│  Connection   │      │  Cipher        │
│  (vless/      │      │  (crypto/      │
│   protocol.go)│      │   chacha20.go) │
│  ┌──────────┐ │      │  ┌──────────┐  │
│  │ -Connect │ │      │  │ -Encrypt │  │
│  │ -Handshake│ │      │  │ -Decrypt │  │
│  │ -Read    │ │      │  │ -Nonce   │  │
│  │ -Write   │ │      │  │ -Validate│  │
│  └──────────┘ │      │  └──────────┘  │
└───────┬───────┘      └──────┬─────────┘
        │                     │
        │ VLESS Protocol      │ Stream Encryption
        │ (TCP)               │ (ChaCha20-Poly1305)
        │                     │
        └──────────┬──────────┘
                   │
         ┌─────────▼──────────┐
         │  Xray Server       │
         │  (VPN Gateway)     │
         └────────────────────┘
```

## Модули проекта

### 1. `/cmd/main.go` - Точка входа

Отвечает за:
- Парсинг флагов командной строки
- Создание и запуск HTTP сервера
- Обработку сигналов завершения

```go
./vpn-client -addr 0.0.0.0:8080
```

### 2. `/internal/crypto/chacha20.go` - Шифрование

**Структура**: `ChaCha20Cipher`

Функции:
- `NewChaCha20Cipher(key)` - Создание шифра
- `Encrypt(plaintext)` - Шифрует данные
- `Decrypt(ciphertext)` - Дешифрует данные
- `GenerateNonce()` - Генерирует безопасный nonce
- `ValidateKey(key)` - Проверяет корректность ключа
- `StreamEncrypt(src, dst)` - Потоковое шифрование
- `StreamDecrypt(src, dst)` - Потоковое дешифрование

**Детали**:
- Использует `golang.org/x/crypto/chacha20`
- Ключ: 32 байта (256 бит)
- Nonce: 12 байт (96 бит)
- Формат зашифрованных данных: [12 bytes nonce] + [N bytes ciphertext]

### 3. `/internal/vless/protocol.go` - Протокол VLESS

**Структуры**:
- `RequestHeader` - Заголовок VLESS запроса
- `ClientConnection` - Соединение с VLESS сервером

**Функции**:
- `EncodeRequest(uuid, address, port)` - Кодирует запрос
- `DecodeRequest(data)` - Декодирует запрос
- `NewClientConnection(addr, port, uuid)` - Создает соединение
- `Handshake(address, destPort)` - Рукопожатие с сервером

**Формат VLESS запроса**:
```
[1 byte: version]
[1 byte: cmd (TCP=1, UDP=2)]
[16 bytes: UUID]
[1 byte: addr type (1=IPv4, 2=domain, 3=IPv6)]
[1+ bytes: address]
[2 bytes: port]
```

### 4. `/internal/client/tunnelmanager.go` - Управление туннелем

**Структура**: `TunnelManager`

Функции:
- `NewTunnelManager(config)` - Создание менеджера
- `Start()` - Запуск туннеля
- `Stop()` - Остановка туннеля
- `IsRunning()` - Проверка статуса
- `GetLogChan()` - Канал логов
- `GetStatusChan()` - Канал статусов

**Логика работы**:
1. Создает локальный TCP слушатель на `127.0.0.1:LocalPort`
2. Для каждого входящего соединения:
   - Создает VLESS соединение с сервером
   - Выполняет рукопожатие
   - Запускает relay для передачи данных в обоих направлениях
3. Шифрует данные ChaCha20 перед передачей

### 5. `/internal/web/server.go` - Веб-сервер

**Структура**: `Server`

**HTTP Endpoints**:
- `GET /` - HTML интерфейс
- `POST /api/connect` - Подключение
- `POST /api/disconnect` - Отключение
- `GET /api/status` - Статус
- `WS /ws` - WebSocket для логов

**Структура запроса подключения**:
```json
{
  "server_addr": "example.com",
  "server_port": 443,
  "uuid": "00000000-0000-0000-0000-000000000000",
  "key": "32-символьный-ключ-chacha20",
  "local_port": 1080
}
```

**WebSocket сообщения**:
```json
// Логи
{"type": "log", "data": "[HH:MM:SS] Сообщение"}

// Статус
{"type": "status", "data": "connected/disconnected"}
```

## Поток данных

### Подключение клиента (входящие данные)

```
1. Браузер → HTTP POST /api/connect
2. Server.handleConnect() валидирует параметры
3. TunnelManager.Start() создает локальный слушатель
4. Клиент подключается к 127.0.0.1:1080
5. Создается VLESS соединение с серверомhandleConnection():
   - Выполняется рукопожатие VLESS
   - Запускается relay с шифрованием
6. Данные: Client → Local Port → ChaCha20.Encrypt() → VLESS → Server
```

### Отправка данных через туннель

```
Client Data (plaintext)
      ↓
ChaCha20.Encrypt([data])
      ↓ (nonce + encrypted_data)
VLESS Protocol
      ↓
Xray Server
      ↓
Target Application
```

### Получение ответа

```
Target Response (plaintext)
      ↓
Xray Server
      ↓
VLESS Protocol (encrypted)
      ↓
ChaCha20.Decrypt(data)
      ↓
Client Application
```

## Безопасность

### Ключевые компоненты безопасности

1. **ChaCha20 Шифрование**
   - Криптографически стойкий поток шифра
   - Каждый пакет имеет уникальный nonce
   - Размер ключа: 256 бит (32 байта)

2. **VLESS Протокол**
   - Использует UUID для аутентификации
   - Поддерживает TLS для защиты соединения
   - Нет аутентификации на уровне протокола (используется UUID)

3. **Валидация данных**
   - Проверка размера ключа (ровно 32 байта)
   - Валидация UUID формата
   - Проверка в интерпретация адреса

### Рекомендации безопасности

⚠️ **Для production использования**:
- Используйте HTTPS для веб-интерфейса
- Добавьте аутентификацию на веб-сервер
- Используйте TLS 1.3 между клиентом и сервером
- Регулярно обновляйте зависимости
- Используйте криптографически стойкие ключи

## Расширяемость

### Добавление новых протоколов

1. Создайте новый файл в `/internal/vless/`
2. Реализуйте интерфейс соединения:
   ```go
   type Connection interface {
       Read(p []byte) (int, error)
       Write(p []byte) (int, error)
       Close() error
   }
   ```

3. Обновите `TunnelManager` для использования нового протокола

### Добавление новых методов шифрования

1. Создайте новый файл в `/internal/crypto/`
2. Реализуйте интерфейс шифра:
   ```go
   type Cipher interface {
       Encrypt(plaintext []byte) ([]byte, error)
       Decrypt(ciphertext []byte) ([]byte, error)
   }
   ```

3. Обновите `TunnelManager` для использования нового шифра

## Производительность

### Оптимизации

1. **Buffered I/O**: Использование 4KB буферов для relay
2. **Goroutines**: Каждое соединение обрабатывается отдельно
3. **Streaming**: Шифрование работает со потоками данных
4. **Zero-copy**: Минимизация копирования данных

### Профилирование

```bash
# CPU профилирование
go run -cpuprofile=cpu.prof ./cmd

# Memory профилирование
go run -memprofile=mem.prof ./cmd

# Анализ
go tool pprof cpu.prof
```

## Тестирование

### Юнит-тесты

```bash
go test ./...

# С покрытием
go test -cover ./...

# С детальным выводом
go test -v ./...
```

### Бенчмарки

```bash
go test -bench=. ./internal/crypto
```

## Развертывание

### Docker

```bash
docker build -t vpn-client .
docker run -p 8080:8080 vpn-client
```

### Systemd (Linux)

```ini
[Unit]
Description=VPN Client
After=network.target

[Service]
Type=simple
ExecStart=/usr/local/bin/vpn-client -addr 0.0.0.0:8080
Restart=always

[Install]
WantedBy=multi-user.target
```

## Версионирование

- **Версия продукта**: v1.0.0
- **Go версия**: 1.22+
- **Лицензия**: MIT
