package main

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
	return fmt.Sprintf("error parsing field %s: %v", e.Field, e.Err)
}

func (e *ParseError) Unwrap() error {
	return e.Err
}

func ParseOptions(spec string, cause error) (Options, error) {
	if cause != nil && cause.Error() != nil {
		return Options{}, cause
	}

	options := Options{
		Labels: make(map[string]string),
	}

	fields := strings.Split(spec, ",")
	for _, field := range fields {
		field = strings.TrimSpace(field)
		if field == "" {
			continue
		}

		keyValue := strings.SplitN(field, "=", 2)
		if len(keyValue) != 2 {
			return options, &ParseError{Field: field, Err: errors.New("invalid field syntax")}
		}

		key := strings.TrimSpace(keyValue[0])
		value := strings.TrimSpace(keyValue[1])

		switch key {
		case "limit":
			if limit, err := strconv.Atoi(value); err != nil || limit < 1 || limit > 1000 {
				return options, &ParseError{Field: key, Value: value, Err: errors.New("invalid limit value")}
			}
			options.Limit = &limit
		case "label":
			if value == "" {
				return options, &ParseError{Field: key, Value: value, Err: errors.New("label value cannot be empty")}
			}
			options.Labels[value] = value
		default:
			return options, &ParseError{Field: key, Value: value, Err: fmt.Errorf("unknown field: %s", key)}
		}
	}

	if options.Limit == nil {
		options.Limit = &100
	}

	return options, nil
}

func main() {
	spec := "limit=500,label1=foo,label2=bar"
	options, err := ParseOptions(spec, nil)
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Printf("Options: %+v\n", options)
	}
}
