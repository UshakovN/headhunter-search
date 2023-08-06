package config

import (
  "fmt"
  "main/pkg/postgres"
  "os"

  "gopkg.in/yaml.v2"
)

type Config struct {
  Postgres *postgres.Config `json:"postgres"`
  Telegram string           `json:"telegram"`
}

func NewConfig(file string) (*Config, error) {
  buf, err := os.ReadFile(file)
  if err != nil {
    return nil, fmt.Errorf("cannot read yaml config file: %s: %v", file, err)
  }
  config := &Config{}

  if err = yaml.Unmarshal(buf, config); err != nil {
    return nil, fmt.Errorf("cannot unmarshal config form yaml to struct: %v", err)
  }
  return config, nil
}
