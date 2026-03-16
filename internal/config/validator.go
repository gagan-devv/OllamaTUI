package config

import (
	"fmt"
	"regexp"
	"strings"
)

// ValidationError represents a configuration validation error
type ValidationError struct {
	Field   string
	Value   interface{}
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error for %s: %s (value: %v)", e.Field, e.Message, e.Value)
}

// ValidationErrors is a collection of validation errors
type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string {
	if len(e) == 0 {
		return "no validation errors"
	}
	var msgs []string
	for _, err := range e {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// Validate validates the configuration and returns any validation errors
func Validate(config *Config) error {
	var errors ValidationErrors

	// Validate model parameters
	if config.Model.Parameters.Temperature < 0 || config.Model.Parameters.Temperature > 2 {
		errors = append(errors, ValidationError{
			Field:   "model.parameters.temperature",
			Value:   config.Model.Parameters.Temperature,
			Message: "must be between 0 and 2",
		})
	}

	if config.Model.Parameters.TopP < 0 || config.Model.Parameters.TopP > 1 {
		errors = append(errors, ValidationError{
			Field:   "model.parameters.top_p",
			Value:   config.Model.Parameters.TopP,
			Message: "must be between 0 and 1",
		})
	}

	if config.Model.Parameters.TopK <= 0 {
		errors = append(errors, ValidationError{
			Field:   "model.parameters.top_k",
			Value:   config.Model.Parameters.TopK,
			Message: "must be greater than 0",
		})
	}

	// Validate UI colors
	if err := validateColor(config.UI.Colors.UserMessage); err != nil {
		errors = append(errors, ValidationError{
			Field:   "ui.colors.user_message",
			Value:   config.UI.Colors.UserMessage,
			Message: err.Error(),
		})
	}

	if err := validateColor(config.UI.Colors.AIMessage); err != nil {
		errors = append(errors, ValidationError{
			Field:   "ui.colors.ai_message",
			Value:   config.UI.Colors.AIMessage,
			Message: err.Error(),
		})
	}

	if err := validateColor(config.UI.Colors.Background); err != nil {
		errors = append(errors, ValidationError{
			Field:   "ui.colors.background",
			Value:   config.UI.Colors.Background,
			Message: err.Error(),
		})
	}

	if err := validateColor(config.UI.Colors.Border); err != nil {
		errors = append(errors, ValidationError{
			Field:   "ui.colors.border",
			Value:   config.UI.Colors.Border,
			Message: err.Error(),
		})
	}

	if err := validateColor(config.UI.Colors.StatusBar); err != nil {
		errors = append(errors, ValidationError{
			Field:   "ui.colors.status_bar",
			Value:   config.UI.Colors.StatusBar,
			Message: err.Error(),
		})
	}

	// Validate theme
	if config.UI.Theme != "dark" && config.UI.Theme != "light" {
		errors = append(errors, ValidationError{
			Field:   "ui.theme",
			Value:   config.UI.Theme,
			Message: "must be 'dark' or 'light'",
		})
	}

	// Validate network settings
	if config.Network.RetryCount < 0 {
		errors = append(errors, ValidationError{
			Field:   "network.retry_count",
			Value:   config.Network.RetryCount,
			Message: "must be non-negative",
		})
	}

	if config.Network.RetryDelay < 0 {
		errors = append(errors, ValidationError{
			Field:   "network.retry_delay",
			Value:   config.Network.RetryDelay,
			Message: "must be non-negative",
		})
	}

	if config.Network.Timeout < 0 {
		errors = append(errors, ValidationError{
			Field:   "network.timeout",
			Value:   config.Network.Timeout,
			Message: "must be non-negative",
		})
	}

	// Validate behavior settings
	if config.Behavior.CacheSizeMB < 0 {
		errors = append(errors, ValidationError{
			Field:   "behavior.cache_size_mb",
			Value:   config.Behavior.CacheSizeMB,
			Message: "must be non-negative",
		})
	}

	if config.Behavior.AutoSaveInterval < 0 {
		errors = append(errors, ValidationError{
			Field:   "behavior.auto_save_interval",
			Value:   config.Behavior.AutoSaveInterval,
			Message: "must be non-negative",
		})
	}

	// Validate advanced settings
	if config.Advanced.MaxHistory < 0 {
		errors = append(errors, ValidationError{
			Field:   "advanced.max_history",
			Value:   config.Advanced.MaxHistory,
			Message: "must be non-negative",
		})
	}

	if config.Advanced.MemoryLimitMB < 0 {
		errors = append(errors, ValidationError{
			Field:   "advanced.memory_limit_mb",
			Value:   config.Advanced.MemoryLimitMB,
			Message: "must be non-negative",
		})
	}

	if len(errors) > 0 {
		return errors
	}

	return nil
}

// validateColor validates that a color is either a valid hex code or a named color
// validateColor validates that a color is either a valid hex code or a named color
func validateColor(color string) error {
	if color == "" {
		return fmt.Errorf("color cannot be empty")
	}

	// Check if it's a hex color (must start with #)
	hexPattern := regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)
	if hexPattern.MatchString(color) {
		return nil
	}

	// Check if it's a named color (common CSS color names)
	namedColors := map[string]bool{
		"black": true, "white": true, "red": true, "green": true, "blue": true,
		"yellow": true, "cyan": true, "magenta": true, "gray": true, "grey": true,
		"orange": true, "purple": true, "pink": true, "brown": true, "lime": true,
		"navy": true, "teal": true, "olive": true, "maroon": true, "aqua": true,
		"silver": true, "fuchsia": true,
	}

	if namedColors[strings.ToLower(color)] {
		return nil
	}

	return fmt.Errorf("must be a valid hex color (e.g., #5FAFD7) or named color")
}
