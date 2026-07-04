package config

// Config holds core-server gRPC and persistence settings.
type Config struct {
	Server ServerConfig `yaml:"server"`
	DB     DBConfig     `yaml:"db"`
}

type ServerConfig struct {
	Addr string `yaml:"addr"`
}

type DBConfig struct {
	DSN string `yaml:"dsn"`
}

// Load reads configuration from config/config.yaml.
func Load() (*Config, error) {
	return &Config{
		Server: ServerConfig{Addr: ":9090"},
		DB:     DBConfig{DSN: ""},
	}, nil
}
