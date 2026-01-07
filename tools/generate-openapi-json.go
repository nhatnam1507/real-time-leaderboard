//go:build ignore
// +build ignore

// This script generates JSON version of OpenAPI spec from YAML
package main

import (
	"encoding/json"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Usage: %s <input-yaml> <output-json>\n", os.Args[0])
		os.Exit(1)
	}

	yamlFile := os.Args[1]
	jsonFile := os.Args[2]

	// Read YAML file
	yamlData, err := os.ReadFile(yamlFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading YAML file: %v\n", err)
		os.Exit(1)
	}

	// Parse YAML
	var data interface{}
	if err := yaml.Unmarshal(yamlData, &data); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing YAML: %v\n", err)
		os.Exit(1)
	}

	// Convert to JSON
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error converting to JSON: %v\n", err)
		os.Exit(1)
	}

	// Write JSON file
	if err := os.WriteFile(jsonFile, jsonData, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing JSON file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Generated %s from %s\n", jsonFile, yamlFile)
}
