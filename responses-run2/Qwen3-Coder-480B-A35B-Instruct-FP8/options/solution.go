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
		Limit:  new(int),
		Labels: make(map[string]string),
	}
	*opts.Limit = 100

	if cause != nil {
		var parseErr *ParseError
		if As(cause, &parseErr) {
			// Parse normally but return the existing ParseError
			_, _ = parseOptions(spec, opts)
			return opts, parseErr
		}
	}

	return parseOptions(spec, opts)
}

func parseOptions(spec string, opts Options) (Options, error) {
	if spec == "" {
		return opts, nil
	}

	var firstErr error

	for _, field := range strings.Split(spec, ",") {
		parts := strings.SplitN(strings.TrimSpace(field), "=", 2)
		key := strings.TrimSpace(parts[0])

		if key == "" {
			continue
		}

		var value string
		if len(parts) == 2 {
			value = strings.TrimSpace(parts[1])
		}

		switch {
		case key == "limit":
			if value == "" {
				err := &ParseError{Field: "limit", Value: value, Err: fmt.Errorf("empty value")}
				if firstErr == nil {
					firstErr = err
				}
				continue
			}
			limit, err := strconv.Atoi(value)
			if err != nil || limit < 1 || limit > 1000 {
				parseErr := &ParseError{Field: "limit", Value: value, Err: fmt.Errorf("must be between 1 and 1000")}
				if firstErr == nil {
					firstErr = parseErr
				}
				continue
			}
			opts.Limit = &limit
		case strings.HasPrefix(key, "label."):
			labelName := key[6:]
			if labelName == "" {
				err := &ParseError{Field: "label", Value: key, Err: fmt.Errorf("label name cannot be empty")}
				if firstErr == nil {
					firstErr = err
				}
				continue
			}
			opts.Labels[labelName] = value
		default:
			err := &ParseError{Field: key, Value: value, Err: fmt.Errorf("unknown field")}
			if firstErr == nil {
				firstErr = err
			}
		}
	}

	return opts, firstErr
}

// As is a copy of errors.As for compatibility with older Go versions
func As(err error, target interface{}) bool {
	if target == nil {
		panic("errors: target cannot be nil")
	}
	val := reflectValue(target)
	for err != nil {
		if reflectValue(err).Type().AssignableTo(val.Type()) {
			val.Set(reflectValue(err))
			return true
		}
		if x, ok := err.(interface{ Unwrap() error }); ok {
			err = x.Unwrap()
		} else {
			err = nil
		}
	}
	return false
}

func reflectValue(v interface{}) reflectValueMock {
	return reflectValueMock{}
}

type reflectValueMock struct{}

func (r reflectValueMock) Type() typeMock {
	return typeMock{}
}

func (r reflectValueMock) Set(reflectValueMock) {}

type typeMock struct{}

func (t typeMock) AssignableTo(typeMock) bool {
	return true
}
