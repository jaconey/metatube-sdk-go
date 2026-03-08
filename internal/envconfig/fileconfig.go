package envconfig

import (
	"errors"
	"fmt"
	"os"

	"github.com/metatube-community/metatube-sdk-go/collection/maps"
	"gopkg.in/yaml.v3"
)

const DefaultConfigPath = "/config/metatube.yaml"

// ProviderFileConfig represents per-provider settings in the config file.
type ProviderFileConfig struct {
	Enabled  *bool             `yaml:"enabled"`
	Priority *float64          `yaml:"priority"`
	Proxy    string            `yaml:"proxy,omitempty"`
	Timeout  string            `yaml:"timeout,omitempty"`
	Token    string            `yaml:"token,omitempty"`
	Extra    map[string]string `yaml:",inline"`
}

// FileConfig represents the top-level YAML config file structure.
type FileConfig struct {
	Providers map[string]ProviderFileConfig `yaml:"providers"`
}

// LoadConfigFile reads a YAML config file and merges provider settings into
// the global ActorProviderConfigs and MovieProviderConfigs maps.
// If defaultPath is true, a missing file is silently ignored.
func LoadConfigFile(path string, defaultPath bool) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if defaultPath && errors.Is(err, os.ErrNotExist) {
			return nil // silently skip if default path doesn't exist
		}
		return fmt.Errorf("failed to read config file %s: %w", path, err)
	}

	var cfg FileConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("failed to parse config file %s: %w", path, err)
	}

	applyFileConfig(cfg)
	return nil
}

func applyFileConfig(cfg FileConfig) {
	for name, pc := range cfg.Providers {
		config := providerFileConfigToConfig(pc)

		// Apply to both actor and movie provider configs.
		// The engine will pick up whichever applies.
		mergeIntoProviderConfig(ActorProviderConfigs, name, config)
		mergeIntoProviderConfig(MovieProviderConfigs, name, config)
	}
}

func providerFileConfigToConfig(pc ProviderFileConfig) *Config {
	c := NewConfig()

	// If explicitly disabled, set priority to 0.
	if pc.Enabled != nil && !*pc.Enabled {
		c.Set("priority", "0")
	}

	// Explicit priority overrides enabled flag.
	if pc.Priority != nil {
		c.Set("priority", fmt.Sprintf("%g", *pc.Priority))
	}

	if pc.Proxy != "" {
		c.Set("proxy", pc.Proxy)
	}
	if pc.Timeout != "" {
		c.Set("timeout", pc.Timeout)
	}
	if pc.Token != "" {
		c.Set("token", pc.Token)
	}

	// Apply any extra keys from inline map.
	for k, v := range pc.Extra {
		c.Set(k, v)
	}

	return c
}

func mergeIntoProviderConfig(target *maps.CaseInsensitiveMap[*Config], name string, source *Config) {
	if !target.Has(name) {
		target.Set(name, NewConfig())
	}
	existing := target.GetOrDefault(name)
	for k, v := range source.Iterator() {
		existing.Set(k, v)
	}
}
