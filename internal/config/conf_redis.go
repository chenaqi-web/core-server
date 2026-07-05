package config

type RedisConfig struct {
	Host     string `json:"host" yaml:"Host"`
	Port     string `json:"port" yaml:"Port"`
	Username string `json:"username" yaml:"username"`
	Password string `json:"password" yaml:"password"`
	DB       int    `json:"db" yaml:"DB"`
}
