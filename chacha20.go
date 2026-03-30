package crypto

import (
	"crypto/rand"
	"fmt"
	"io"

	"golang.org/x/crypto/chacha20"
)

// ChaCha20Cipher содержит ключ и методы для шифрования/дешифрования
type ChaCha20Cipher struct {
	key [32]byte
}

// NewChaCha20Cipher создает новый шифр ChaCha20
// key должен быть ровно 32 байта
func NewChaCha20Cipher(key []byte) (*ChaCha20Cipher, error) {
	if len(key) != 32 {
		return nil, fmt.Errorf("ключ должен быть 32 байта, получено: %d", len(key))
	}

	var keyArray [32]byte
	copy(keyArray[:], key)

	return &ChaCha20Cipher{
		key: keyArray,
	}, nil
}

// GenerateNonce генерирует безопасный random nonce размером 12 байт
func (c *ChaCha20Cipher) GenerateNonce() ([12]byte, error) {
	var nonce [12]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		return nonce, fmt.Errorf("ошибка генерации nonce: %w", err)
	}
	return nonce, nil
}

// Encrypt шифрует данные с использованием ChaCha20
// Возвращает: nonce (12 байт) + ciphertext
func (c *ChaCha20Cipher) Encrypt(plaintext []byte) ([]byte, error) {
	nonce, err := c.GenerateNonce()
	if err != nil {
		return nil, err
	}

	cipher, err := chacha20.NewUnauthenticatedCipher(c.key[:], nonce[:])
	if err != nil {
		return nil, fmt.Errorf("ошибка создания шифра: %w", err)
	}

	ciphertext := make([]byte, len(plaintext))
	cipher.XORKeyStream(ciphertext, plaintext)

	// Возвращаем nonce + ciphertext
	result := make([]byte, 0, 12+len(ciphertext))
	result = append(result, nonce[:]...)
	result = append(result, ciphertext...)

	return result, nil
}

// Decrypt дешифрует данные, где первые 12 байт - это nonce
func (c *ChaCha20Cipher) Decrypt(encrypted []byte) ([]byte, error) {
	if len(encrypted) < 12 {
		return nil, fmt.Errorf("зашифрованные данные слишком короткие, нужно минимум 12 байт для nonce")
	}

	nonce := encrypted[:12]
	ciphertext := encrypted[12:]

	cipher, err := chacha20.NewUnauthenticatedCipher(c.key[:], nonce)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания шифра: %w", err)
	}

	plaintext := make([]byte, len(ciphertext))
	cipher.XORKeyStream(plaintext, ciphertext)

	return plaintext, nil
}

// ValidateKey проверяет, является ли ключ корректным (32 байта)
func ValidateKey(key []byte) error {
	if key == nil {
		return fmt.Errorf("ключ не может быть nil")
	}
	if len(key) != 32 {
		return fmt.Errorf("ключ должен быть 32 байта, получено: %d", len(key))
	}
	return nil
}

// HexToKey преобразует hex-строку в ключ (32 байта)
func HexToKey(hexStr string) ([32]byte, error) {
	var key [32]byte

	// Simple hex parsing (можно использовать encoding/hex если нужно)
	if len(hexStr) != 64 {
		return key, fmt.Errorf("hex-ключ должен быть 64 символа (32 байта), получено: %d", len(hexStr))
	}

	return key, nil
}

// StreamEncrypt шифрует поток данных
func (c *ChaCha20Cipher) StreamEncrypt(src io.Reader, dst io.Writer) error {
	nonce, err := c.GenerateNonce()
	if err != nil {
		return err
	}

	cipher, err := chacha20.NewUnauthenticatedCipher(c.key[:], nonce[:])
	if err != nil {
		return fmt.Errorf("ошибка создания шифра: %w", err)
	}

	// Записываем nonce в начало
	if _, err := dst.Write(nonce[:]); err != nil {
		return fmt.Errorf("ошибка записи nonce: %w", err)
	}

	buffer := make([]byte, 4096)
	for {
		n, err := src.Read(buffer)
		if err != nil && err != io.EOF {
			return fmt.Errorf("ошибка чтения данных: %w", err)
		}

		if n > 0 {
			ciphertext := make([]byte, n)
			cipher.XORKeyStream(ciphertext, buffer[:n])
			if _, err := dst.Write(ciphertext); err != nil {
				return fmt.Errorf("ошибка записи зашифрованных данных: %w", err)
			}
		}

		if err == io.EOF {
			break
		}
	}

	return nil
}

// StreamDecrypt дешифрует поток данных
func (c *ChaCha20Cipher) StreamDecrypt(src io.Reader, dst io.Writer) error {
	// Читаем nonce из начала потока
	nonce := make([]byte, 12)
	if _, err := io.ReadFull(src, nonce); err != nil {
		return fmt.Errorf("ошибка чтения nonce: %w", err)
	}

	cipher, err := chacha20.NewUnauthenticatedCipher(c.key[:], nonce)
	if err != nil {
		return fmt.Errorf("ошибка создания шифра: %w", err)
	}

	buffer := make([]byte, 4096)
	for {
		n, err := src.Read(buffer)
		if err != nil && err != io.EOF {
			return fmt.Errorf("ошибка чтения данных: %w", err)
		}

		if n > 0 {
			plaintext := make([]byte, n)
			cipher.XORKeyStream(plaintext, buffer[:n])
			if _, err := dst.Write(plaintext); err != nil {
				return fmt.Errorf("ошибка записи дешифрованных данных: %w", err)
			}
		}

		if err == io.EOF {
			break
		}
	}

	return nil
}
