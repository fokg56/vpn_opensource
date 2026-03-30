package vless

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"time"
)

// CommandType тип команды в протоколе VLESS
type CommandType byte

const (
	CmdTCP CommandType = 1
	CmdUDP CommandType = 2
)

// Handshake структура для рукопожатия VLESS
type Handshake struct {
	UUID    [16]byte    // UUID пользователя
	AddrLen byte        // Длина адреса
	Addr    []byte      // Адрес назначения (домен или IP)
	Port    uint16      // Порт назначения
	Command CommandType // Тип команды (TCP/UDP)
}

// RequestHeader заголовок запроса VLESS
type RequestHeader struct {
	Version     byte        // Версия протокола VLESS
	Command     CommandType // Команда
	AddressType byte        // Тип адреса (1 = IPv4, 2 = домен, 3 = IPv6)
	Address     []byte      // Адрес
	Port        uint16      // Порт
	Timestamp   uint64      // Временная метка
}

// NewRequestHeader создает новый заголовок запроса
func NewRequestHeader(uuid [16]byte, cmdType CommandType, address string, port uint16) *RequestHeader {
	addrBytes := []byte(address)

	return &RequestHeader{
		Version:     0,
		Command:     cmdType,
		AddressType: 2, // 2 = domain name
		Address:     addrBytes,
		Port:        port,
		Timestamp:   uint64(time.Now().Unix()),
	}
}

// EncodeRequest кодирует запрос в VLESS формат
func EncodeRequest(uuid [16]byte, address string, port uint16) ([]byte, error) {
	addrBytes := []byte(address)
	if len(addrBytes) > 255 {
		return nil, fmt.Errorf("адрес слишком длинный: %d байт (максимум 255)", len(addrBytes))
	}

	// Структура запроса:
	// 1 byte - версия (0)
	// 1 byte - команда (1 = TCP)
	// 16 bytes - UUID
	// 1 byte - тип адреса (2 = домен, 1 = IPv4, 3 = IPv6)
	// 1 byte - длина адреса (если домен)
	// N bytes - адрес
	// 2 bytes - порт

	buf := make([]byte, 0, 25+len(addrBytes))

	// Версия
	buf = append(buf, 0)

	// Команда
	buf = append(buf, byte(CmdTCP))

	// UUID
	buf = append(buf, uuid[:]...)

	// Тип адреса
	buf = append(buf, 2) // домен

	// Длина адреса
	buf = append(buf, byte(len(addrBytes)))

	// Адрес
	buf = append(buf, addrBytes...)

	// Порт (big-endian)
	portBuf := make([]byte, 2)
	binary.BigEndian.PutUint16(portBuf, port)
	buf = append(buf, portBuf...)

	return buf, nil
}

// DecodeRequest декодирует запрос VLESS
func DecodeRequest(data []byte) (*RequestHeader, error) {
	if len(data) < 20 {
		return nil, fmt.Errorf("данные слишком короткие для заголовка VLESS")
	}

	header := &RequestHeader{
		Version: data[0],
	}

	if header.Version != 0 {
		return nil, fmt.Errorf("неподдерживаемая версия VLESS: %d", header.Version)
	}

	offset := 1
	header.Command = CommandType(data[offset])
	offset++

	// UUID (16 байт)
	uuidBytes := make([]byte, 16)
	copy(uuidBytes, data[offset:offset+16])
	offset += 16

	// Тип адреса
	header.AddressType = data[offset]
	offset++

	// Адрес
	switch header.AddressType {
	case 1: // IPv4
		if len(data) < offset+4+2 {
			return nil, fmt.Errorf("недостаточно данных для IPv4 адреса")
		}
		header.Address = make([]byte, 4)
		copy(header.Address, data[offset:offset+4])
		offset += 4

	case 2: // Домен
		if len(data) < offset+1 {
			return nil, fmt.Errorf("недостаточно данных для длины домена")
		}
		addrLen := data[offset]
		offset++

		if len(data) < offset+int(addrLen)+2 {
			return nil, fmt.Errorf("недостаточно данных для домена и порта")
		}

		header.Address = make([]byte, addrLen)
		copy(header.Address, data[offset:offset+int(addrLen)])
		offset += int(addrLen)

	case 3: // IPv6
		if len(data) < offset+16+2 {
			return nil, fmt.Errorf("недостаточно данных для IPv6 адреса")
		}
		header.Address = make([]byte, 16)
		copy(header.Address, data[offset:offset+16])
		offset += 16

	default:
		return nil, fmt.Errorf("неизвестный тип адреса: %d", header.AddressType)
	}

	// Порт (big-endian)
	if len(data) < offset+2 {
		return nil, fmt.Errorf("недостаточно данных для порта")
	}
	header.Port = binary.BigEndian.Uint16(data[offset : offset+2])

	return header, nil
}

// ClientConnection соединение VLESS клиента
type ClientConnection struct {
	conn   net.Conn
	addr   string
	port   uint16
	UUID   [16]byte
	closed bool
}

// NewClientConnection создает новое VLESS соединение
func NewClientConnection(serverAddr string, serverPort uint16, uuid [16]byte) (*ClientConnection, error) {
	address := fmt.Sprintf("%s:%d", serverAddr, serverPort)
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, fmt.Errorf("ошибка подключения к серверу: %w", err)
	}

	return &ClientConnection{
		conn:   conn,
		addr:   serverAddr,
		port:   serverPort,
		UUID:   uuid,
		closed: false,
	}, nil
}

// Handshake выполняет рукопожатие с сервером
func (cc *ClientConnection) Handshake(address string, destPort uint16) error {
	request, err := EncodeRequest(cc.UUID, address, destPort)
	if err != nil {
		return fmt.Errorf("ошибка кодирования запроса: %w", err)
	}

	if _, err := cc.conn.Write(request); err != nil {
		return fmt.Errorf("ошибка отправки хендшейка: %w", err)
	}

	// Читаем ответ (простой ок-ответ)
	response := make([]byte, 1)
	if _, err := io.ReadFull(cc.conn, response); err != nil {
		return fmt.Errorf("ошибка получения ответа сервера: %w", err)
	}

	return nil
}

// Write записывает данные в соединение
func (cc *ClientConnection) Write(data []byte) (int, error) {
	if cc.closed {
		return 0, fmt.Errorf("соединение закрыто")
	}
	return cc.conn.Write(data)
}

// Read читает данные из соединения
func (cc *ClientConnection) Read(data []byte) (int, error) {
	if cc.closed {
		return 0, fmt.Errorf("соединение закрыто")
	}
	return cc.conn.Read(data)
}

// Close закрывает соединение
func (cc *ClientConnection) Close() error {
	cc.closed = true
	return cc.conn.Close()
}

// SetDeadline устанавливает дедлайн для соединения
func (cc *ClientConnection) SetDeadline(t time.Time) error {
	return cc.conn.SetDeadline(t)
}

// SetReadDeadline устанавливает дедлайн для чтения
func (cc *ClientConnection) SetReadDeadline(t time.Time) error {
	return cc.conn.SetReadDeadline(t)
}

// SetWriteDeadline устанавливает дедлайн для записи
func (cc *ClientConnection) SetWriteDeadline(t time.Time) error {
	return cc.conn.SetWriteDeadline(t)
}

// LocalAddr возвращает локальный адрес соединения
func (cc *ClientConnection) LocalAddr() net.Addr {
	return cc.conn.LocalAddr()
}

// RemoteAddr возвращает удаленный адрес соединения
func (cc *ClientConnection) RemoteAddr() net.Addr {
	return cc.conn.RemoteAddr()
}
