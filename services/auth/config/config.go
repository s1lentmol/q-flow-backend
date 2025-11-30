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
	Env      string
	TokenTTL time.Duration
	DB       DBConfig
	GRPC     GRPCConfig
}

type DBConfig struct {
	Username string
	Password string
	Host     string
	Port     string
	Name     string
}

type GRPCConfig struct {
	Port    int           `yaml:"port"`
	Timeout time.Duration `yaml:"timeout"`
}

//go:embed config.yaml
var defaultYAML []byte

func Load() (*Config, error) {
	var appConfig Config

	viper.SetConfigType("yaml")

	if err := viper.ReadConfig(bytes.NewBuffer(defaultYAML)); err != nil {
		return nil, fmt.Errorf("config: %w", err)
	}

	viper.AutomaticEnv()
	viper.SetEnvPrefix("env")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	err := viper.Unmarshal(&appConfig)
	if err != nil {
		return nil, fmt.Errorf("config: %w", err)
	}

	return &appConfig, nil
}

func (cfg *Config) GetDSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.DB.Username,
		cfg.DB.Password,
		cfg.DB.Host,
		cfg.DB.Port,
		cfg.DB.Name,
	)
}
