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

// ParseError describes a failure while parsing a single field.
type ParseError struct {
	Field string // the key that caused the error
	Value string // the raw value string (may be empty)
	Err   error  // underlying error
}

// Error implements the error interface.
func (e *ParseError) Error() string {
	return fmt.Sprintf("parse error on %s=%s: %v", e.Field, e.Value, e.Err)
}

// Unwrap returns the wrapped error.
func (e *ParseError) Unwrap() error { return e.Err }

// ParseOptions parses a comma‑separated specification string into an Options
// value.  If cause (or any error it wraps) is a *ParseError, the function still
// parses spec but returns the original *ParseError as the error result.
func ParseOptions(spec string, cause error) (Options, error) {
	// default values
	defLimit := 100
	opts := Options{
		Limit:  &defLimit,
		Labels: make(map[string]string),
	}

	// If the supplied cause already contains a *ParseError, parse spec
	// normally but return that original error.
	var causePE *ParseError
	if errors.As(cause, &causePE) {
		parsed, _ := parseSpec(spec, opts) // ignore any new errors
		return parsed, causePE
	}

	parsed, err := parseSpec(spec, opts)
	if err != nil {
		return parsed, err
	}
	return parsed, nil
}

// parseSpec does the actual work of parsing spec into opts.
// It returns the partially filled Options together with a *ParseError on the
// first encountered problem, or nil if parsing succeeded.
func parseSpec(spec string, base Options) (Options, error) {
	opts := base

	// Empty spec is valid.
	if strings.TrimSpace(spec) == "" {
		return opts, nil
	}

	fields := strings.Split(spec, ",")
	for _, f := range fields {
		f = strings.TrimSpace(f)
		if f == "" {
			// Skip empty entries (e.g., trailing commas).
			continue
		}

		kv := strings.SplitN(f, "=", 2)
		if len(kv) != 2 {
			return opts, &ParseError{
				Field: f,
				Value: "",
				Err:   fmt.Errorf("missing \"=\""),
			}
		}
		key := strings.TrimSpace(kv[0])
		val := strings.TrimSpace(kv[1])

		if key == "" {
			return opts, &ParseError{
				Field: key,
				Value: val,
				Err:   fmt.Errorf("empty key"),
			}
		}

		switch {
		case key == "limit":
			n, err := strconv.Atoi(val)
			if err != nil {
				return opts, &ParseError{Field: key, Value: val, Err: err}
			}
			if n < 1 || n > 1000 {
				return opts, &ParseError{
					Field: key,
					Value: val,
					Err:   fmt.Errorf("value %d out of range (1‑1000)", n),
				}
			}
			// Set a new pointer to avoid mutating the default.
			limit := n
			opts.Limit = &limit

		case strings.HasPrefix(key, "label."):
			name := strings.TrimPrefix(key, "label.")
			if name == "" {
				return opts, &ParseError{
					Field: key,
					Value: val,
					Err:   fmt.Errorf("label name is empty"),
				}
			}
			if opts.Labels == nil {
				opts.Labels = make(map[string]string)
			}
			// Last occurrence wins.
			opts.Labels[name] = val

		default:
			return opts, &ParseError{
				Field: key,
				Value: val,
				Err:   fmt.Errorf("unknown field"),
			}
		}
	}

	// Ensure invariants.
	if opts.Labels == nil {
		opts.Labels = make(map[string]string)
	}
	if opts.Limit == nil {
		def := 100
		opts.Limit = &def
	}
	return opts, nil
}
