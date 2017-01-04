package config

import (
	"errors"
	"fmt"
)

var (
	// ErrParse is a generic parsing error
	ErrParse = errors.New("Could not parse config")
)

//FieldError is a custom error for missing fields in the config
type FieldError struct {
	Name string
}

func (e FieldError) Error() string {
	return "Config is missing required field: " + e.Name
}

// KeyError is a custom error for a key without a registered parser
type KeyError struct {
	Name string
}

func (e KeyError) Error() string {
	return "Could not parse config. Unrecognized key: " + e.Name
}

// ValueError is a custom error for any error parsing/validating a value
type ValueError struct {
	Name string
	// The underlying error
	Err error
}

func (e ValueError) Error() string {
	return fmt.Sprintf("Error parsing value %s: %s", e.Name, e.Err)
}

// NewValueError wraps a regular error into a ValueError, or returns nil if the
// error is nil
func NewValueError(name string, err error) error {
	if err == nil {
		return nil
	}
	return ValueError{Name: name, Err: err}
}
