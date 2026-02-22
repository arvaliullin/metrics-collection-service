package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// LoadFileConfig читает JSON-файл конфигурации по указанному пути и десериализует его в dst.
func LoadFileConfig(path string, dst any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("ошибка чтения файла конфигурации: %w", err)
	}
	if err := json.Unmarshal(data, dst); err != nil {
		return fmt.Errorf("ошибка разбора файла конфигурации: %w", err)
	}
	return nil
}
