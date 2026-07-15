package config

// LogConfig 日志配置
// Mode: console=仅终端, file=仅文件
type LogConfig struct {
	Level    string `yaml:"Level"`    // debug / info / warn / error
	Mode     string `yaml:"Mode"`     // console / file
	Filename string `yaml:"Filename"` // file 模式时的日志路径
}
