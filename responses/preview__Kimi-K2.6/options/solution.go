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
	if e == nil {
		return "<nil>"
	}
	return fmt.Sprintf("options: invalid field %q value %q: %v", e.Field, e.Value, e.Err)
}

func (e *ParseError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

func ParseOptions(spec string, cause error) (Options, error) {
	var existing *ParseError
	if errors.As(cause, &existing) && existing != nil {
		opts, _ := parse(spec)
		return opts, existing
	}
	opts, perr := parse(spec)
	if perr != nil {
		return opts, perr
	}
	return opts, nil
}

func parse(spec string) (Options, *ParseError) {
	const defaultLimit = 100
	limit := defaultLimit
	opts := Options{
		Limit:  &limit,
		Labels: make(map[string]string),
	}

	if spec == "" {
		return opts, nil
	}

	fields := strings.Split(spec, ",")
	for _, f := range fields {
		key, value, ok := strings.Cut(f, "=")
		if !ok {
			return opts, &ParseError{
				Field: strings.TrimSpace(f),
				Value: "",
				Err:   errors.New("missing '='"),
			}
		}

		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)

		if key == "" {
			return opts, &ParseError{
				Field: key,
				Value: value,
				Err:   errors.New("empty key"),
			}
		}

		switch {
		case key == "limit":
			n, err := strconv.Atoi(value)
			if err != nil {
				return opts, &ParseError{Field: "limit", Value: value, Err: err}
			}
			if n < 1 || n > 1000 {
				return opts, &ParseError{
					Field: "limit",
					Value: value,
					Err:   errors.New("limit must be between 1 and 1000"),
				}
			}
			*opts.Limit = n
		case strings.HasPrefix(key, "label."):
			name := key[len("label."):]
			if name == "" {
				return opts, &ParseError{
					Field: key,
					Value: value,
					Err:   errors.New("label name must be non-empty"),
				}
			}
			opts.Labels[name] = value
		default:
			return opts, &ParseError{
				Field: key,
				Value: value,
				Err:   errors.New("unknown field"),
			}
		}
	}

	return opts, nil
}
