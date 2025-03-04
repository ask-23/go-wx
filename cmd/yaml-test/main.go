package main

import (
	"fmt"
	"log"

	"gopkg.in/yaml.v3"
)

func main() {
	// A simple program to test yaml.v3
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
		log.Fatalf("Failed to parse YAML: %v", err)
	}

	fmt.Printf("Name: %s, Value: %d\n", result.Name, result.Value)
}
