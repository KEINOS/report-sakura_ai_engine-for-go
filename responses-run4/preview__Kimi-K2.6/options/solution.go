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
	return fmt.Sprintf("options: parse error: field=%q value=%q: %v", e.Field, e.Value, e.Err)
}

func (e *ParseError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

func ParseOptions(spec string, cause error) (Options, error) {
	var opts Options
	opts.Limit = new(int)
	*opts.Limit = 100
	opts.Labels = make(map[string]string)

	var causePE *ParseError
	hasCausePE := errors.As(cause, &causePE)

	if spec == "" {
		if hasCausePE {
			return opts, causePE
		}
		return opts, nil
	}

	fail := func(field, value string, err error) (Options, error) {
		if hasCausePE {
			return opts, causePE
		}
		return opts, &ParseError{Field: field, Value: value, Err: err}
	}

	for _, f := range strings.Split(spec, ",") {
		key, val, ok := strings.Cut(f, "=")
		if !ok {
			return fail(strings.TrimSpace(f), "", errors.New("missing '='"))
		}

		key = strings.TrimSpace(key)
		val = strings.TrimSpace(val)

		if key == "" {
			return fail(key, val, errors.New("empty key"))
		}

		switch {
		case key == "limit":
			if val == "" {
				return fail(key, val, errors.New("empty limit"))
			}
			n, err := strconv.Atoi(val)
			if err != nil || n < 1 || n > 1000 {
				return fail(key, val, errors.New("invalid limit"))
			}
			opts.Limit = new(int)
			*opts.Limit = n

		case strings.HasPrefix(key, "label."):
			name := strings.TrimPrefix(key, "label.")
			if name == "" {
				return fail(key, val, errors.New("empty label name"))
			}
			opts.Labels[name] = val

		default:
			return fail(key, val, errors.New("unknown field"))
		}
	}

	if hasCausePE {
		return opts, causePE
	}
	return opts, nil
}
