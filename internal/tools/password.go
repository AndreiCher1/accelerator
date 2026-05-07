package tools

import (
	"accelerator/internal/core/config"
	"crypto/rand"
)

func GeneratePassword(cfg config.Config) (string, error) {
	b := make([]byte, cfg.GeneratePasswordLength)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	// Используем URLEncoding, чтобы строка была безопасной для URL
	return string(b), nil
}
