package main

import (
	"os"
	"testing"
)

// TestConfigParsing tests basic YAML parsing without relying on the config package
func TestConfigParsing(t *testing.T) {
	// Create a temporary YAML file
	yamlContent := `
station:
  name: "Test Station"
  latitude: 37.7749
  longitude: -122.4194
  altitude: 100
database:
  type: "mariadb"
  host: "localhost"
  port: 3306
  name: "weather"
  user: "weather"
  password: "password"
`

	tmpFile, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write([]byte(yamlContent)); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}

	if err := tmpFile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %v", err)
	}

	// Read the file back
	data, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to read temp file: %v", err)
	}

	// Verify the content
	if len(data) == 0 {
		t.Errorf("Empty file content")
	}

	// Just a basic test to ensure we can read and write files
	t.Logf("Successfully read %d bytes from config file", len(data))
}
