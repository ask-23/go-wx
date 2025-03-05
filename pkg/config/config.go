package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

// Config represents the main application configuration
type Config struct {
	Station    StationConfig     `yaml:"station"`
	Database   DatabaseConfig    `yaml:"database"`
	Collector  CollectorConfig   `yaml:"collector"`
	Server     ServerConfig      `yaml:"server"`
	Publishers []PublisherConfig `yaml:"publishers"`
	Logging    LoggingConfig     `yaml:"logging"`
}

// StationConfig contains information about the weather station
type StationConfig struct {
	Name     string         `yaml:"name"`
	Location LocationConfig `yaml:"location"`
}

// LocationConfig contains geographic information
type LocationConfig struct {
	Latitude  float64 `yaml:"latitude"`
	Longitude float64 `yaml:"longitude"`
	Altitude  float64 `yaml:"altitude"` // meters
}

// DatabaseConfig contains database connection settings
type DatabaseConfig struct {
	Type     string `yaml:"type"` // mariadb or postgres
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Name     string `yaml:"name"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
}

// CollectorConfig contains settings for data collection
type CollectorConfig struct {
	Type     string       `yaml:"type"` // interceptor or other methods
	Device   DeviceConfig `yaml:"device"`
	Interval int          `yaml:"interval"` // seconds
}

// DeviceConfig contains information about the weather device
type DeviceConfig struct {
	Type    string `yaml:"type"`    // ecowitt, etc
	Model   string `yaml:"model"`   // GW1000, etc
	Address string `yaml:"address"` // IP address
	Port    int    `yaml:"port"`    // Port to listen on
}

// ServerConfig contains web server settings
type ServerConfig struct {
	Type    string    `yaml:"type"` // caddy or nginx
	Port    int       `yaml:"port"`
	Address string    `yaml:"address"`
	SSL     SSLConfig `yaml:"ssl"`
}

// SSLConfig contains SSL/TLS settings
type SSLConfig struct {
	Enabled  bool   `yaml:"enabled"`
	CertFile string `yaml:"cert_file"`
	KeyFile  string `yaml:"key_file"`
}

// PublisherConfig contains settings for weather data publishing
type PublisherConfig struct {
	Name      string            `yaml:"name"`
	Enabled   bool              `yaml:"enabled"`
	StationID string            `yaml:"station_id,omitempty"`
	APIKey    string            `yaml:"api_key,omitempty"`
	URL       string            `yaml:"url,omitempty"`
	Method    string            `yaml:"method,omitempty"`
	Headers   map[string]string `yaml:"headers,omitempty"`
	Interval  int               `yaml:"interval"` // seconds
}

// LoggingConfig contains logging settings
type LoggingConfig struct {
	Level      string `yaml:"level"`
	File       string `yaml:"file"`
	MaxSize    int    `yaml:"max_size"`    // MB
	MaxBackups int    `yaml:"max_backups"` // number of backup files
	MaxAge     int    `yaml:"max_age"`     // days
}

// LoadConfig loads the configuration from a YAML file
func LoadConfig(filename string) (*Config, error) {
	// Read the configuration file
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	// Parse the YAML configuration
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("error parsing config file: %w", err)
	}

	// Apply default values if needed
	applyDefaults(&config)

	// Validate configuration
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

// applyDefaults fills in default values for missing configuration settings
func applyDefaults(config *Config) {
	// Set default server port if not specified
	if config.Server.Port == 0 {
		config.Server.Port = 8080
	}

	// Set default server address if not specified
	if config.Server.Address == "" {
		config.Server.Address = "0.0.0.0"
	}

	// Set default database port if not specified
	if config.Database.Port == 0 {
		config.Database.Port = 3306 // Default MySQL/MariaDB port
	}

	// Set default logging level if not specified
	if config.Logging.Level == "" {
		config.Logging.Level = "info"
	}

	// Set default collector interval if not specified
	if config.Collector.Interval == 0 {
		config.Collector.Interval = 60 // 1 minute default
	}
}

// validateConfig verifies that the configuration is valid
func validateConfig(config *Config) error {
	// Validate database configuration
	if config.Database.Type != "mariadb" && config.Database.Type != "postgres" {
		return fmt.Errorf("database type must be 'mariadb' or 'postgres'")
	}

	// Validate server configuration
	if config.Server.Type != "caddy" && config.Server.Type != "nginx" {
		return fmt.Errorf("server type must be 'caddy' or 'nginx'")
	}

	// If SSL is enabled, verify that certificate and key files are specified
	if config.Server.SSL.Enabled {
		if config.Server.SSL.CertFile == "" || config.Server.SSL.KeyFile == "" {
			return fmt.Errorf("SSL is enabled but certificate or key file is not specified")
		}
	}

	return nil
}
