package config

import (
	"time"
)

// Config represents the application configuration
type Config struct {
	Model    ModelConfig    `mapstructure:"model" yaml:"model"`
	UI       UIConfig       `mapstructure:"ui" yaml:"ui"`
	Paths    PathsConfig    `mapstructure:"paths" yaml:"paths"`
	Behavior BehaviorConfig `mapstructure:"behavior" yaml:"behavior"`
	Network  NetworkConfig  `mapstructure:"network" yaml:"network"`
	Keybindings map[string]string `mapstructure:"keybindings" yaml:"keybindings"`
	Advanced AdvancedConfig `mapstructure:"advanced" yaml:"advanced"`
}

// ModelConfig contains model-related settings
type ModelConfig struct {
	Default    string          `mapstructure:"default" yaml:"default"`
	SystemPrompt string        `mapstructure:"system_prompt" yaml:"system_prompt"`
	Parameters ModelParameters `mapstructure:"parameters" yaml:"parameters"`
}

// ModelParameters contains model generation parameters
type ModelParameters struct {
	Temperature float64 `mapstructure:"temperature" yaml:"temperature"`
	TopP        float64 `mapstructure:"top_p" yaml:"top_p"`
	TopK        int     `mapstructure:"top_k" yaml:"top_k"`
}

// UIConfig contains UI-related settings
type UIConfig struct {
	Theme       string      `mapstructure:"theme" yaml:"theme"`
	VimMode     bool        `mapstructure:"vim_mode" yaml:"vim_mode"`
	ShowMetrics bool        `mapstructure:"show_metrics" yaml:"show_metrics"`
	Colors      ColorScheme `mapstructure:"colors" yaml:"colors"`
}

// ColorScheme defines UI color settings
type ColorScheme struct {
	UserMessage string `mapstructure:"user_message" yaml:"user_message"`
	AIMessage   string `mapstructure:"ai_message" yaml:"ai_message"`
	Background  string `mapstructure:"background" yaml:"background"`
	Border      string `mapstructure:"border" yaml:"border"`
	StatusBar   string `mapstructure:"status_bar" yaml:"status_bar"`
}

// PathsConfig contains path-related settings
type PathsConfig struct {
	TestFolder string `mapstructure:"test_folder" yaml:"test_folder"`
	ConfigDir  string `mapstructure:"config_dir" yaml:"config_dir"`
	PluginDir  string `mapstructure:"plugin_dir" yaml:"plugin_dir"`
}

// BehaviorConfig contains behavior-related settings
type BehaviorConfig struct {
	AutoSave         bool          `mapstructure:"auto_save" yaml:"auto_save"`
	AutoSaveInterval time.Duration `mapstructure:"auto_save_interval" yaml:"auto_save_interval"`
	CacheEnabled     bool          `mapstructure:"cache_enabled" yaml:"cache_enabled"`
	CacheSizeMB      int64         `mapstructure:"cache_size_mb" yaml:"cache_size_mb"`
	ConfirmDestructive bool        `mapstructure:"confirm_destructive" yaml:"confirm_destructive"`
}

// NetworkConfig contains network-related settings
type NetworkConfig struct {
	RetryCount int           `mapstructure:"retry_count" yaml:"retry_count"`
	RetryDelay time.Duration `mapstructure:"retry_delay" yaml:"retry_delay"`
	Timeout    time.Duration `mapstructure:"timeout" yaml:"timeout"`
}

// AdvancedConfig contains advanced settings
type AdvancedConfig struct {
	DebugMode     bool   `mapstructure:"debug_mode" yaml:"debug_mode"`
	LogLevel      string `mapstructure:"log_level" yaml:"log_level"`
	MaxHistory    int    `mapstructure:"max_history" yaml:"max_history"`
	MemoryLimitMB int64  `mapstructure:"memory_limit_mb" yaml:"memory_limit_mb"`
}

// GetDefault returns a Config with default values
func GetDefault() *Config {
	return &Config{
		Model: ModelConfig{
			Default:      "qwen3.5:4b",
			SystemPrompt: "You are a helpful assistant.",
			Parameters: ModelParameters{
				Temperature: 0.7,
				TopP:        0.9,
				TopK:        40,
			},
		},
		UI: UIConfig{
			Theme:       "dark",
			VimMode:     false,
			ShowMetrics: true,
			Colors: ColorScheme{
				UserMessage: "#5FAFD7",
				AIMessage:   "#FFD787",
				Background:  "#1E1E1E",
				Border:      "#3E3E3E",
				StatusBar:   "#2E2E2E",
			},
		},
		Paths: PathsConfig{
			TestFolder: "test_output",
			ConfigDir:  "~/.ollama-go",
			PluginDir:  "~/.ollama-go/plugins",
		},
		Behavior: BehaviorConfig{
			AutoSave:           true,
			AutoSaveInterval:   5 * time.Second,
			CacheEnabled:       true,
			CacheSizeMB:        100,
			ConfirmDestructive: true,
		},
		Network: NetworkConfig{
			RetryCount: 3,
			RetryDelay: 2 * time.Second,
			Timeout:    30 * time.Second,
		},
		Keybindings: map[string]string{
			"quit":         "ctrl+c",
			"submit":       "enter",
			"multiline":    "shift+enter",
			"search":       "ctrl+f",
			"theme_toggle": "ctrl+t",
			"help":         "f1",
			"copy":         "ctrl+y",
		},
		Advanced: AdvancedConfig{
			DebugMode:     false,
			LogLevel:      "info",
			MaxHistory:    1000,
			MemoryLimitMB: 500,
		},
	}
}
