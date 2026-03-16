package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// ConfigManager handles loading and saving configuration
type ConfigManager struct {
	configPath string
}

// NewConfigManager creates a new configuration manager
func NewConfigManager(configPath string) *ConfigManager {
	if configPath == "" {
		configPath = getDefaultConfigPath()
	}
	return &ConfigManager{
		configPath: configPath,
	}
}

// getDefaultConfigPath returns the default configuration file path
func getDefaultConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "./config.yaml"
	}
	return filepath.Join(homeDir, ".ollama-go", "config.yaml")
}

// Load loads the configuration from the YAML file
// If the file doesn't exist, it creates it with default values
func (cm *ConfigManager) Load() (*Config, error) {
	// Check if config file exists
	if _, err := os.Stat(cm.configPath); os.IsNotExist(err) {
		// Create default config
		defaultConfig := GetDefault()
		if err := cm.Save(defaultConfig); err != nil {
			return nil, fmt.Errorf("failed to create default config: %w", err)
		}
		return defaultConfig, nil
	}

	// Configure viper
	v := viper.New()
	v.SetConfigFile(cm.configPath)
	v.SetConfigType("yaml")

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Unmarshal into Config struct
	config := &Config{}
	if err := v.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate configuration
	if err := Validate(config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return config, nil
}

// Save saves the configuration to the YAML file
func (cm *ConfigManager) Save(config *Config) error {
	// Ensure config directory exists
	configDir := filepath.Dir(cm.configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Configure viper
	v := viper.New()
	v.SetConfigFile(cm.configPath)
	v.SetConfigType("yaml")

	// Set all config values
	v.Set("model", config.Model)
	v.Set("ui", config.UI)
	v.Set("paths", config.Paths)
	v.Set("behavior", config.Behavior)
	v.Set("network", config.Network)
	v.Set("keybindings", config.Keybindings)
	v.Set("advanced", config.Advanced)

	// Write config file
	if err := v.WriteConfig(); err != nil {
		// If file doesn't exist, create it
		if os.IsNotExist(err) {
			if err := v.SafeWriteConfig(); err != nil {
				return fmt.Errorf("failed to write config file: %w", err)
			}
		} else {
			return fmt.Errorf("failed to write config file: %w", err)
		}
	}

	return nil
}

// GetConfigPath returns the path to the configuration file
func (cm *ConfigManager) GetConfigPath() string {
	return cm.configPath
}
