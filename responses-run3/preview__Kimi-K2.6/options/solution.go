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
		return fmt.Sprintf("options: field %q value %q: %v", e.Field, e.Value, e.Err)
	}
	return fmt.Sprintf("options: field %q value %q", e.Field, e.Value)
}

func (e *ParseError) Unwrap() error {
	return e.Err
}

// ParseOptions parses a comma-separated key=value spec into Options.
// If cause contains a *ParseError anywhere in its chain, the spec is parsed
// normally and the original *ParseError is returned.
func ParseOptions(spec string, cause error) (Options, error) {
	var prev *ParseError
	if errors.As(cause, &prev) {
		opts, _ := parse(spec)
		return opts, prev
	}
	return parse(spec)
}

func parse(spec string) (Options, error) {
	defaultLimit := 100
	opts := Options{
		Limit:  &defaultLimit,
		Labels: make(map[string]string),
	}

	if spec == "" {
		return opts, nil
	}

	for _, field := range strings.Split(spec, ",") {
		field = strings.TrimSpace(field)
		if field == "" {
			return opts, &ParseError{Field: "", Value: "", Err: errors.New("empty field")}
		}

		key, value, ok := strings.Cut(field, "=")
		if !ok {
			return opts, &ParseError{Field: field, Value: "", Err: errors.New("missing '='")}
		}

		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)

		if key == "" {
			return opts, &ParseError{Field: key, Value: value, Err: errors.New("empty key")}
		}

		switch {
		case key == "limit":
			v, err := strconv.Atoi(value)
			if err != nil {
				return opts, &ParseError{Field: "limit", Value: value, Err: fmt.Errorf("invalid integer: %w", err)}
			}
			if v < 1 || v > 1000 {
				return opts, &ParseError{Field: "limit", Value: value, Err: errors.New("out of range")}
			}
			opts.Limit = &v
		case strings.HasPrefix(key, "label."):
			name := strings.TrimPrefix(key, "label.")
			if name == "" {
				return opts, &ParseError{Field: key, Value: value, Err: errors.New("empty label name")}
			}
			opts.Labels[name] = value
		default:
			return opts, &ParseError{Field: key, Value: value, Err: errors.New("unknown field")}
		}
	}

	return opts, nil
}
