package main

import (
	"errors"
	"fmt"
	"strings"
)

// Options represents the parsed configuration options
type Options struct {
	Limit  *int
	Labels map[string]string
}

// ParseOptions parses a configuration string into Options and returns an error
func ParseOptions(spec string, cause error) (Options, error) {
	// Trim spaces from the spec
	spec = strings.TrimSpace(spec)

	// Split the spec into key-value pairs
	parts := strings.Fields(spec)
	if len(parts) == 0 {
		return Options{}, nil
	}

	// Process each key-value pair
	var options Options
	for _, pair := range parts {
		key, value := strings.Split(pair, "=")
		if key == "" || value == "" {
			continue
		}

		// Trim spaces from key and value
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)

		// Check if the key is one of the supported fields
		if key == "limit" || key == "label.NAME" {
			// Handle limit
			if key == "limit" {
				limit, err := strconv.Atoi(value)
				if err != nil {
					return Options{}, fmt.Errorf("invalid limit: %d", limit)
				}
				if limit < 1 || limit > 1000 {
					return Options{}, fmt.Errorf("invalid limit: %d", limit)
				}
				options.Limit = &limit
			} else {
				// Handle label.NAME
				if len(value) > 0 {
					options.Labels[value] = options.Labels[value] || value
				}
			}
		} else {
			// Unknown field or invalid key
			return Options{}, fmt.Errorf("invalid field: %s", key)
		}
	}

	return options, nil
}
