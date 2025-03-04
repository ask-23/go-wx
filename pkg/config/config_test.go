package config

import (
	"os"
	"testing"
)

// TestLoadConfig tests the loading of configuration from a YAML file
func TestLoadConfig(t *testing.T) {
	// Create a temporary test config file
	tempFile, err := os.CreateTemp("", "config_test_*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	// Write test configuration to the file
	testConfig := `
station:
  name: "Test Station"
  location:
    latitude: 42.0
    longitude: -71.0
    altitude: 100
database:
  type: "mariadb"
  host: "localhost"
  port: 3306
  name: "test_db"
  user: "testuser"
  password: "testpass"
collector:
  type: "interceptor"
  device:
    type: "ecowitt"
    model: "GW1000"
    address: "192.168.1.100"
    port: 8080
  interval: 60
server:
  type: "caddy"
  port: 8080
  address: "0.0.0.0"
  ssl:
    enabled: false
    cert_file: "cert.pem"
    key_file: "key.pem"
publishers:
  - name: "wunderground"
    enabled: true
    station_id: "TESTID"
    api_key: "testkey"
    interval: 300
logging:
  level: "info"
  file: "test.log"
  max_size: 10
  max_backups: 3
  max_age: 28
`
	if _, err := tempFile.Write([]byte(testConfig)); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	if err := tempFile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}

	// Test LoadConfig function
	cfg, err := LoadConfig(tempFile.Name())
	if err != nil {
		t.Fatalf("LoadConfig returned error: %v", err)
	}

	// Verify config values
	if cfg.Station.Name != "Test Station" {
		t.Errorf("Expected station name 'Test Station', got '%s'", cfg.Station.Name)
	}

	if cfg.Database.Type != "mariadb" {
		t.Errorf("Expected database type 'mariadb', got '%s'", cfg.Database.Type)
	}

	if cfg.Collector.Device.Type != "ecowitt" {
		t.Errorf("Expected device type 'ecowitt', got '%s'", cfg.Collector.Device.Type)
	}

	if cfg.Server.Port != 8080 {
		t.Errorf("Expected server port 8080, got %d", cfg.Server.Port)
	}

	if !cfg.Publishers[0].Enabled {
		t.Errorf("Expected first publisher to be enabled")
	}

	if cfg.Logging.Level != "info" {
		t.Errorf("Expected logging level 'info', got '%s'", cfg.Logging.Level)
	}
}

// TestValidateConfig tests the configuration validation logic
func TestValidateConfig(t *testing.T) {
	// Test valid config
	validConfig := &Config{
		Station: StationConfig{
			Name: "Test Station",
			Location: LocationConfig{
				Latitude:  42.0,
				Longitude: -71.0,
				Altitude:  100,
			},
		},
		Database: DatabaseConfig{
			Type:     "mariadb",
			Host:     "localhost",
			Port:     3306,
			Name:     "test_db",
			User:     "testuser",
			Password: "testpass",
		},
		Collector: CollectorConfig{
			Type: "interceptor",
			Device: DeviceConfig{
				Type:    "ecowitt",
				Model:   "GW1000",
				Address: "192.168.1.100",
				Port:    8080,
			},
			Interval: 60,
		},
		Server: ServerConfig{
			Type:    "caddy",
			Port:    8080,
			Address: "0.0.0.0",
			SSL: SSLConfig{
				Enabled:  false,
				CertFile: "cert.pem",
				KeyFile:  "key.pem",
			},
		},
		Publishers: []PublisherConfig{
			{
				Name:      "wunderground",
				Enabled:   true,
				StationID: "TESTID",
				APIKey:    "testkey",
				Interval:  300,
			},
		},
		Logging: LoggingConfig{
			Level:      "info",
			File:       "test.log",
			MaxSize:    10,
			MaxBackups: 3,
			MaxAge:     28,
		},
	}

	if err := validateConfig(validConfig); err != nil {
		t.Errorf("validateConfig returned error for valid config: %v", err)
	}

	// Test invalid config - missing station name
	invalidConfig := &Config{
		Station: StationConfig{
			Name: "", // Empty station name
			Location: LocationConfig{
				Latitude:  42.0,
				Longitude: -71.0,
				Altitude:  100,
			},
		},
		Database: DatabaseConfig{
			Type:     "mariadb",
			Host:     "localhost",
			Port:     3306,
			Name:     "test_db",
			User:     "testuser",
			Password: "testpass",
		},
		Collector: CollectorConfig{
			Type: "interceptor",
			Device: DeviceConfig{
				Type:    "ecowitt",
				Model:   "GW1000",
				Address: "192.168.1.100",
				Port:    8080,
			},
			Interval: 60,
		},
	}

	if err := validateConfig(invalidConfig); err == nil {
		t.Errorf("validateConfig did not return error for invalid config (missing station name)")
	}
}

// TestApplyDefaults tests the default value application
func TestApplyDefaults(t *testing.T) {
	// Create minimal config
	minimalConfig := &Config{
		Station: StationConfig{
			Name: "Test Station",
		},
		Database: DatabaseConfig{
			Type: "mariadb",
			Host: "localhost",
			Name: "gowx",
			User: "gowx",
		},
	}

	// Apply defaults
	applyDefaults(minimalConfig)

	// Check defaults were applied
	if minimalConfig.Database.Port != 3306 {
		t.Errorf("Default database port not applied, expected 3306, got %d", minimalConfig.Database.Port)
	}

	if minimalConfig.Server.Port != 8080 {
		t.Errorf("Default server port not applied, expected 8080, got %d", minimalConfig.Server.Port)
	}

	if minimalConfig.Logging.Level != "info" {
		t.Errorf("Default logging level not applied, expected 'info', got '%s'", minimalConfig.Logging.Level)
	}

	if minimalConfig.Collector.Interval != 60 {
		t.Errorf("Default collector interval not applied, expected 60, got %d", minimalConfig.Collector.Interval)
	}
}
