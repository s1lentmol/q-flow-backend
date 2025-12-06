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
	DB       DBConfig
	GRPC     GRPCConfig
	Telegram TelegramConfig
}

type DBConfig struct {
	Username string
	Password string
	Host     string
	Port     string
	Name     string `mapstructure:"database"`
}

type GRPCConfig struct {
	Port    int           `yaml:"port"`
	Timeout time.Duration `yaml:"timeout"`
}

type TelegramConfig struct {
	Token string `yaml:"token" mapstructure:"token"`
	Bot   string `yaml:"bot_name" mapstructure:"bot_name"`
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

func (c *Config) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		c.DB.Username, c.DB.Password, c.DB.Host, c.DB.Port, c.DB.Name)
}
