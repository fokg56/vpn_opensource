package client

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/user/vpn-client/internal/crypto"
	"github.com/user/vpn-client/internal/vless"
)

// TunnelConfig конфигурация туннеля
type TunnelConfig struct {
	ServerAddr  string
	ServerPort  uint16
	UUID        [16]byte
	ChaCha20Key []byte
	LocalPort   uint16
}

// TunnelManager управляет VPN туннелем
type TunnelManager struct {
	config     *TunnelConfig
	cipher     *crypto.ChaCha20Cipher
	listener   net.Listener
	isRunning  bool
	mutex      sync.RWMutex
	logChan    chan string
	stopChan   chan bool
	statusChan chan string
}

// NewTunnelManager создает новый менеджер туннеля
func NewTunnelManager(config *TunnelConfig) (*TunnelManager, error) {
	// Проверяем ключ
	if err := crypto.ValidateKey(config.ChaCha20Key); err != nil {
		return nil, fmt.Errorf("некорректный ключ ChaCha20: %w", err)
	}

	// Создаем шифр
	cipher, err := crypto.NewChaCha20Cipher(config.ChaCha20Key)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания шифра: %w", err)
	}

	tm := &TunnelManager{
		config:     config,
		cipher:     cipher,
		isRunning:  false,
		logChan:    make(chan string, 100),
		stopChan:   make(chan bool),
		statusChan: make(chan string, 10),
	}

	return tm, nil
}

// Start запускает туннель
func (tm *TunnelManager) Start() error {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	if tm.isRunning {
		return fmt.Errorf("туннель уже запущен")
	}

	// Создаем локальный слушатель
	listenAddr := fmt.Sprintf("127.0.0.1:%d", tm.config.LocalPort)
	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		tm.sendLog(fmt.Sprintf("Ошибка создания слушателя: %v", err))
		return fmt.Errorf("ошибка создания слушателя на %s: %w", listenAddr, err)
	}

	tm.listener = listener
	tm.isRunning = true
	tm.sendLog(fmt.Sprintf("Туннель запущен на %s", listenAddr))
	tm.sendStatus("connected")

	// Запускаем goroutine для приема соединений
	go tm.acceptConnections()

	return nil
}

// acceptConnections принимает входящие соединения
func (tm *TunnelManager) acceptConnections() {
	for {
		select {
		case <-tm.stopChan:
			tm.sendLog("Остановка приема соединений")
			return
		default:
		}

		// Устанавливаем таймаут для Accept
		tm.listener.(*net.TCPListener).SetDeadline(time.Now().Add(time.Second))

		conn, err := tm.listener.Accept()
		if err != nil {
			if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
				continue
			}
			if tm.isRunning {
				tm.sendLog(fmt.Sprintf("Ошибка при приеме соединения: %v", err))
			}
			continue
		}

		tm.sendLog(fmt.Sprintf("Новое соединение: %s", conn.RemoteAddr()))

		// Обрабатываем соединение в отдельной goroutine
		go tm.handleConnection(conn)
	}
}

// handleConnection обрабатывает входящее соединение
func (tm *TunnelManager) handleConnection(localConn net.Conn) {
	defer localConn.Close()

	// Создаем VLESS соединение с сервером
	vlessConn, err := vless.NewClientConnection(tm.config.ServerAddr, tm.config.ServerPort, tm.config.UUID)
	if err != nil {
		tm.sendLog(fmt.Sprintf("Ошибка подключения к VPN серверу: %v", err))
		return
	}
	defer vlessConn.Close()

	// Выполняем рукопожатие
	// Для простоты, используем localhost и случайный порт
	if err := vlessConn.Handshake("127.0.0.1", 443); err != nil {
		tm.sendLog(fmt.Sprintf("Ошибка рукопожатия: %v", err))
		return
	}

	tm.sendLog("Рукопожатие успешно")

	// Создаем bidirectional relay между localConn и vlessConn
	go tm.relayData(localConn, vlessConn, true)
	tm.relayData(vlessConn, localConn, false)
}

// relayData передает данные между двумя соединениями с шифрованием
func (tm *TunnelManager) relayData(src net.Conn, dst net.Conn, encrypt bool) {
	defer src.Close()
	defer dst.Close()

	buffer := make([]byte, 4096)
	for {
		n, err := src.Read(buffer)
		if err != nil {
			if err != io.EOF {
				tm.sendLog(fmt.Sprintf("Ошибка чтения: %v", err))
			}
			return
		}

		if n > 0 {
			var data []byte
			var writeErr error

			if encrypt {
				// Шифруем данные
				data, writeErr = tm.cipher.Encrypt(buffer[:n])
				if writeErr != nil {
					tm.sendLog(fmt.Sprintf("Ошибка шифрования: %v", writeErr))
					return
				}
			} else {
				// Дешифруем данные
				data, writeErr = tm.cipher.Decrypt(buffer[:n])
				if writeErr != nil {
					tm.sendLog(fmt.Sprintf("Ошибка дешифрования: %v", writeErr))
					return
				}
			}

			// Записываем в целевое соединение
			if _, err := dst.Write(data); err != nil {
				tm.sendLog(fmt.Sprintf("Ошибка записи: %v", err))
				return
			}
		}
	}
}

// Stop останавливает туннель
func (tm *TunnelManager) Stop() error {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	if !tm.isRunning {
		return fmt.Errorf("туннель не запущен")
	}

	tm.isRunning = false
	tm.sendLog("Остановка туннеля...")
	tm.sendStatus("disconnected")

	if tm.listener != nil {
		tm.listener.Close()
	}

	tm.stopChan <- true

	tm.sendLog("Туннель остановлен")
	return nil
}

// IsRunning возвращает статус туннеля
func (tm *TunnelManager) IsRunning() bool {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()
	return tm.isRunning
}

// GetLogChan возвращает канал логов
func (tm *TunnelManager) GetLogChan() <-chan string {
	return tm.logChan
}

// GetStatusChan возвращает канал статусов
func (tm *TunnelManager) GetStatusChan() <-chan string {
	return tm.statusChan
}

// sendLog отправляет логическое сообщение
func (tm *TunnelManager) sendLog(msg string) {
	select {
	case tm.logChan <- fmt.Sprintf("[%s] %s", time.Now().Format("15:04:05"), msg):
	default:
		// Если канал переполнен, игнорируем
	}
}

// sendStatus отправляет статус
func (tm *TunnelManager) sendStatus(status string) {
	select {
	case tm.statusChan <- status:
	default:
		// Если канал переполнен, игнорируем
	}
}
