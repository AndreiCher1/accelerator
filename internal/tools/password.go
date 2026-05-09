package tools

import (
	"accelerator/internal/core/config"
	"crypto/rand"
)

const (
    // Набор символов для временного пароля
    passwordChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

func GeneratePassword(cfg *config.Config) (string, error) {
    length := cfg.GeneratePasswordLength
    if length <= 0 {
        length = 12 // значение по умолчанию
    }

    result := make([]byte, length)
    // Используем crypto/rand для чтения случайных байтов, но отображаем их на символы passwordChars
    randomBytes := make([]byte, length)
    _, err := rand.Read(randomBytes)
    if err != nil {
        return "", err
    }

    for i := 0; i < length; i++ {
        // Берём один байт и отображаем на индекс в passwordChars
		randomIndex := int(randomBytes[i]) % len(passwordChars)
        result[i] = passwordChars[randomIndex]
    }
    return string(result), nil
}