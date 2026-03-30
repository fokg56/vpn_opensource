# VPN Client - REST API Документация

## Базовая информация

- **Базовый URL**: `http://localhost:8080`
- **Версия API**: v1.0
- **Формат**: JSON (для запросов и ответов)
- **WebSocket**: `ws://localhost:8080/ws`

---

## Endpoints

### 1. GET `/`

Возвращает полную веб-страницу с интерфейсом.

**Параметры**: Нет

**Пример запроса**:
```bash
curl http://localhost:8080/
```

**Ответ**: HTML страница (200 OK)

---

### 2. POST `/api/connect`

Инициирует подключение к VPN серверу.

**URL**: `POST /api/connect`

**Headers**:
```
Content-Type: application/json
```

**Body** (JSON):
```json
{
  "server_addr": "string",
  "server_port": "number",
  "uuid": "string (UUID format)",
  "key": "string (32 символа)",
  "local_port": "number"
}
```

**Параметры**:

| Параметр | Тип | Обязателен | Описание |
|----------|-----|-----------|---------|
| `server_addr` | string | Да | IP адрес или доменное имя Xray сервера |
| `server_port` | number | Да | Порт Xray сервера (1-65535) |
| `uuid` | string | Да | UUID клиента в формате 00000000-0000-0000-0000-000000000000 |
| `key` | string | Да | ChaCha20 ключ, минимум 32 символа |
| `local_port` | number | Да | Локальный порт SOCKS5 прокси (1-65535) |

**Пример запроса**:
```bash
curl -X POST http://localhost:8080/api/connect \
  -H "Content-Type: application/json" \
  -d '{
    "server_addr": "example.com",
    "server_port": 443,
    "uuid": "550e8400-e29b-41d4-a716-446655440000",
    "key": "4f0d5c5e3a8b7c9f2e1a4b6c8d0f1e2a3c4b5d6e7f8a9b0c1d2e3f4a5b6c7d",
    "local_port": 1080
  }'
```

**Успешный ответ** (200 OK):
```json
{
  "status": "connected",
  "message": "Подключено к example.com:443"
}
```

**Ошибки**:

| Код | Описание | Причина |
|-----|---------|---------|
| 400 | Bad Request | Некорректный JSON, недостаточно параметров |
| 400 | Invalid UUID | UUID не в формате 00000000-0000-0000-0000-000000000000 |
| 400 | Invalid key | Ключ не 32 символа или содержит недопустимые символы |
| 500 | Internal Server Error | Ошибка при создании или запуске туннеля |

**Пример ошибки** (400):
```json
{
  "error": "Некорректный ключ: ключ должен быть 32 байта, получено: 16"
}
```

---

### 3. POST `/api/disconnect`

Отключается от VPN сервера и останавливает локальный прокси.

**URL**: `POST /api/disconnect`

**Headers**: Нет

**Body**: Нет

**Пример запроса**:
```bash
curl -X POST http://localhost:8080/api/disconnect
```

**Успешный ответ** (200 OK):
```json
{
  "status": "disconnected",
  "message": "Отключено"
}
```

**Ошибки**:

| Код | Описание |
|-----|---------|
| 400 | Туннель не запущен |
| 500 | Ошибка при остановке туннеля |

---

### 4. GET `/api/status`

Получает текущий статус подключения.

**URL**: `GET /api/status`

**Parameters**: Нет

**Пример запроса**:
```bash
curl http://localhost:8080/api/status
```

**Ответ - Подключено** (200 OK):
```json
{
  "status": "connected"
}
```

**Ответ - Отключено** (200 OK):
```json
{
  "status": "disconnected"
}
```

---

### 5. WebSocket `/ws`

WebSocket соединение для получения логов и обновлений статуса в реальном времени.

**URL**: `ws://localhost:8080/ws`

**Protocol**: WebSocket

#### Сообщения от сервера:

**Тип: Log**
```json
{
  "type": "log",
  "data": "[15:04:05] Туннель запущен на 127.0.0.1:1080"
}
```

**Тип: Status**
```json
{
  "type": "status",
  "data": "connected" | "disconnected"
}
```

#### Примеры логов:

```
[15:04:05] Туннель запущен на 127.0.0.1:1080
[15:04:06] Новое соединение: 127.0.0.1:54321
[15:04:06] Рукопожатие успешно
[15:04:10] Ошибка при чтении: connection reset by peer
```

#### Пример клиента на JavaScript:

```javascript
const ws = new WebSocket('ws://localhost:8080/ws');

ws.onopen = () => {
  console.log('WebSocket подключен');
};

ws.onmessage = (event) => {
  const msg = JSON.parse(event.data);
  if (msg.type === 'log') {
    console.log('Log:', msg.data);
  } else if (msg.type === 'status') {
    console.log('Status:', msg.data);
  }
};

ws.onerror = (error) => {
  console.error('WebSocket error:', error);
};

ws.onclose = () => {
  console.log('WebSocket закрыт');
};
```

#### Пример клиента на Python:

```python
import websocket
import json

def on_open(ws):
    print("WebSocket подключен")

def on_message(ws, message):
    data = json.loads(message)
    if data['type'] == 'log':
        print(f"Log: {data['data']}")
    elif data['type'] == 'status':
        print(f"Status: {data['data']}")

def on_error(ws, error):
    print(f"Error: {error}")

def on_close(ws):
    print("WebSocket закрыт")

ws = websocket.WebSocketApp(
    "ws://localhost:8080/ws",
    on_open=on_open,
    on_message=on_message,
    on_error=on_error,
    on_close=on_close
)

ws.run_forever()
```

---

## Коды статусов HTTP

| Код | Значение | Описание |
|-----|----------|---------|
| 200 | OK | Запрос выполен успешно |
| 400 | Bad Request | Ошибка в параметрах запроса |
| 404 | Not Found | Endpoint не найден |
| 405 | Method Not Allowed | HTTP метод не поддерживается |
| 500 | Internal Server Error | Ошибка сервера |

---

## Примеры использования

### curl

**Подключение**:
```bash
curl -X POST http://localhost:8080/api/connect \
  -H "Content-Type: application/json" \
  -d '{
    "server_addr": "vpn.example.com",
    "server_port": 443,
    "uuid": "550e8400-e29b-41d4-a716-446655440000",
    "key": "4f0d5c5e3a8b7c9f2e1a4b6c8d0f1e2a3c4b5d6e7f8a9b0c1d2e3f4a5b6c7d",
    "local_port": 1080
  }'
```

**Проверка статуса**:
```bash
curl http://localhost:8080/api/status
```

**Отключение**:
```bash
curl -X POST http://localhost:8080/api/disconnect
```

### Python (requests)

```python
import requests
import json

BASE_URL = "http://localhost:8080"

# Подключение
response = requests.post(
    f"{BASE_URL}/api/connect",
    json={
        "server_addr": "vpn.example.com",
        "server_port": 443,
        "uuid": "550e8400-e29b-41d4-a716-446655440000",
        "key": "4f0d5c5e3a8b7c9f2e1a4b6c8d0f1e2a3c4b5d6e7f8a9b0c1d2e3f4a5b6c7d",
        "local_port": 1080
    }
)
print("Connect:", response.json())

# Проверка статуса
response = requests.get(f"{BASE_URL}/api/status")
print("Status:", response.json())

# Отключение
response = requests.post(f"{BASE_URL}/api/disconnect")
print("Disconnect:", response.json())
```

### JavaScript (fetch)

```javascript
const BASE_URL = "http://localhost:8080";

// Подключение
fetch(`${BASE_URL}/api/connect`, {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    server_addr: "vpn.example.com",
    server_port: 443,
    uuid: "550e8400-e29b-41d4-a716-446655440000",
    key: "4f0d5c5e3a8b7c9f2e1a4b6c8d0f1e2a3c4b5d6e7f8a9b0c1d2e3f4a5b6c7d",
    local_port: 1080
  })
})
.then(r => r.json())
.then(data => console.log("Connected:", data));

// Проверка статуса
fetch(`${BASE_URL}/api/status`)
  .then(r => r.json())
  .then(data => console.log("Status:", data));

// Отключение
fetch(`${BASE_URL}/api/disconnect`, { method: 'POST' })
  .then(r => r.json())
  .then(data => console.log("Disconnected:", data));
```

---

## Ограничения и особенности

### Ограничения

- **Один активный туннель**: Одновременно может работать только одно подключение
- **Локальный слушатель**: Прокси доступен только локально (127.0.0.1)
- **Размер ключа**: Ровно 32 символа для ChaCha20
- **UUID формат**: Должен быть в формате 00000000-0000-0000-0000-000000000000

### Особенности

- **WebSocket auto-reconnect**: Браузер автоматически переподключается при разрыве
- **Логи в памяти**: Логи хранятся в канале, при переполнении игнорируются новые
- **Graceful shutdown**: При остановке приложения соединения корректно закрываются
- **Concurrent connections**: Каждый клиент обрабатывается отдельной goroutine

---

## Best Practices

### Безопасность

✅ Всегда используйте HTTPS в production  
✅ Добавьте аутентификацию на веб-интерфейс  
✅ Используйте крепкие ключи (не 12345678...)  
✅ Валидируйте все входные данные на клиенте  
✅ Используйте TLS 1.3 для соединения с сервером  

### Производительность

✅ Переиспользуйте HTTP соединения  
✅ Закрывайте WebSocket при завершении  
✅ Не отправляйте огромное количество запросов подряд  

### Надежность

✅ Обрабатывайте ошибки сети при подключении  
✅ Реализуйте retry логику для reconnect  
✅ Мониторьте логи для выявления проблем  
✅ Используйте health checks через `/api/status`  

---

## Версия

**Current Version**: 1.0.0  
**Last Updated**: 2024  
**Compatibility**: Go 1.22+
