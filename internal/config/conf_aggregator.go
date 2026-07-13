package config

import "time"

type CountAggregatorConfig struct {
	FlushInterval string `yaml:"FlushInterval"`
	BufferSize    int    `yaml:"BufferSize"`
	DBTimeout     string `yaml:"DBTimeout"`
}

func (c CountAggregatorConfig) FlushDuration() time.Duration {
	d, err := time.ParseDuration(c.FlushInterval)
	if err != nil || d <= 0 {
		return 5 * time.Second
	}
	return d
}

func (c CountAggregatorConfig) DBDuration() time.Duration {
	d, err := time.ParseDuration(c.DBTimeout)
	if err != nil || d <= 0 {
		return 10 * time.Second
	}
	return d
}

func (c CountAggregatorConfig) BufferCapacity() int {
	if c.BufferSize <= 0 {
		return 100
	}
	return c.BufferSize
}
