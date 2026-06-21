package options

import (
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
	return fmt.Sprintf("invalid %s=%q: %v", e.Field, e.Value, e.Err)
}

func (e *ParseError) Unwrap() error {
	return e.Err
}

func ParseOptions(spec string, cause error) (Options, error) {
	// Check if cause is already a ParseError
	var parseErr *ParseError
	if errors.As(cause, &parseErr) {
		// If so, we'll return it as-is after parsing
	}

	opts := Options{
		Limit:  new(int), // Default limit is 100
		Labels: make(map[string]string),
	}
	*opts.Limit = 100

	if spec == "" {
		return opts, parseErr
	}

	parts := strings.Split(spec, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			err := &ParseError{
				Field: "field",
				Value: part,
				Err:   fmt.Errorf("missing '='"),
			}
			return opts, err
		}

		key := strings.TrimSpace(kv[0])
		value := strings.TrimSpace(kv[1])

		switch {
		case key == "limit":
			limit, err := strconv.Atoi(value)
			if err != nil {
				err = &ParseError{
					Field: "limit",
					Value: value,
					Err:   fmt.Errorf("not an integer"),
				}
				return opts, err
			}
			if limit < 1 || limit > 1000 {
				err = &ParseError{
					Field: "limit",
					Value: value,
					Err:   fmt.Errorf("must be between 1 and 1000"),
				}
				return opts, err
			}
			opts.Limit = &limit
		case strings.HasPrefix(key, "label."):
			name := strings.TrimPrefix(key, "label.")
			if name == "" {
				err := &ParseError{
					Field: "label",
					Value: key,
					Err:   fmt.Errorf("empty name"),
				}
				return opts, err
			}
			opts.Labels[name] = value
		default:
			err := &ParseError{
				Field: "field",
				Value: key,
				Err:   fmt.Errorf("unknown field"),
			}
			return opts, err
		}
	}

	if parseErr != nil {
		return opts, parseErr
	}
	return opts, nil
}
