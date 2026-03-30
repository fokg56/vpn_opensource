# VPN Client - Troubleshooting & FAQ

## Часто встречающиеся проблемы

### 1. ❌ Port already in use

**Ошибка**: `Address already in use`

**Причины**:
- Порт 8080 уже занят другим процессом
- Предыдущий экземпляр приложения не закрыт

**Решения**:

**Linux/macOS**:
```bash
# Найти процесс на порту
lsof -i :8080

# Убить процесс
kill -9 <PID>

# Или использовать другой порт
./vpn-client -addr 127.0.0.1:8081
```

**Windows**:
```powershell
# Найти процесс
Get-Process | Where-Object {$_.Handles -gt 500}

# Или используйте CMD
netstat -ano | findstr :8080

# Убить процесс
taskkill /PID <PID> /F

# Или использовать другой порт
vpn-client.exe -addr 127.0.0.1:8081
```

---

### 2. ❌ Connection refused

**Ошибка**: `connection refused` при подключении к VPN серверу

**Причины**:
- Неверный адрес или порт сервера
- Сервер не запущен
- Брандмауэр блокирует соединение

**Решения**:

1. **Проверить хост и порт**:
```bash
# Linux/macOS
nc -zv example.com 443
telnet example.com 443

# Windows PowerShell
Test-NetConnection -ComputerName example.com -Port 443
```

2. **Убедиться что сервер запущен**:
```bash
# Если используется Xray
ps aux | grep xray

# Проверить логи
tail -f /var/log/xray/error.log
```

3. **Проверить брандмауэр**:
```bash
# Linux (iptables)
sudo iptables -L -n

# macOS (pfctl)
sudo pfctl -s rules

# Windows (Windows Defender)
Get-NetFirewallRule
```

---

### 3. ❌ Invalid UUID

**Ошибка**: `Некорректный UUID`

**Проблема**: UUID не в правильном формате

**Правильный формат**: `00000000-0000-0000-0000-000000000000`

**Решения**:

```bash
# Генерировать UUID
# Linux/macOS
uuidgen

# Windows PowerShell
[guid]::NewGuid().ToString()

# Online
Посетите https://www.uuidgenerator.net/
```

**Без дефисов - НЕПРАВИЛЬНО**:
```
550e8400e29b41d4a716446655440000  ❌
```

**С дефисами - ПРАВИЛЬНО**:
```
550e8400-e29b-41d4-a716-446655440000  ✅
```

---

### 4. ❌ Invalid key

**Ошибка**: `Ключ должен быть 32 байта`

**Проблема**: Ключ не является 32-символьным или недопустимого типа

**Решения**:

**Генерировать 32-байтовый ключ**:

```bash
# Метод 1: OpenSSL (Рекомендуется)
openssl rand -hex 32

# Метод 2: Python
python3 -c "import os; print(os.urandom(32).hex())"

# Метод 3: Linux /dev/urandom
head -c 32 /dev/urandom | xxd -p

# Метод 4: Windows PowerShell
$bytes = New-Object byte[] 32
[System.Security.Cryptography.RNGCryptoServiceProvider]::Create().GetBytes($bytes)
[System.BitConverter]::ToString($bytes) -replace '-'
```

**Проверить длину ключа**:
```bash
# Должно быть 64 символа (128 символов = 64 байта)
echo -n "4f0d5c5e3a8b7c9f2e1a4b6c8d0f1e2a3c4b5d6e7f8a9b0c1d2e3f4a5b6c7d" | wc -c
# Вывод: 64
```

---

### 5. ❌ Cannot access localhost:8080

**Ошибка**: `ERR_CONNECTION_REFUSED` или `Cannot reach server`

**Причины**:
- Приложение не запущено
- Используется неправильный адрес
- Полет в браузере не на http://localhost:8080

**Решения**:

1. **Убедиться что приложение запущено**:
```bash
# Проверить процесс
ps aux | grep vpn-client
```

2. **Правильный адрес**:
```
✅  http://localhost:8080
✅  http://127.0.0.1:8080
✅  http://[::1]:8080        (IPv6)

❌  https://localhost:8080  (при запуске на http)
❌  localhost:8080           (без http://)
```

3. **Проверить что порт открыт**:
```bash
netstat -tuln | grep 8080  # Linux
netstat -ano | findstr 8080  # Windows
```

---

### 6. ❌ WebSocket не подключается

**Ошибка**: В консоли браузера `WebSocket is closed with code 1006`

**Причины**:
- Приложение перезагружается
- Проблема с CORS
- Неправильный протокол (https → wss вместо ws)

**Решения**:

1. **Проверить консоль браузера** (F12 → Console):
```javascript
// Должны увидеть
"WebSocket подключен"
```

2. **Проверить указан правильный протокол**:
```
✅  ws://localhost:8080/ws     (для http)
✅  wss://localhost:8080/ws    (для https)

❌  http://localhost:8080/ws   (неправильный протокол)
```

3. **Перезагрузить приложение**:
```bash
# Остановить приложение (Ctrl+C)
# Перезапустить с теми же параметрами
./vpn-client -addr 0.0.0.0:8080
```

---

### 7. ❌ Нет логов в браузере

**Проблема**: Страница загружается но логи не появляются

**Причины**:
- JavaScript отключен
- WebSocket не подключается
- Ошибка в коде

**Решения**:

1. **Проверить JavaScript**:
   - Откройте Settings в браузере
   - Убедитесь что JavaScript включен

2. **Проверить консоль браузера** (F12):
   - Перейдите на вкладку "Console"
   - Посмотрите на ошибки
   - Перейдите на вкладку "Network"
   - Проверьте есть ли WebSocket соединение

3. **Обновить страницу**:
```
Ctrl+R (или Cmd+R на Mac)
```

4. **Очистить кеш браузера**:
```
F12 → Application → Clear Site Data
```

---

### 8. ❌ DNS не работает

**Проблема**: Странички загружаются но DNS запросы не проходят

**Причины**:
- DNS прокси не настроен
- Браузер использует другой DNS

**Решения**:

1. **Использовать IP адреса вместо доменов** (временное решение):
```
IP: 1.1.1.1      → Cloudflare
IP: 8.8.8.8      → Google
IP: 208.67.222.222 → OpenDNS
```

2. **Настроить DNS прокси** (например, tinyproxy):
```bash
# Установка
apt-get install tinyproxy  # Linux

# Запуск
tinyproxy -c /etc/tinyproxy/tinyproxy.conf
```

3. **Использовать DoH (DNS over HTTPS)**:
```
Firefox → Settings → Network Settings
Checked "Enable DNS over HTTPS"
```

---

### 9. ❌ Низкая скорость

**Проблема**: Соединение медленное, высокая latency

**Причины**:
- Плохая сеть
- Сервер перегружен
- Проблема с маршрутизацией

**Решения**:

1. **Проверить пинг**:
```bash
ping -c 4 example.com     # Linux/macOS
ping -n 4 example.com     # Windows
```

2. **Проверить трафик туннеля**:
```bash
# Linux - мониторить порт
watch -n 1 'netstat -tuln | grep 1080'
```

3. **Использовать более близкий сервер**:
- Выберите сервер географически ближе
- Может быть лучше пропускная способность

4. **Проверить нагрузку на сервер**:
```bash
# На сервере
top
free -m
```

---

### 10. ❌ VLESS Handshake failed

**Ошибка**: `Ошибка рукопожатия: ...`

**Причины**:
- UUID не совпадает с сервером
- Неправильная конфигурация сервера
- Проблема с соединением

**Решения**:

1. **Проверить UUID**:
   - На клиенте и сервере должны быть одинаковые UUID
   - Проверить в конфигурации сервера

2. **Проверить конфигурацию Xray**:
```bash
# Валидировать JSON конфигурацию
xray test -config /path/to/config.json
```

3. **Включить debug логирование**:
```json
{
  "log": {
    "loglevel": "debug"
  }
}
```

4. **Проверить TLS сертификаты**:
```bash
# Проверить сертификат
openssl x509 -in cert.crt -text -noout
```

---

## FAQ (Часто задаваемые вопросы)

### Q: Какой Go версии мне нужен?

**A**: Go 1.22 или выше. Проверьте version:
```bash
go version
```

---

### Q: Могу ли я использовать это на мобильном телефоне?

**A**: Сейчас нет, это только для ПК с браузером. Можно развернуть на мобильном сервере через терминал приложение типа Termux.

---

### Q: Как я могу использовать прокси в приложении X?

**A**: Зависит от приложения:

- **Firefox**: Settings → Network → SOCKS Host: 127.0.0.1, Port: 1080
- **Chrome**: Требует расширение или использует системные ориентир
- **curl**: `curl -x socks5://127.0.0.1:1080 https://example.com`
- **Python**: `requests.proxies = {'http': 'socks5://127.0.0.1:1080'}`

---

### Q: Безопасно ли использовать этот клиент?

**A**: Код открытый и можно аудировать. Однако:
- ✅ Используется современный шифр (ChaCha20)
- ⚠️ Используйте HTTPS в production
- ⚠️ Добавьте аутентификацию на веб-интерфейс
- ✅ Регулярно обновляйте зависимости

---

### Q: Сколько соединений я могу создать?

**A**: Теоретически не ограничено, но зависит от:
- Пропускной способности сервера
- Ресурсов компьютера (RAM, CPU)
- Лимитов ОС на количество файловых дескрипторов

---

### Q: Могу ли я использовать UDP?

**A**: Сейчас поддерживается только TCP. Для UDP надо:
1. Добавить поддержку в `internal/vless/protocol.go`
2. Создать UDP слушатель в `internal/client/tunnelmanager.go`
3. Обновить веб-интерфейс

---

### Q: Как я могу восстановить пароль/ключ?

**A**: Если вы забыли ключ:
1. Создайте новый ключ: `openssl rand -hex 32`
2. Обновите конфигурацию сервера
3. Используйте новый ключ в клиенте

---

### Q: Где хранится моя конфигурация?

**A**: Конфигурация не сохраняется. Вы должны вводить параметры каждый раз. Для сохранения:
1. Создайте JSON файл с параметрами
2. Модифицируйте веб-интерфейс для загрузки из файла

---

## Логирование и отладка

### Включение DEBUG режима

Отредактируйте `internal/web/server.go`:
```go
log.Printf("DEBUG: %v", variable)
```

### Анализ логов

**С использованием grep**:
```bash
./vpn-client 2>&1 | grep -i "error\|warning"
```

**Сохранение в файл**:
```bash
./vpn-client > vpn-client.log 2>&1
tail -f vpn-client.log
```

### Трассировка сетевых пакетов

**Linux**:
```bash
sudo tcpdump -i lo -n 'port 1080 or port 8080'
```

---

## Производительность и оптимизация

### Проверка использования ресурсов

**Linux/macOS**:
```bash
ps aux | grep vpn-client
top -p <PID>
```

**Windows**:
```powershell
Get-Process vpn-client
```

### Профилирование

```bash
go run -cpuprofile=cpu.prof ./cmd
# ... используйте приложение ...
# Ctrl+C для завершения

go tool pprof cpu.prof
(pprof) top
(pprof) web
```

---

## Получение помощи

### Проверить логи

1. Откройте консоль браузера (F12)
2. Перейдите на вкладку "Console"
3. Посмотрите на ошибки

### Создать issue на GitHub

Предоставьте:
- Версия Go (`go version`)
- ОС и версия
- Полный текст ошибки
- Шаги воспроизведения

### Куда смотреть в коде

Если ошибка при подключении:
- `internal/web/server.go` → `handleConnect()`
- `internal/client/tunnelmanager.go` → `Start()`

Если ошибка при шифровании:
- `internal/crypto/chacha20.go`

Если ошибка с VLESS:
- `internal/vless/protocol.go`

---

## Дополнительные ресурсы

- [Go Documentation](https://golang.org/doc/)
- [Xray Documentation](https://xtls.github.io/)
- [VLESS Protocol](https://github.com/xtls/xray-core/wiki/VLESS)
- [ChaCha20 RFC](https://tools.ietf.org/html/rfc7539)

---

**Last Updated**: 2024  
**Версия**: 1.0.0
