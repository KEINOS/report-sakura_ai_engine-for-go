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
	return fmt.Sprintf("invalid %s value %q: %v", e.Field, e.Value, e.Err)
}

func (e *ParseError) Unwrap() error {
	return e.Err
}

func ParseOptions(spec string, cause error) (Options, error) {
	opts := Options{
		Labels: make(map[string]string),
	}

	// Check if cause contains a ParseError
	var parseErr *ParseError
	if cause != nil && As(cause, &parseErr) {
		// Parse normally but return the existing ParseError
		_, _ = parseOptions(spec, &opts)
		return opts, parseErr
	}

	return parseOptions(spec, &opts)
}

func parseOptions(spec string, opts *Options) (Options, error) {
	if spec == "" {
		opts.Limit = intPtr(100)
		return *opts, nil
	}

	for _, field := range strings.Split(spec, ",") {
		field = strings.TrimSpace(field)
		if field == "" {
			continue
		}

		parts := strings.SplitN(field, "=", 2)
		if len(parts) != 2 {
			return *opts, &ParseError{
				Field: "field",
				Value: field,
				Err:   fmt.Errorf("missing '='"),
			}
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch {
		case key == "limit":
			limit, err := strconv.Atoi(value)
			if err != nil || limit < 1 || limit > 1000 {
				return *opts, &ParseError{
					Field: "limit",
					Value: value,
					Err:   fmt.Errorf("must be an integer from 1 through 1000"),
				}
			}
			opts.Limit = &limit
		case strings.HasPrefix(key, "label."):
			labelName := key[6:]
			if labelName == "" {
				return *opts, &ParseError{
					Field: "label name",
					Value: "",
					Err:   fmt.Errorf("must be non-empty"),
				}
			}
			opts.Labels[labelName] = value
		default:
			return *opts, &ParseError{
				Field: key,
				Value: value,
				Err:   fmt.Errorf("unknown field"),
			}
		}
	}

	if opts.Limit == nil {
		opts.Limit = intPtr(100)
	}

	return *opts, nil
}

func intPtr(i int) *int {
	return &i
}

// As is a helper to mimic errors.As for our specific case
func As(err error, target interface{}) bool {
	if err == nil {
		return false
	}
	if perr, ok := err.(*ParseError); ok {
		if target, ok := target.(**ParseError); ok {
			*target = perr
			return true
		}
	}
	return As(fmt.Errorf("%w", err), target)
}
