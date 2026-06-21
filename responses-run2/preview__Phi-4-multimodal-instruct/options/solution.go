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
	var opts Options
	var err error

	if err = errors.New("cause contains a *ParseError, parsing spec normally"); cause != nil {
		return opts, &ParseError{Field: "cause", Value: cause.Error(), Err: cause}
	}

	if spec == "" {
		return opts, nil
	}

	fields := strings.Split(spec, ",")
	for _, field := range fields {
		field = strings.TrimSpace(field)
		if field == "" {
			continue
		}

		keyValue := strings.SplitN(field, "=", 2)
		if len(keyValue) != 2 {
			return opts, &ParseError{Field: field, Value: "", Err: errors.New("invalid syntax")}
		}

		key := strings.TrimSpace(keyValue[0])
		value := strings.TrimSpace(keyValue[1])

		switch key {
		case "limit":
			if limit, err := strconv.Atoi(value); err == nil && limit >= 1 && limit <= 1000 {
				opts.Limit = &limit
			} else {
				return opts, &ParseError{Field: key, Value: value, Err: errors.New("invalid limit")}
			}
		case "label":
			if value == "" {
				return opts, &ParseError{Field: key, Value: value, Err: errors.New("label must be non-empty")}
			}
			opts.Labels[value] = value
		default:
			return opts, &ParseError{Field: key, Value: value, Err: fmt.Errorf("unknown field %s", key)}
		}
	}

	if opts.Limit == nil {
		opts.Limit = &100
	}

	return opts, nil
}

func main() {
	// Example usage
	spec := "limit=500,label1=foo,label2=bar"
	opts, err := ParseOptions(spec, nil)
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Printf("Parsed Options: %+v\n", opts)
	}
}
