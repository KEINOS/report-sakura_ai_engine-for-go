package options

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// Options holds the parsed configuration.
type Options struct {
	Limit  *int
	Labels map[string]string
}

// ParseError describes a problem with a single field while parsing.
type ParseError struct {
	Field string // the key as it appeared in the spec
	Value string // the raw value (empty if not applicable)
	Err   error  // underlying error
}

// Error implements the error interface.
func (p *ParseError) Error() string {
	if p.Value == "" {
		return fmt.Sprintf("field %q: %v", p.Field, p.Err)
	}
	return fmt.Sprintf("field %q (value %q): %v", p.Field, p.Value, p.Err)
}

// Unwrap returns the wrapped error.
func (p *ParseError) Unwrap() error { return p.Err }

// ParseOptions parses a comma‑separated specification string.
// If cause (or any error it wraps) is a *ParseError, that error is returned
// together with the options parsed from spec, ignoring any parsing errors
// from spec itself.
func ParseOptions(spec string, cause error) (Options, error) {
	// start with defaults
	opts := Options{
		Limit:  func() *int { v := 100; return &v }(),
		Labels: make(map[string]string),
	}

	// Helper to return a ParseError preserving the partially built opts.
	newParseError := func(field, value string, err error) (*ParseError, Options) {
		return &ParseError{Field: field, Value: value, Err: err}, opts
	}

	// Fast‑path: if spec is empty, just return defaults (unless cause has a ParseError).
	if strings.TrimSpace(spec) == "" {
		if pe := findParseError(cause); pe != nil {
			return opts, pe
		}
		return opts, nil
	}

	// Split on commas, but keep empty entries to detect syntax errors like ",,".
	fields := strings.Split(spec, ",")
	for _, raw := range fields {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			// ignore empty fragments (e.g., trailing commas)
			continue
		}
		kvIdx := strings.Index(raw, "=")
		if kvIdx < 0 {
			return newParseError(raw, "", fmt.Errorf("missing \"=\""))
		}
		key := strings.TrimSpace(raw[:kvIdx])
		val := strings.TrimSpace(raw[kvIdx+1:])

		switch {
		case key == "limit":
			i, err := strconv.Atoi(val)
			if err != nil {
				return newParseError(key, val, fmt.Errorf("invalid integer: %w", err))
			}
			if i < 1 || i > 1000 {
				return newParseError(key, val, fmt.Errorf("value out of range (1‑1000)"))
			}
			// store a new pointer
			v := i
			opts.Limit = &v

		case strings.HasPrefix(key, "label."):
			name := strings.TrimPrefix(key, "label.")
			if name == "" {
				return newParseError(key, val, fmt.Errorf("label name is empty"))
			}
			opts.Labels[name] = val

		default:
			return newParseError(key, val, fmt.Errorf("unknown field"))
		}
	}

	// If cause (or any wrapped error) is a *ParseError, return it instead of any spec error.
	if pe := findParseError(cause); pe != nil {
		return opts, pe
	}
	return opts, nil
}

// findParseError walks the error chain of err and returns the first *ParseError found.
func findParseError(err error) *ParseError {
	var pe *ParseError
	if err != nil && errors.As(err, &pe) {
		return pe
	}
	return nil
}
