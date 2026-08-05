package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

const configPath = "conf/config.yaml"

type Config struct {
	Server          ServerConfig          `yaml:"server"`
	Mysql           MySQLConfig           `yaml:"Mysql"`
	Redis           RedisConfig           `yaml:"Redis"`
	Email           EmailConfig           `yaml:"Email"`
	Kafka           KafkaConfig           `yaml:"Kafka"`
	CountAggregator CountAggregatorConfig `yaml:"CountAggregator"`
	Log             LogConfig             `yaml:"Log"`
}

type EmailConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	From     string `yaml:"from"`
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
	return cfg, nil
}
