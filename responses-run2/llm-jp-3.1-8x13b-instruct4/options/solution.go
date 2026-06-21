package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// Options contains the parsed options from the spec string.
type Options struct {
	Limit  *int
	Labels map[string]string
}

// ParseError is an error that implements error and Unwrap.
type ParseError struct {
	Field   string
	Value   string
	Err     error
	Unwrap  error
	Unwrapf string
}

func (e ParseError) Error() string {
	return fmt.Sprintf("ParseError{Field: %s, Value: %s, Err: %v, Unwrap: %v, Unwrapf: %s}", e.Field, e.Value, e.Err, e.Unwrap, e.Unwrapf)
}

func (e ParseError) Unwrap() error {
	return e.Err
}

func (e ParseError) Unwrapf() string {
	return e.Unwrapf
}

// ParseOptions parses the spec string and returns the Options.
func ParseOptions(spec string, cause error) (Options, error) {
	var (
		limit int
		labels map[string]string
		err    error
	)

	// Split spec into fields
	fields := strings.Split(spec, ",")

	// Parse limit
	if len(fields) > 0 {
		field := fields[0]
		if strings.HasPrefix(field, "limit=") {
			value, err := strconv.Atoi(field[6:])
			if err != nil {
				return Options{}, ParseError{Field: "limit", Value: field, Err: err}
			}
			if value < 1 || value > 1000 {
				return Options{}, ParseError{Field: "limit", Value: field, Err: errors.New("limit must be between 1 and 1000")}
			}
			limit = &value
		}
	}

	// Parse labels
	if len(fields) > 0 {
		field := fields[0]
		if strings.HasPrefix(field, "label=") {
			parts := strings.Split(field[6:], "=")
			if len(parts) != 2 {
				return Options{}, ParseError{Field: "label", Value: field, Err: errors.New("invalid label format")}
			}
			labels[parts[0]] = parts[1]
		}
	}

	// Check for cause error
	if cause != nil {
		if pe, ok := cause.(*ParseError); ok {
			return Options{Limit: limit, Labels: labels}, pe
		}
		return Options{Limit: limit, Labels: labels}, cause
	}

	return Options{Limit: limit, Labels: labels}, nil
}

func main() {
	spec := "limit=100,label=key1=value1,label=key2=value2"
	cause := errors.New("some cause error")

	options, err := ParseOptions(spec, cause)
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("Options:", options)
	}
}
