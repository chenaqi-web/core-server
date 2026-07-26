package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfigLoad(t *testing.T) {
	t.Setenv("JWT_ACCESS_SECRET", "test-access-secret")
	t.Setenv("JWT_REFRESH_SECRET", "test-refresh-secret")
	t.Setenv("QQ_SMTP_USERNAME", "test@qq.com")
	t.Setenv("QQ_SMTP_AUTH_CODE", "test-auth-code")

	require.NoError(t, os.Chdir("../..")) // internal/config -> core-server
	cfg, err := Load()
	require.NoError(t, err)
	require.Equal(t, "test-access-secret", cfg.Auth.JWTAccessSecret)
	require.Equal(t, "test-refresh-secret", cfg.Auth.JWTRefreshSecret)
	require.Equal(t, "test@qq.com", cfg.Auth.QQSMTPUsername)
	require.Equal(t, "test-auth-code", cfg.Auth.QQSMTPAuthCode)
}
