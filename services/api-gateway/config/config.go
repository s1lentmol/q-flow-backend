package config

import (
	"bytes"
	_ "embed"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Env  string
	HTTP HTTPConfig
	GRPC GRPCConfig
	App  AppConfig
}

type HTTPConfig struct {
	Port int `yaml:"port"`
}

type GRPCConfig struct {
	AuthAddress         string        `yaml:"auth_address"`
	QueueAddress        string        `yaml:"queue_address"`
	NotificationAddress string        `yaml:"notification_address"`
	Timeout             time.Duration `yaml:"timeout"`
}

type AppConfig struct {
	ID     int    `yaml:"id"`
	Secret string `yaml:"secret"`
}

//go:embed config.yaml
var defaultYAML []byte

func Load() (*Config, error) {
	var cfg Config

	viper.SetConfigType("yaml")
	if err := viper.ReadConfig(bytes.NewBuffer(defaultYAML)); err != nil {
		return nil, fmt.Errorf("config: %w", err)
	}

	viper.AutomaticEnv()
	viper.SetEnvPrefix("env")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("config: %w", err)
	}

	return &cfg, nil
}
