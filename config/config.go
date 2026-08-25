package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Address         string
	DatabasePath    string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
	MaxBodyBytes    int64
	Environment     string
}

func Defaults() Config {
	return Config{Address: "127.0.0.1:8080", DatabasePath: "knowledge.db", ReadTimeout: 5 * time.Second, WriteTimeout: 10 * time.Second, ShutdownTimeout: 5 * time.Second, MaxBodyBytes: 1 << 20, Environment: "development"}
}

func Load() (Config, error) {
	c := Defaults()
	if value := strings.TrimSpace(os.Getenv("KNOWLEDGE_ADDRESS")); value != "" {
		c.Address = value
	}
	if value := strings.TrimSpace(os.Getenv("KNOWLEDGE_DB")); value != "" {
		c.DatabasePath = value
	}
	if value := strings.TrimSpace(os.Getenv("KNOWLEDGE_ENV")); value != "" {
		c.Environment = value
	}
	if value := strings.TrimSpace(os.Getenv("KNOWLEDGE_MAX_BODY")); value != "" {
		parsed, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return Config{}, fmt.Errorf("parse max body: %w", err)
		}
		c.MaxBodyBytes = parsed
	}
	if err := c.Validate(); err != nil {
		return Config{}, err
	}
	return c, nil
}

func (c Config) Validate() error {
	if strings.TrimSpace(c.Address) == "" {
		return fmt.Errorf("address is required")
	}
	if strings.TrimSpace(c.DatabasePath) == "" {
		return fmt.Errorf("database path is required")
	}
	if c.ReadTimeout <= 0 || c.WriteTimeout <= 0 || c.ShutdownTimeout <= 0 {
		return fmt.Errorf("timeouts must be positive")
	}
	if c.MaxBodyBytes < 1024 {
		return fmt.Errorf("max body must be at least 1024")
	}
	switch c.Environment {
	case "development", "test", "production":
	default:
		return fmt.Errorf("unsupported environment")
	}
	return nil
}

func (c Config) IsProduction() bool { return c.Environment == "production" }
