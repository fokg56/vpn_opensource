package crypto

import (
	"testing"
)

// TestNewChaCha20Cipher тестирует создание нового шифра
func TestNewChaCha20Cipher(t *testing.T) {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}

	cipher, err := NewChaCha20Cipher(key)
	if err != nil {
		t.Fatalf("Ошибка при создании шифра: %v", err)
	}

	if cipher == nil {
		t.Fatal("Шифр не должен быть nil")
	}
}

// TestInvalidKeyLength тестирует обработку некорректной длины ключа
func TestInvalidKeyLength(t *testing.T) {
	invalidKeys := [][]byte{
		make([]byte, 16),  // Слишком короткий
		make([]byte, 64),  // Слишком длинный
		nil,               // Nil
	}

	for _, key := range invalidKeys {
		_, err := NewChaCha20Cipher(key)
		if err == nil {
			t.Errorf("Должна быть ошибка для ключа длины %d", len(key))
		}
	}
}

// TestEncryptDecrypt тестирует шифрование и дешифрование
func TestEncryptDecrypt(t *testing.T) {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}

	cipher, err := NewChaCha20Cipher(key)
	if err != nil {
		t.Fatalf("Ошибка при создании шифра: %v", err)
	}

	plaintext := []byte("Это тестовое сообщение для шифрования")

	// Шифруем
	encrypted, err := cipher.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Ошибка при шифровании: %v", err)
	}

	if len(encrypted) != len(plaintext)+12 {
		t.Errorf("Неправильная длина зашифрованного текста: ожидается %d, получено %d",
			len(plaintext)+12, len(encrypted))
	}

	// Дешифруем
	decrypted, err := cipher.Decrypt(encrypted)
	if err != nil {
		t.Fatalf("Ошибка при дешифровании: %v", err)
	}

	// Проверяем оригинальный текст
	if string(decrypted) != string(plaintext) {
		t.Errorf("Дешифрованный текст не соответствует оригиналу\nОжидается: %s\nПолучено: %s",
			string(plaintext), string(decrypted))
	}
}

// TestGenerateNonce тестирует генерацию nonce
func TestGenerateNonce(t *testing.T) {
	key := make([]byte, 32)
	cipher, _ := NewChaCha20Cipher(key)

	nonce1, err := cipher.GenerateNonce()
	if err != nil {
		t.Fatalf("Ошибка при генерации nonce: %v", err)
	}

	nonce2, err := cipher.GenerateNonce()
	if err != nil {
		t.Fatalf("Ошибка при генерации second nonce: %v", err)
	}

	// Nonce должны быть случайными (очень маловероятно что совпадут)
	if nonce1 == nonce2 {
		t.Error("Два сгенерированных nonce не должны быть одинаковыми")
	}
}

// TestValidateKey тестирует валидацию ключа
func TestValidateKey(t *testing.T) {
	tests := []struct {
		name    string
		key     []byte
		wantErr bool
	}{
		{
			name:    "Корректный ключ",
			key:     make([]byte, 32),
			wantErr: false,
		},
		{
			name:    "Короткий ключ",
			key:     make([]byte, 16),
			wantErr: true,
		},
		{
			name:    "Длинный ключ",
			key:     make([]byte, 64),
			wantErr: true,
		},
		{
			name:    "Nil ключ",
			key:     nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateKey(tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateKey() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// BenchmarkEncrypt бенчмарк для шифрования
func BenchmarkEncrypt(b *testing.B) {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}

	cipher, _ := NewChaCha20Cipher(key)
	plaintext := make([]byte, 1024)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cipher.Encrypt(plaintext)
	}
}

// BenchmarkDecrypt бенчмарк для дешифрования
func BenchmarkDecrypt(b *testing.B) {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}

	cipher, _ := NewChaCha20Cipher(key)
	plaintext := make([]byte, 1024)
	encrypted, _ := cipher.Encrypt(plaintext)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cipher.Decrypt(encrypted)
	}
}
