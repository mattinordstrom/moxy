package config

import (
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Defaults DefaultConfig `yaml:"defaults"`
	Admin    AdminConfig   `yaml:"admin"`
}

type DefaultConfig struct {
	DefaultRoute       string `yaml:"route"`
	PayloadArchivePath string `yaml:"payloadArchivePath"`
}
type AdminConfig struct {
	AutoSortJSONKeys bool `yaml:"autoSortJsonKeys"`
}

var AppConfig Config

func LoadConfig(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	err = yaml.Unmarshal(data, &AppConfig)
	if err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return nil
}
