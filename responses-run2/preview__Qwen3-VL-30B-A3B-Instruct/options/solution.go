package options

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type Options struct {
	Limit  *int
	Labels map[string]string
}

type ParseError struct {
	Field string
	Value string
	Err   error
}

func (e *ParseError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (value: %s)", e.Field, e.Err.Error(), e.Value)
	}
	return fmt.Sprintf("%s: invalid value %q", e.Field, e.Value)
}

func (e *ParseError) Unwrap() error {
	return e.Err
}

func ParseOptions(spec string, cause error) (Options, error) {
	// If cause contains a ParseError, return it directly
	if cause != nil {
		if _, ok := cause.(*ParseError); ok {
			// Parse the spec normally, but return the original ParseError
			opts, _ := parseSpec(spec)
			return opts, cause
		}
		// Check if cause wraps a ParseError
		if err := errors.Unwrap(cause); err != nil {
			if _, ok := err.(*ParseError); ok {
				opts, _ := parseSpec(spec)
				return opts, cause
			}
		}
	}

	// Parse the spec
	opts, err := parseSpec(spec)
	if err != nil {
		return opts, err
	}

	return opts, nil
}

func parseSpec(spec string) (Options, error) {
	var opts Options
	opts.Limit = new(int)
	*opts.Limit = 100 // default limit
	opts.Labels = make(map[string]string)

	if spec == "" {
		return opts, nil
	}

	// Split by commas and trim spaces
	parts := strings.Split(spec, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Split by first equals sign
		pair := strings.SplitN(part, "=", 2)
		if len(pair) != 2 {
			// Invalid syntax
			return opts, &ParseError{Field: "syntax", Value: part, Err: errors.New("invalid syntax")}
		}

		key := strings.TrimSpace(pair[0])
		value := strings.TrimSpace(pair[1])

		switch key {
		case "limit":
			if value == "" {
				return opts, &ParseError{Field: "limit", Value: value, Err: errors.New("empty limit")}
			}
			limit, err := strconv.Atoi(value)
			if err != nil {
				return opts, &ParseError{Field: "limit", Value: value, Err: err}
			}
			if limit < 1 || limit > 1000 {
				return opts, &ParseError{Field: "limit", Value: value, Err: errors.New("limit must be between 1 and 1000")}
			}
			opts.Limit = &limit
		case "label":
			// label.NAME
			if value == "" {
				return opts, &ParseError{Field: "label", Value: value, Err: errors.New("empty label value")}
			}
			// Check if key is label.NAME
			if !strings.HasPrefix(key, "label.") {
				return opts, &ParseError{Field: "label", Value: key, Err: errors.New("invalid label key")}
			}
			name := key[6:] // remove "label."
			if name == "" {
				return opts, &ParseError{Field: "label", Value: key, Err: errors.New("label name cannot be empty")}
			}
			opts.Labels[name] = value
		default:
			// Unknown field
			return opts, &ParseError{Field: key, Value: value, Err: errors.New("unknown field")}
		}
	}

	return opts, nil
}
