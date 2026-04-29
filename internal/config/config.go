package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server struct {
		Port int    `yaml:"port"`
		Env  string `yaml:"env"`
	} `yaml:"server"`
	Database struct {
		Path string `yaml:"path"`
	} `yaml:"database"`
	Polling struct {
		IntervalSeconds int `yaml:"interval_seconds"`
		TimeoutSeconds  int `yaml:"timeout_seconds"`
	} `yaml:"polling"`
	Providers []ProviderConfig `yaml:"providers"`
}

type ProviderConfig struct {
	Name    string `yaml:"name"`
	Slug    string `yaml:"slug"`
	Enabled bool   `yaml:"enabled"`
	Type    string `yaml:"type"`
	URL     string `yaml:"url"`
	APIURL  string `yaml:"api_url"`
}

func Load(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var cfg Config
	if err := yaml.NewDecoder(file).Decode(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
