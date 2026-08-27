package config

import (
	"os"

	"github.com/BurntSushi/toml"
)

// Load читает TOML-файл конфигурации и возвращает структуру Config.
func Load(configFile string) (Config, error) {
	var cfg Config

	data, err := os.ReadFile(configFile)
	if err != nil {
		return cfg, err
	}

	err = toml.Unmarshal(data, &cfg)
	return cfg, err
}
