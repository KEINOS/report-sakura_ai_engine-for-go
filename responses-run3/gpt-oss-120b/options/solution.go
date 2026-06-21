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
	Value string // the raw value (empty if not applicable)
	Err   error  // underlying error, may be nil
}

// Error implements the error interface.
func (e *ParseError) Error() string {
	if e == nil {
		return "<nil>"
	}
	if e.Err != nil {
		return fmt.Sprintf("parse error on field %q with value %q: %v", e.Field, e.Value, e.Err)
	}
	return fmt.Sprintf("parse error on field %q with value %q", e.Field, e.Value)
}

// Unwrap returns the wrapped error.
func (e *ParseError) Unwrap() error { return e.Err }

// ParseOptions parses a comma‑separated specification string.
// It returns the resulting Options and either nil or a *ParseError.
// If the supplied cause error (or any error it wraps) is a *ParseError,
// the function still parses the spec but returns that original *ParseError.
func ParseOptions(spec string, cause error) (Options, error) {
	// initialise defaults
	defLimit := 100
	opts := Options{
		Limit:  &defLimit,
		Labels: make(map[string]string),
	}

	var parseErr error

	// Trim surrounding whitespace; an empty spec is valid.
	if strings.TrimSpace(spec) != "" {
		fields := strings.Split(spec, ",")
		for _, f := range fields {
			f = strings.TrimSpace(f)
			if f == "" {
				// ignore empty fragments such as ", ,"
				continue
			}

			key, val, ok := strings.Cut(f, "=")
			if !ok {
				parseErr = &ParseError{
					Field: f,
					Value: "",
					Err:   fmt.Errorf("missing \"=\""),
				}
				break
			}
			key = strings.TrimSpace(key)
			val = strings.TrimSpace(val)

			if key == "" {
				parseErr = &ParseError{
					Field: key,
					Value: val,
					Err:   fmt.Errorf("empty key"),
				}
				break
			}

			switch {
			case key == "limit":
				n, err := strconv.Atoi(val)
				if err != nil {
					parseErr = &ParseError{Field: key, Value: val, Err: err}
					break
				}
				if n < 1 || n > 1000 {
					parseErr = &ParseError{
						Field: key,
						Value: val,
						Err:   fmt.Errorf("limit out of range (1‑1000)"),
					}
					break
				}
				// store a fresh copy so the pointer remains stable
				v := n
				opts.Limit = &v

			case strings.HasPrefix(key, "label."):
				name := key[len("label."):]
				if name == "" {
					parseErr = &ParseError{
						Field: key,
						Value: val,
						Err:   fmt.Errorf("empty label name"),
					}
					break
				}
				opts.Labels[name] = val

			default:
				parseErr = &ParseError{
					Field: key,
					Value: val,
					Err:   fmt.Errorf("unknown field"),
				}
			}

			if parseErr != nil {
				break
			}
		}
	}

	// If the supplied cause (or any wrapped error) is a *ParseError,
	// return that error instead of any parsing error we may have produced.
	var causePE *ParseError
	if errors.As(cause, &causePE) {
		return opts, causePE
	}
	return opts, parseErr
}
