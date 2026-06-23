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
	if e.Value == "" {
		return "options: invalid field " + strconv.Quote(e.Field) + ": " + e.Err.Error()
	}
	return "options: invalid field " + strconv.Quote(e.Field) + " value " + strconv.Quote(e.Value) + ": " + e.Err.Error()
}

func (e *ParseError) Unwrap() error {
	return e.Err
}

func ParseOptions(spec string, cause error) (Options, error) {
	var existing *ParseError
	_ = errors.As(cause, &existing)

	defaultLimit := 100
	opts := Options{
		Limit:  &defaultLimit,
		Labels: make(map[string]string),
	}

	fail := func(pe *ParseError) (Options, error) {
		if existing != nil {
			return opts, existing
		}
		return opts, pe
	}

	if spec == "" {
		if existing != nil {
			return opts, existing
		}
		return opts, nil
	}

	for _, field := range strings.Split(spec, ",") {
		field = strings.TrimSpace(field)
		if field == "" {
			return fail(&ParseError{Field: "", Value: "", Err: errors.New("empty field")})
		}

		key, val, ok := strings.Cut(field, "=")
		if !ok {
			return fail(&ParseError{Field: field, Value: "", Err: errors.New("missing '='")})
		}

		key = strings.TrimSpace(key)
		val = strings.TrimSpace(val)
		if key == "" {
			return fail(&ParseError{Field: key, Value: val, Err: errors.New("empty key")})
		}

		if key == "limit" {
			n, err := strconv.Atoi(val)
			if err != nil {
				return fail(&ParseError{Field: key, Value: val, Err: errors.New("invalid integer")})
			}
			if n < 1 || n > 1000 {
				return fail(&ParseError{Field: key, Value: val, Err: errors.New("limit out of range")})
			}
			limit := n
			opts.Limit = &limit
		} else if name, ok := strings.CutPrefix(key, "label."); ok {
			if name == "" {
				return fail(&ParseError{Field: key, Value: val, Err: errors.New("empty label name")})
			}
			opts.Labels[name] = val
		} else {
			return fail(&ParseError{Field: key, Value: val, Err: errors.New("unknown field")})
		}
	}

	if existing != nil {
		return opts, existing
	}
	return opts, nil
}
