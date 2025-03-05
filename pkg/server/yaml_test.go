package server_test

import (
	"testing"

	"gopkg.in/yaml.v2"
)

func TestYamlParsing(t *testing.T) {
	// A simple test to ensure yaml.v2 is working
	yamlData := []byte(`
name: test
value: 123
`)

	var result struct {
		Name  string `yaml:"name"`
		Value int    `yaml:"value"`
	}

	err := yaml.Unmarshal(yamlData, &result)
	if err != nil {
		t.Fatalf("Failed to parse YAML: %v", err)
	}

	if result.Name != "test" {
		t.Errorf("Expected name 'test', got '%s'", result.Name)
	}

	if result.Value != 123 {
		t.Errorf("Expected value 123, got %d", result.Value)
	}
}
