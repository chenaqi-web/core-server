package config

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfigLoad(t *testing.T) {
	require.NoError(t, os.Chdir("../..")) // internal/config -> core-server
	cfg, err := Load()
	require.NoError(t, err)
	fmt.Printf("%v", cfg)
}
