package config

import (
	"fmt"
	"time"
)

const (
	defaultAccessExpire  = 20 * time.Minute
	defaultRefreshExpire = 7 * 24 * time.Hour
)

type AuthConfig struct {
	AccessExpire  string `yaml:"access_expire"`
	RefreshExpire string `yaml:"refresh_expire"`

	JWTAccessSecret  string `yaml:"-"`
	JWTRefreshSecret string `yaml:"-"`
	QQSMTPUsername   string `yaml:"-"`
	QQSMTPAuthCode   string `yaml:"-"`
}

func (c AuthConfig) AccessDuration() (time.Duration, error) {
	return parsePositiveDuration(c.AccessExpire, defaultAccessExpire, "auth access_expire")
}

func (c AuthConfig) RefreshDuration() (time.Duration, error) {
	return parsePositiveDuration(c.RefreshExpire, defaultRefreshExpire, "auth refresh_expire")
}

func parsePositiveDuration(value string, fallback time.Duration, field string) (time.Duration, error) {
	if value == "" {
		return fallback, nil
	}
	duration, err := time.ParseDuration(value)
	if err != nil || duration <= 0 {
		return 0, fmt.Errorf("%s must be a positive duration", field)
	}
	return duration, nil
}
