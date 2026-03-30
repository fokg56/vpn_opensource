package web

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/user/vpn-client/internal/client"
	"github.com/user/vpn-client/internal/crypto"
)

// Server содержит состояние веб-сервера
type Server struct {
	addr          string
	tunnelManager *client.TunnelManager
	upgrader      websocket.Upgrader
	mutex         sync.RWMutex
	wsClients     map[*websocket.Conn]bool
}

// NewServer создает новый веб-сервер
func NewServer(addr string) *Server {
	return &Server{
		addr:      addr,
		upgrader:  websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }},
		wsClients: make(map[*websocket.Conn]bool),
	}
}

// Start запускает веб-сервер
func (s *Server) Start() error {
	http.HandleFunc("/", s.handleIndex)
	http.HandleFunc("/api/connect", s.handleConnect)
	http.HandleFunc("/api/disconnect", s.handleDisconnect)
	http.HandleFunc("/api/status", s.handleStatus)
	http.HandleFunc("/ws", s.handleWebSocket)

	log.Printf("Веб-сервер запущен на http://%s", s.addr)
	return http.ListenAndServe(s.addr, nil)
}

// handleIndex обслуживает главную страницу
func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(indexHTML))
}

// ConnectRequest структура запроса на подключение
type ConnectRequest struct {
	ServerAddr string `json:"server_addr"`
	ServerPort uint16 `json:"server_port"`
	UUID       string `json:"uuid"`
	Key        string `json:"key"`
	LocalPort  uint16 `json:"local_port"`
}

// handleConnect обрабатывает запрос подключения
func (s *Server) handleConnect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не допускается", http.StatusMethodNotAllowed)
		return
	}

	var req ConnectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Ошибка декодирования: %v", err), http.StatusBadRequest)
		return
	}

	// Проверяем параметры
	if req.ServerAddr == "" {
		http.Error(w, "Адрес сервера не может быть пустым", http.StatusBadRequest)
		return
	}

	if req.ServerPort == 0 {
		http.Error(w, "Порт сервера должен быть больше 0", http.StatusBadRequest)
		return
	}

	// Парсим UUID
	parsedUUID, err := uuid.Parse(req.UUID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Некорректный UUID: %v", err), http.StatusBadRequest)
		return
	}

	// Проверяем ключ
	keyBytes := []byte(req.Key)
	if err := crypto.ValidateKey(keyBytes); err != nil {
		http.Error(w, fmt.Sprintf("Некорректный ключ: %v", err), http.StatusBadRequest)
		return
	}

	// Создаем конфигурацию туннеля
	config := &client.TunnelConfig{
		ServerAddr:  req.ServerAddr,
		ServerPort:  req.ServerPort,
		UUID:        [16]byte(parsedUUID),
		ChaCha20Key: keyBytes,
		LocalPort:   req.LocalPort,
	}

	// Создаем менеджер туннеля
	tu, err := client.NewTunnelManager(config)
	if err != nil {
		http.Error(w, fmt.Sprintf("Ошибка создания туннеля: %v", err), http.StatusInternalServerError)
		return
	}

	// Запускаем туннель
	if err := tu.Start(); err != nil {
		http.Error(w, fmt.Sprintf("Ошибка запуска туннеля: %v", err), http.StatusInternalServerError)
		return
	}

	s.mutex.Lock()
	s.tunnelManager = tu
	s.mutex.Unlock()

	// Запускаем goroutine для отправки логов и статусов в WebSocket
	go s.broadcastStatus(tu)
	go s.broadcastLogs(tu)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "connected",
		"message": fmt.Sprintf("Подключено к %s:%d", req.ServerAddr, req.ServerPort),
	})
}

// handleDisconnect обрабатывает запрос отключения
func (s *Server) handleDisconnect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Метод не допускается", http.StatusMethodNotAllowed)
		return
	}

	s.mutex.Lock()
	tm := s.tunnelManager
	s.tunnelManager = nil
	s.mutex.Unlock()

	if tm == nil {
		http.Error(w, "Туннель не запущен", http.StatusBadRequest)
		return
	}

	if err := tm.Stop(); err != nil {
		http.Error(w, fmt.Sprintf("Ошибка остановки туннеля: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "disconnected",
		"message": "Отключено",
	})
}

// handleStatus возвращает статус туннеля
func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	s.mutex.RLock()
	tm := s.tunnelManager
	s.mutex.RUnlock()

	w.Header().Set("Content-Type", "application/json")

	if tm == nil || !tm.IsRunning() {
		json.NewEncoder(w).Encode(map[string]string{
			"status": "disconnected",
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"status": "connected",
	})
}

// handleWebSocket обрабатывает WebSocket соединения
func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Ошибка WebSocket upgrade: %v", err)
		return
	}
	defer conn.Close()

	s.mutex.Lock()
	s.wsClients[conn] = true
	s.mutex.Unlock()

	defer func() {
		s.mutex.Lock()
		delete(s.wsClients, conn)
		s.mutex.Unlock()
	}()

	// Читаем сообщения в цикле (для поддержки keep-alive)
	for {
		var msg map[string]interface{}
		if err := conn.ReadJSON(&msg); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Ошибка WebSocket: %v", err)
			}
			break
		}
	}
}

// broadcastLogs отправляет логи через WebSocket
func (s *Server) broadcastLogs(tm *client.TunnelManager) {
	for log := range tm.GetLogChan() {
		msg := map[string]interface{}{
			"type": "log",
			"data": log,
		}
		s.broadcastMessage(msg)
	}
}

// broadcastStatus отправляет статус через WebSocket
func (s *Server) broadcastStatus(tm *client.TunnelManager) {
	for status := range tm.GetStatusChan() {
		msg := map[string]interface{}{
			"type": "status",
			"data": status,
		}
		s.broadcastMessage(msg)
	}
}

// broadcastMessage отправляет сообщение всем WebSocket клиентам
func (s *Server) broadcastMessage(msg map[string]interface{}) {
	s.mutex.RLock()
	clients := make([]*websocket.Conn, 0, len(s.wsClients))
	for conn := range s.wsClients {
		clients = append(clients, conn)
	}
	s.mutex.RUnlock()

	for _, conn := range clients {
		if err := conn.WriteJSON(msg); err != nil {
			log.Printf("Ошибка отправки WebSocket сообщения: %v", err)
			s.mutex.Lock()
			delete(s.wsClients, conn)
			s.mutex.Unlock()
			conn.Close()
		}
	}
}

// indexHTML содержит HTML страницу
const indexHTML = `<!DOCTYPE html>
<html lang="ru">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>VPN Client</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }

        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            display: flex;
            justify-content: center;
            align-items: center;
            padding: 20px;
        }

        .container {
            background: white;
            border-radius: 10px;
            box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
            max-width: 500px;
            width: 100%;
            padding: 40px;
        }

        h1 {
            text-align: center;
            color: #333;
            margin-bottom: 30px;
            font-size: 28px;
        }

        .form-group {
            margin-bottom: 20px;
        }

        label {
            display: block;
            margin-bottom: 8px;
            color: #555;
            font-weight: 500;
            font-size: 14px;
        }

        input[type="text"],
        input[type="number"],
        input[type="password"] {
            width: 100%;
            padding: 12px;
            border: 2px solid #e0e0e0;
            border-radius: 5px;
            font-size: 14px;
            transition: border-color 0.3s;
        }

        input[type="text"]:focus,
        input[type="number"]:focus,
        input[type="password"]:focus {
            outline: none;
            border-color: #667eea;
        }

        .button-group {
            display: flex;
            gap: 10px;
            margin: 30px 0;
        }

        button {
            flex: 1;
            padding: 12px;
            border: none;
            border-radius: 5px;
            font-size: 16px;
            font-weight: 600;
            cursor: pointer;
            transition: all 0.3s;
        }

        .btn-connect {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
        }

        .btn-connect:hover {
            transform: translateY(-2px);
            box-shadow: 0 10px 20px rgba(102, 126, 234, 0.3);
        }

        .btn-disconnect {
            background: #f44336;
            color: white;
        }

        .btn-disconnect:hover {
            background: #da190b;
        }

        .btn-disconnect:disabled {
            background: #ccc;
            cursor: not-allowed;
        }

        .status {
            padding: 15px;
            border-radius: 5px;
            margin: 20px 0;
            text-align: center;
            font-weight: 600;
        }

        .status.connected {
            background: #d4edda;
            color: #155724;
            border: 1px solid #c3e6cb;
        }

        .status.disconnected {
            background: #f8d7da;
            color: #721c24;
            border: 1px solid #f5c6cb;
        }

        .status.connecting {
            background: #fff3cd;
            color: #856404;
            border: 1px solid #ffeeba;
        }

        .logs {
            background: #f5f5f5;
            border: 1px solid #e0e0e0;
            border-radius: 5px;
            padding: 15px;
            height: 200px;
            overflow-y: auto;
            font-family: 'Courier New', monospace;
            font-size: 12px;
            color: #333;
            margin: 20px 0;
        }

        .log-entry {
            padding: 5px 0;
            border-bottom: 1px solid #e8e8e8;
        }

        .log-entry:last-child {
            border-bottom: none;
        }

        .info-text {
            font-size: 12px;
            color: #999;
            margin-top: 10px;
            text-align: center;
        }

        .spinner {
            display: inline-block;
            width: 10px;
            height: 10px;
            border: 2px solid #f3f3f3;
            border-top: 2px solid #667eea;
            border-radius: 50%;
            animation: spin 1s linear infinite;
            margin-right: 5px;
        }

        @keyframes spin {
            0% { transform: rotate(0deg); }
            100% { transform: rotate(360deg); }
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>🔒 VPN Client</h1>

        <div class="status disconnected" id="status">Отключено</div>

        <form id="vpnForm">
            <div class="form-group">
                <label for="serverAddr">Адрес сервера</label>
                <input type="text" id="serverAddr" placeholder="example.com или 192.168.1.1" required>
            </div>

            <div class="form-group">
                <label for="serverPort">Порт</label>
                <input type="number" id="serverPort" placeholder="443" value="443" required>
            </div>

            <div class="form-group">
                <label for="uuid">UUID</label>
                <input type="text" id="uuid" placeholder="00000000-0000-0000-0000-000000000000" required>
            </div>

            <div class="form-group">
                <label for="key">ChaCha20 Ключ (32 байта, hex или текст)</label>
                <input type="password" id="key" placeholder="Минимум 32 символа" required>
            </div>

            <div class="form-group">
                <label for="localPort">Локальный порт</label>
                <input type="number" id="localPort" placeholder="1080" value="1080" required>
            </div>

            <div class="button-group">
                <button type="submit" class="btn-connect" id="connectBtn">Подключиться</button>
                <button type="button" class="btn-disconnect" id="disconnectBtn" disabled>Отключиться</button>
            </div>
        </form>

        <div class="logs" id="logs"></div>
        <div class="info-text">Логи подключения отображаются выше</div>
    </div>

    <script>
        const form = document.getElementById('vpnForm');
        const status = document.getElementById('status');
        const logs = document.getElementById('logs');
        const connectBtn = document.getElementById('connectBtn');
        const disconnectBtn = document.getElementById('disconnectBtn');
        let ws = null;
        let connected = false;

        // Устанавливаем WebSocket соединение
        function connectWebSocket() {
            const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
            ws = new WebSocket(protocol + '//' + window.location.host + '/ws');

            ws.onopen = () => {
                console.log('WebSocket подключен');
            };

            ws.onmessage = (event) => {
                const message = JSON.parse(event.data);
                if (message.type === 'log') {
                    addLog(message.data);
                } else if (message.type === 'status') {
                    updateStatus(message.data);
                }
            };

            ws.onerror = (error) => {
                console.error('WebSocket ошибка:', error);
            };

            ws.onclose = () => {
                console.log('WebSocket закрыт, переподключение через 3 сек...');
                setTimeout(connectWebSocket, 3000);
            };
        }

        function addLog(message) {
            const entry = document.createElement('div');
            entry.className = 'log-entry';
            entry.textContent = message;
            logs.appendChild(entry);
            logs.scrollTop = logs.scrollHeight;
        }

        function updateStatus(statusStr) {
            if (statusStr === 'connected') {
                status.textContent = '✅ Подключено';
                status.className = 'status connected';
                connected = true;
                connectBtn.disabled = true;
                disconnectBtn.disabled = false;
            } else if (statusStr === 'disconnected') {
                status.textContent = '❌ Отключено';
                status.className = 'status disconnected';
                connected = false;
                connectBtn.disabled = false;
                disconnectBtn.disabled = true;
            }
        }

        form.addEventListener('submit', async (e) => {
            e.preventDefault();

            const serverAddr = document.getElementById('serverAddr').value;
            const serverPort = parseInt(document.getElementById('serverPort').value);
            const uuid = document.getElementById('uuid').value;
            let keyValue = document.getElementById('key').value;
            const localPort = parseInt(document.getElementById('localPort').value);

            // Если ключ текстовый, дополняем до 32 символов
            if (keyValue.length < 32) {
                keyValue = keyValue.padEnd(32, '0');
            }

            status.textContent = '⏳ Подключение...';
            status.className = 'status connecting';

            try {
                const response = await fetch('/api/connect', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({
                        server_addr: serverAddr,
                        server_port: serverPort,
                        uuid: uuid,
                        key: keyValue,
                        local_port: localPort
                    })
                });

                if (!response.ok) {
                    const error = await response.text();
                    addLog('Ошибка: ' + error);
                    updateStatus('disconnected');
                }
            } catch (error) {
                addLog('Ошибка подключения: ' + error.message);
                updateStatus('disconnected');
            }
        });

        disconnectBtn.addEventListener('click', async () => {
            try {
                const response = await fetch('/api/disconnect', { method: 'POST' });
                if (!response.ok) {
                    const error = await response.text();
                    addLog('Ошибка: ' + error);
                }
            } catch (error) {
                addLog('Ошибка отключения: ' + error.message);
            }
        });

        // Проверяем статус при загрузке
        async function checkStatus() {
            try {
                const response = await fetch('/api/status');
                const data = await response.json();
                updateStatus(data.status);
            } catch (error) {
                console.error('Ошибка при проверке статуса:', error);
            }
        }

        connectWebSocket();
        checkStatus();
        // Проверяем статус каждые 5 секунд
        setInterval(checkStatus, 5000);
    </script>
</body>
</html>`
