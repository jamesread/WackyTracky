package runtimeconfig

import (
	"net"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_DefaultsAndFile(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	require.NoError(t, os.WriteFile(cfgPath, []byte(`
database:
  driver: todotxt
  database: ./testdata/todotxt
iam:
  database: ./data/iam.sqlite
showFooter: false
`), 0o644))

	origWD, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() { _ = os.Chdir(origWD) })

	t.Setenv("PORT", "")
	require.NoError(t, Load())
	assert.Equal(t, "todotxt", RuntimeConfig.Database.Driver)
	assert.Equal(t, "./testdata/todotxt", RuntimeConfig.Database.Database)
	assert.False(t, RuntimeConfig.ShowFooter)
	assert.True(t, RuntimeConfig.ShowVersionNumber)
	assert.NotEmpty(t, ConfigFileUsed())
}

func TestLoad_EnvOverridesDatabase(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "config.yaml"), []byte("database:\n  driver: yamlfiles\n"), 0o644))
	origWD, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() { _ = os.Chdir(origWD) })

	t.Setenv("PORT", "")
	t.Setenv("DATABASE_DRIVER", "todotxt")
	t.Setenv("DATABASE_DATABASE", "/tmp/todo")
	require.NoError(t, Load())
	assert.Equal(t, "todotxt", RuntimeConfig.Database.Driver)
	assert.Equal(t, "/tmp/todo", RuntimeConfig.Database.Database)
}

func TestResolveListenAddress_UsesPORT(t *testing.T) {
	cfg := defaultConfig()
	cfg.ListenAddress = "0.0.0.0:9999"
	port := freeTCPPort(t)
	t.Setenv("PORT", port)
	require.NoError(t, resolveListenAddress(&cfg))
	assert.Equal(t, ":"+port, cfg.ListenAddress)
}

func freeTCPPort(t *testing.T) string {
	t.Helper()
	ln, err := net.Listen("tcp", ":0")
	require.NoError(t, err)
	defer ln.Close()
	_, portStr, err := net.SplitHostPort(ln.Addr().String())
	require.NoError(t, err)
	_, err = strconv.Atoi(portStr)
	require.NoError(t, err)
	return portStr
}
