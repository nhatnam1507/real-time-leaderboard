//go:build ignore
// +build ignore

// This script generates JSON version of OpenAPI spec from YAML and validates it using kin-openapi
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/getkin/kin-openapi/openapi3"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Usage: %s <input-yaml> <output-json>\n", os.Args[0])
		os.Exit(1)
	}

	yamlFile := os.Args[1]
	jsonFile := os.Args[2]

	ctx := context.Background()

	// Load and validate YAML file using kin-openapi
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromFile(yamlFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading OpenAPI YAML file: %v\n", err)
		os.Exit(1)
	}

	// Validate the OpenAPI specification
	if err := doc.Validate(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Error validating OpenAPI specification: %v\n", err)
		os.Exit(1)
	}

	// Convert to JSON
	jsonData, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error converting to JSON: %v\n", err)
		os.Exit(1)
	}

	// Write JSON file
	if err := os.WriteFile(jsonFile, jsonData, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing JSON file: %v\n", err)
		os.Exit(1)
	}

	// Validate the generated JSON file
	jsonDoc, err := loader.LoadFromFile(jsonFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading generated JSON file: %v\n", err)
		os.Exit(1)
	}

	if err := jsonDoc.Validate(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Error validating generated JSON file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Generated and validated %s from %s\n", jsonFile, yamlFile)
}
