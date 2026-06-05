package notifier

import (
	"errors"
	"os"
	"strings"
)

func SaveChatID(filePath string, chatID string) error {
	if strings.TrimSpace(chatID) == "" {
		return errors.New("empty chatID")
	}
	return os.WriteFile(filePath, []byte(strings.TrimSpace(chatID)), 0o644)
}

func LoadChatID(filePath string) (string, error) {
	b, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(b)), nil
}
