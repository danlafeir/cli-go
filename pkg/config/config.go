package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

const (
	ConfigFileName = "config"
)

var configPath string

// InitConfig initializes the configuration system with the given path.
// path must be non-empty; callers are responsible for providing the config dir.
func InitConfig(path string) error {
	if path == "" {
		return fmt.Errorf("config path must not be empty")
	}

	configPath = path
	if err := os.MkdirAll(configPath, 0700); err != nil {
		return fmt.Errorf("failed to create config dir: %w", err)
	}
	
	viper.AddConfigPath(configPath)
	viper.SetConfigName(ConfigFileName)
	viper.SetConfigType("yaml")
	
	// Try to read existing config, but don't error if it doesn't exist
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("failed to read config file: %w", err)
		}
	}
	
	return nil
}

func FetchConfig(cmd string) (map[string]interface{}, error) {
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file doesn't exist yet, return empty map
			return make(map[string]interface{}), nil
		}
		return nil, err
	}
	config := viper.GetStringMap(cmd)
	return config, nil
}

// GetConfigValue retrieves a single configuration value for a command.
// Returns the value and true if found, or nil and false if not found.
func GetConfigValue(cmd, key string) (interface{}, bool) {
	fullKey := fmt.Sprintf("%s.%s", cmd, key)
	if viper.IsSet(fullKey) {
		return viper.Get(fullKey), true
	}
	return nil, false
}

// ListConfig returns all configured commands and their keys.
// The returned map has command names as keys and slices of config keys as values.
func ListConfig() (map[string][]string, error) {
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return make(map[string][]string), nil
		}
		return nil, err
	}

	result := make(map[string][]string)
	allSettings := viper.AllSettings()

	for cmd, value := range allSettings {
		if section, ok := value.(map[string]interface{}); ok {
			keys := make([]string, 0, len(section))
			for key := range section {
				keys = append(keys, key)
			}
			result[cmd] = keys
		}
	}

	return result, nil
}

func SetConfigValue(cmd, key string, value interface{}) {
	viper.Set(fmt.Sprintf("%s.%s", cmd, key), value)
}

func WriteConfig() error {
	configFile := filepath.Join(configPath, ConfigFileName+".yaml")
	return viper.WriteConfigAs(configFile)
}

func DeleteConfigValue(cmd, key string) error {
	// First read the current config to ensure we have the latest state
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("failed to read config file: %w", err)
		}
	}
	
	// Get all settings and make a deep copy to modify
	allSettings := viper.AllSettings()
	
	// Navigate to the command section and delete the key
	if cmdSection, exists := allSettings[cmd].(map[string]interface{}); exists {
		delete(cmdSection, key)
		// Reset viper and rebuild the configuration
		viper.Reset()
		viper.AddConfigPath(configPath)
		viper.SetConfigName(ConfigFileName)
		viper.SetConfigType("yaml")
		
		// Set all the modified settings
		for k, v := range allSettings {
			viper.Set(k, v)
		}
	}
	
	return WriteConfig()
}

// ConfigSchema is a list of allowed key patterns under a namespace.
// Use * as a wildcard for any single path segment.
// Example: "jira.teams.*.project" matches "jira.teams.plat.project"
type ConfigSchema []string

// ValidateNamespace checks all keys in config under namespace against schema.
// Returns unrecognized keys (present in config but not matching any pattern).
func ValidateNamespace(namespace string, schema ConfigSchema) []string {
	cfg, err := FetchConfig(namespace)
	if err != nil || len(cfg) == 0 {
		return nil
	}

	keys := flattenKeys("", cfg)
	var unrecognized []string
	for _, key := range keys {
		if !matchesSchema(key, schema) {
			unrecognized = append(unrecognized, key)
		}
	}
	return unrecognized
}

// ClearNamespace prints the current config for the namespace, notifies the
// user the config is invalid, and removes all config under the namespace.
func ClearNamespace(namespace string) error {
	cfg, err := FetchConfig(namespace)
	if err != nil {
		return err
	}

	if len(cfg) > 0 {
		out, _ := yaml.Marshal(cfg)
		fmt.Fprintf(os.Stderr, "Current %s config:\n%s\n", namespace, string(out))
	}

	fmt.Fprintf(os.Stderr, "Config for '%s' contains unrecognized keys and has been cleared. Please reconfigure.\n", namespace)

	allSettings := viper.AllSettings()
	delete(allSettings, namespace)

	viper.Reset()
	viper.AddConfigPath(configPath)
	viper.SetConfigName(ConfigFileName)
	viper.SetConfigType("yaml")

	for k, v := range allSettings {
		viper.Set(k, v)
	}

	return WriteConfig()
}

// flattenKeys recursively produces dot-separated leaf key paths from a nested map.
func flattenKeys(prefix string, m map[string]interface{}) []string {
	var keys []string
	for k, v := range m {
		full := k
		if prefix != "" {
			full = prefix + "." + k
		}
		if sub, ok := v.(map[string]interface{}); ok {
			keys = append(keys, flattenKeys(full, sub)...)
		} else {
			keys = append(keys, full)
		}
	}
	return keys
}

// matchesSchema checks if a key matches any pattern in the schema.
func matchesSchema(key string, schema ConfigSchema) bool {
	for _, pattern := range schema {
		if matchPattern(key, pattern) {
			return true
		}
	}
	return false
}

// matchPattern checks if a dot-separated key matches a dot-separated pattern
// where * matches any single segment.
func matchPattern(key, pattern string) bool {
	keyParts := strings.Split(key, ".")
	patParts := strings.Split(pattern, ".")
	if len(keyParts) != len(patParts) {
		return false
	}
	for i := range keyParts {
		if patParts[i] != "*" && patParts[i] != keyParts[i] {
			return false
		}
	}
	return true
}
