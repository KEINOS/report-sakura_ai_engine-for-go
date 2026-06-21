package options

import (
	"errors"
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
		return e.Err.Error()
	}
	return e.Field
}

func (e *ParseError) Unwrap() error {
	return e.Err
}

func ParseOptions(spec string, cause error) (Options, error) {
	var pe *ParseError
	if cause != nil {
		errors.As(cause, &pe)
	}

	opts := Options{
		Labels: make(map[string]string),
	}
	limit := 100
	opts.Limit = &limit

	if spec == "" {
		if pe != nil {
			return opts, pe
		}
		return opts, nil
	}

	for _, part := range strings.Split(spec, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			err := &ParseError{Field: part, Value: "", Err: errors.New("missing '='")}
			if pe != nil {
				return opts, pe
			}
			return opts, err
		}

		key := strings.TrimSpace(kv[0])
		val := strings.TrimSpace(kv[1])

		switch {
		case key == "limit":
			n, err := strconv.Atoi(val)
			if err != nil || n < 1 || n > 1000 {
				parseErr := &ParseError{Field: "limit", Value: val, Err: err}
				if pe != nil {
					return opts, pe
				}
				return opts, parseErr
			}
			opts.Limit = &n
		case strings.HasPrefix(key, "label."):
			name := key[6:]
			if name == "" {
				parseErr := &ParseError{Field: key, Value: val, Err: errors.New("label name cannot be empty")}
				if pe != nil {
					return opts, pe
				}
				return opts, parseErr
			}
			opts.Labels[name] = val
		default:
			parseErr := &ParseError{Field: key, Value: val, Err: errors.New("unknown field")}
			if pe != nil {
				return opts, pe
			}
			return opts, parseErr
		}
	}

	if pe != nil {
		return opts, pe
	}
	return opts, nil
}
