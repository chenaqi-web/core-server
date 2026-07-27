package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

const configPath = "conf/config.yaml"

type Config struct {
	Server          ServerConfig          `yaml:"server"`
	Auth            AuthConfig            `yaml:"auth"`
	Mysql           MySQLConfig           `yaml:"Mysql"`
	Redis           RedisConfig           `yaml:"Redis"`
	Kafka           KafkaConfig           `yaml:"Kafka"`
	CountAggregator CountAggregatorConfig `yaml:"CountAggregator"`
	Log             LogConfig             `yaml:"Log"`
}

type ServerConfig struct {
	Addr string `yaml:"addr"`
}

func Load() (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("read config %s: %w", configPath, err)
	}

	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config %s: %w", configPath, err)
	}
	cfg.Auth.JWTAccessSecret = os.Getenv("JWT_ACCESS_SECRET")
	cfg.Auth.JWTRefreshSecret = os.Getenv("JWT_REFRESH_SECRET")
	cfg.Auth.QQSMTPUsername = os.Getenv("QQ_SMTP_USERNAME")
	cfg.Auth.QQSMTPAuthCode = os.Getenv("QQ_SMTP_AUTH_CODE")
	return cfg, nil
}
