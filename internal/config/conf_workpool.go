package config

import "time"

type WorkpoolConfig struct {
	Workers         int    `yaml:"Workers"`
	QueueSize       int    `yaml:"QueueSize"`
	ShutdownTimeout string `yaml:"ShutdownTimeout"`
}

func (c WorkpoolConfig) WorkerCount() int {
	if c.Workers <= 0 {
		return 8
	}
	return c.Workers
}

func (c WorkpoolConfig) BufferSize() int {
	if c.QueueSize <= 0 {
		return 1024
	}
	return c.QueueSize
}

func (c WorkpoolConfig) StopTimeout() time.Duration {
	d, err := time.ParseDuration(c.ShutdownTimeout)
	if err != nil || d <= 0 {
		return 10 * time.Second
	}
	return d
}
