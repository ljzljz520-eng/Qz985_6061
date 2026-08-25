package config

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

type Option func(*Config) error

func WithAddress(value string) Option {
	return func(c *Config) error {
		value = strings.TrimSpace(value)
		if value == "" {
			return fmt.Errorf("address is empty")
		}
		c.Address = value
		return nil
	}
}
func WithDatabasePath(value string) Option {
	return func(c *Config) error {
		value = strings.TrimSpace(value)
		if value == "" {
			return fmt.Errorf("database path is empty")
		}
		clean := filepath.Clean(value)
		if clean == "." {
			return fmt.Errorf("database path must name a file")
		}
		c.DatabasePath = clean
		return nil
	}
}
func WithTimeouts(read, write, shutdown time.Duration) Option {
	return func(c *Config) error {
		if read <= 0 || write <= 0 || shutdown <= 0 {
			return fmt.Errorf("timeouts must be positive")
		}
		c.ReadTimeout = read
		c.WriteTimeout = write
		c.ShutdownTimeout = shutdown
		return nil
	}
}
func WithEnvironment(value string) Option {
	return func(c *Config) error { c.Environment = strings.ToLower(strings.TrimSpace(value)); return nil }
}
func Build(options ...Option) (Config, error) {
	c := Defaults()
	for _, option := range options {
		if option == nil {
			continue
		}
		if err := option(&c); err != nil {
			return Config{}, err
		}
	}
	if err := c.Validate(); err != nil {
		return Config{}, err
	}
	return c, nil
}
func (c Config) RuntimeSummary() string {
	return fmt.Sprintf("address=%s database=%s environment=%s", c.Address, c.DatabasePath, c.Environment)
}
