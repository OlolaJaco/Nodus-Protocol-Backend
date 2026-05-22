package utils

import (
	"errors"
	"fmt"
	"strings"
)

// ConfigError collects multiple configuration issues into one error.
type ConfigError struct {
	fields []string
}

func (e *ConfigError) Add(field, reason string) {
	e.fields = append(e.fields, fmt.Sprintf("%s: %s", field, reason))
}

func (e *ConfigError) Error() string {
	return "config validation failed:\n  " + strings.Join(e.fields, "\n  ")
}

func (e *ConfigError) HasErrors() bool {
	return len(e.fields) > 0
}

// RequireEnv returns an error if the environment value is empty.
func RequireEnv(name, value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("required environment variable %s is not set", name)
	}
	return nil
}

// ErrMissingConfig is returned when a required config block is absent.
var ErrMissingConfig = errors.New("missing required configuration")
