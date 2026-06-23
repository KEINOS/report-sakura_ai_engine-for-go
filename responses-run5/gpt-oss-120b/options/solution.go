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
	Value string // the raw value that caused the error
	Err   error  // underlying error
}

// Error implements the error interface.
func (p *ParseError) Error() string {
	return fmt.Sprintf("invalid %s=%s: %v", p.Field, p.Value, p.Err)
}

// Unwrap returns the wrapped error.
func (p *ParseError) Unwrap() error { return p.Err }

// ParseOptions parses a comma‑separated specification string.
// If cause (or any error it wraps) is a *ParseError, the function parses the
// spec as far as possible and returns the same *ParseError from cause.
// Otherwise it returns the first parsing error it encounters.
func ParseOptions(spec string, cause error) (Options, error) {
	// default options
	defLimit := 100
	opts := Options{
		Limit:  &defLimit,
		Labels: make(map[string]string),
	}

	// If cause already contains a *ParseError, remember it.
	var causePE *ParseError
	if errors.As(cause, &causePE) {
		// Parse spec normally but ignore any new errors – we will return causePE.
		_ = parseSpec(spec, &opts)
		return opts, causePE
	}

	// Normal parsing – stop at first error.
	if err := parseSpec(spec, &opts); err != nil {
		return opts, err
	}
	return opts, nil
}

// parseSpec parses spec into opts and returns the first *ParseError encountered, if any.
func parseSpec(spec string, opts *Options) error {
	if strings.TrimSpace(spec) == "" {
		return nil
	}
	fields := strings.Split(spec, ",")
	for _, f := range fields {
		f = strings.TrimSpace(f)
		if f == "" {
			continue
		}
		kvIdx := strings.Index(f, "=")
		if kvIdx < 0 {
			return &ParseError{
				Field: f,
				Value: "",
				Err:   errors.New("missing '='"),
			}
		}
		key := strings.TrimSpace(f[:kvIdx])
		val := strings.TrimSpace(f[kvIdx+1:])

		switch {
		case key == "limit":
			n, err := strconv.Atoi(val)
			if err != nil {
				return &ParseError{Field: key, Value: val, Err: err}
			}
			if n < 1 || n > 1000 {
				return &ParseError{
					Field: key,
					Value: val,
					Err:   fmt.Errorf("value %d out of range (1‑1000)", n),
				}
			}
			// set new limit
			limit := n
			opts.Limit = &limit

		case strings.HasPrefix(key, "label."):
			name := strings.TrimPrefix(key, "label.")
			if name == "" {
				return &ParseError{
					Field: key,
					Value: val,
					Err:   errors.New("label name is empty"),
				}
			}
			if opts.Labels == nil {
				opts.Labels = make(map[string]string)
			}
			opts.Labels[name] = val // last value wins

		default:
			return &ParseError{
				Field: key,
				Value: val,
				Err:   fmt.Errorf("unknown field %q", key),
			}
		}
	}
	return nil
}
