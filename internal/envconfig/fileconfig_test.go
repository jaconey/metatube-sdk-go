package envconfig

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfigFile(t *testing.T) {
	// Reset configs
	os.Clearenv()
	InitAllEnvConfigs()

	configYAML := `
providers:
  fanza:
    enabled: false
  avbase:
    enabled: true
    priority: 3.5
    proxy: "socks5://proxy:1080"
    timeout: "30s"
    token: "my-token"
  theporndb:
    enabled: false
    priority: 2.0
`

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "metatube.yaml")
	err := os.WriteFile(configPath, []byte(configYAML), 0o644)
	require.NoError(t, err)

	err = LoadConfigFile(configPath, false)
	require.NoError(t, err)

	// fanza should be disabled (priority=0)
	priority, err := MovieProviderConfigs.GetOrDefault("fanza").GetFloat64("priority")
	require.NoError(t, err)
	assert.Equal(t, float64(0), priority)

	// avbase should have priority 3.5 (explicit priority overrides enabled=true)
	priority, err = MovieProviderConfigs.GetOrDefault("avbase").GetFloat64("priority")
	require.NoError(t, err)
	assert.Equal(t, 3.5, priority)

	// avbase proxy
	proxy, err := MovieProviderConfigs.GetOrDefault("avbase").GetString("proxy")
	require.NoError(t, err)
	assert.Equal(t, "socks5://proxy:1080", proxy)

	// avbase timeout
	timeout, err := MovieProviderConfigs.GetOrDefault("avbase").GetString("timeout")
	require.NoError(t, err)
	assert.Equal(t, "30s", timeout)

	// avbase token
	token, err := MovieProviderConfigs.GetOrDefault("avbase").GetString("token")
	require.NoError(t, err)
	assert.Equal(t, "my-token", token)

	// theporndb: enabled=false but priority=2.0 explicitly set, priority wins
	priority, err = MovieProviderConfigs.GetOrDefault("theporndb").GetFloat64("priority")
	require.NoError(t, err)
	assert.Equal(t, 2.0, priority)
}

func TestLoadConfigFile_DefaultPathMissing(t *testing.T) {
	// Should silently skip when defaultPath=true and file doesn't exist
	err := LoadConfigFile("/nonexistent/path/metatube.yaml", true)
	assert.NoError(t, err)
}

func TestLoadConfigFile_ExplicitPathMissing(t *testing.T) {
	// Should error when defaultPath=false and file doesn't exist
	err := LoadConfigFile("/nonexistent/path/metatube.yaml", false)
	assert.Error(t, err)
}

func TestLoadConfigFile_MergesWithEnvConfig(t *testing.T) {
	// Set env config first
	os.Clearenv()
	os.Setenv("MT_MOVIE_PROVIDER_FANZA__PROXY", "http://env-proxy:8080")
	InitAllEnvConfigs()

	configYAML := `
providers:
  fanza:
    timeout: "60s"
`
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "metatube.yaml")
	err := os.WriteFile(configPath, []byte(configYAML), 0o644)
	require.NoError(t, err)

	err = LoadConfigFile(configPath, false)
	require.NoError(t, err)

	// Env-set proxy should still be present
	proxy, err := MovieProviderConfigs.GetOrDefault("fanza").GetString("proxy")
	require.NoError(t, err)
	assert.Equal(t, "http://env-proxy:8080", proxy)

	// File-set timeout should also be present
	timeout, err := MovieProviderConfigs.GetOrDefault("fanza").GetString("timeout")
	require.NoError(t, err)
	assert.Equal(t, "60s", timeout)
}
