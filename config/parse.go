package config

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

type Endpoint struct {
	Host string
	Port int
}

func ParseEndpoint(address string) (Endpoint, error) {
	host, portText, err := net.SplitHostPort(strings.TrimSpace(address))
	if err != nil {
		return Endpoint{}, fmt.Errorf("parse endpoint: %w", err)
	}
	port, err := strconv.Atoi(portText)
	if err != nil || port < 1 || port > 65535 {
		return Endpoint{}, fmt.Errorf("invalid port")
	}
	if host == "" {
		host = "0.0.0.0"
	}
	return Endpoint{Host: host, Port: port}, nil
}

func (e Endpoint) String() string { return net.JoinHostPort(e.Host, strconv.Itoa(e.Port)) }

func (e Endpoint) IsLoopback() bool { ip := net.ParseIP(e.Host); return ip != nil && ip.IsLoopback() }

func (c Config) Endpoint() (Endpoint, error) { return ParseEndpoint(c.Address) }

func (c Config) ValidateDatabaseLocation() error {
	if strings.ContainsAny(c.DatabasePath, "\x00") {
		return fmt.Errorf("database path contains nul")
	}
	if strings.HasSuffix(strings.ToLower(c.DatabasePath), "/") {
		return fmt.Errorf("database path points to directory")
	}
	return nil
}

func Merge(base Config, overrides Config) Config {
	result := base
	if overrides.Address != "" {
		result.Address = overrides.Address
	}
	if overrides.DatabasePath != "" {
		result.DatabasePath = overrides.DatabasePath
	}
	if overrides.ReadTimeout > 0 {
		result.ReadTimeout = overrides.ReadTimeout
	}
	if overrides.WriteTimeout > 0 {
		result.WriteTimeout = overrides.WriteTimeout
	}
	if overrides.ShutdownTimeout > 0 {
		result.ShutdownTimeout = overrides.ShutdownTimeout
	}
	if overrides.MaxBodyBytes > 0 {
		result.MaxBodyBytes = overrides.MaxBodyBytes
	}
	if overrides.Environment != "" {
		result.Environment = overrides.Environment
	}
	return result
}

func EnvironmentAllowed(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "development", "test", "production":
		return true
	default:
		return false
	}
}

func (c Config) Redacted() map[string]string {
	return map[string]string{"address": c.Address, "database": c.DatabasePath, "environment": c.Environment, "read_timeout": c.ReadTimeout.String(), "write_timeout": c.WriteTimeout.String(), "shutdown_timeout": c.ShutdownTimeout.String(), "max_body": strconv.FormatInt(c.MaxBodyBytes, 10)}
}

func ParseBoolean(value string, fallback bool) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return fallback
	}
}

func NormalizeAddress(address string) string {
	address = strings.TrimSpace(address)
	if strings.HasPrefix(address, ":") {
		return "127.0.0.1" + address
	}
	if address == "" {
		return Defaults().Address
	}
	return address
}
