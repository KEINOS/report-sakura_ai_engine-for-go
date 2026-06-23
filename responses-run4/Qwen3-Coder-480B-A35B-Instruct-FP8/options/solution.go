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

	var firstErr error

	for _, field := range strings.Split(spec, ",") {
		field = strings.TrimSpace(field)
		if field == "" {
			continue
		}

		parts := strings.SplitN(field, "=", 2)
		if len(parts) != 2 {
			if firstErr == nil {
				firstErr = &ParseError{
					Field: "field",
					Value: field,
					Err:   fmt.Errorf("missing '=' separator"),
				}
			}
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch {
		case key == "limit":
			if limit, err := strconv.Atoi(value); err != nil || limit < 1 || limit > 1000 {
				if firstErr == nil {
					firstErr = &ParseError{
						Field: "limit",
						Value: value,
						Err:   fmt.Errorf("must be an integer between 1 and 1000"),
					}
				}
			} else {
				opts.Limit = &limit
			}
		case strings.HasPrefix(key, "label."):
			labelName := key[6:]
			if labelName == "" {
				if firstErr == nil {
					firstErr = &ParseError{
						Field: "label name",
						Value: key,
						Err:   fmt.Errorf("label name cannot be empty"),
					}
				}
				continue
			}
			opts.Labels[labelName] = value
		default:
			if firstErr == nil {
				firstErr = &ParseError{
					Field: key,
					Value: value,
					Err:   fmt.Errorf("unknown field"),
				}
			}
		}
	}

	if opts.Limit == nil {
		opts.Limit = intPtr(100)
	}

	return *opts, firstErr
}

func intPtr(i int) *int {
	return &i
}

// As is a simplified version of errors.As for this specific use case
func As(err error, target interface{}) bool {
	if target == nil {
		panic("errors: target cannot be nil")
	}
	
	for err != nil {
		if perr, ok := err.(*ParseError); ok {
			*target.(**ParseError) = perr
			return true
		}
		if werr, ok := err.(interface{ Unwrap() error }); ok {
			err = werr.Unwrap()
		} else {
			break
		}
	}
	return false
}
