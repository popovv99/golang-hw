package main

import (
	"os"

	"github.com/BurntSushi/toml"
	"github.com/popovv99/golang-hw/hw12_13_14_15_16_calendar/internal/config"
)

func NewConfig(configFile string) (config.Config, error) {
	var cfg config.Config

	data, err := os.ReadFile(configFile)
	if err != nil {
		return cfg, err
	}

	err = toml.Unmarshal(data, &cfg)
	return cfg, err
}
