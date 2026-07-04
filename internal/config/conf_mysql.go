package config

import "fmt"

type MySQLConfig struct {
	Host        string `json:"host" yaml:"Host"`
	Port        string `json:"port" yaml:"Port"`
	Username    string `json:"username" yaml:"username"`
	Password    string `json:"password" yaml:"password"`
	DBName      string `json:"dbname" yaml:"dbname"`
	Config      string `json:"config" yaml:"config"`
	MaxIdleConn int    `json:"max_idle_conn" yaml:"Max_Idle_Conn"`
	MaxOpenConn int    `json:"max_open_conn" yaml:"Max_Open_Conn"`
}

func (c MySQLConfig) DSN() string {
	cfg := c.Config
	if cfg == "" {
		cfg = "charset=utf8mb4&parseTime=True&loc=Local"
	}
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?%s",
		c.Username, c.Password, c.Host, c.Port, c.DBName, cfg)
}
